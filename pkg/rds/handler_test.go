package rds

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/coldog/rds-operator/pkg/apis/rds/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockRDS struct {
	rdsiface.RDSAPI
	mock.Mock
}

func (m *mockRDS) DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*rds.DescribeDBInstancesOutput), args.Error(1)
}

func (m *mockRDS) CreateDBInstance(input *rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*rds.CreateDBInstanceOutput), args.Error(1)
}

type mockSDK struct {
	mock.Mock
	obj sdk.Object
}

func (m *mockSDK) Create(object sdk.Object) error {
	m.obj = object
	return m.Called(object).Error(0)
}

func (m *mockSDK) Update(object sdk.Object) error {
	m.obj = object
	return m.Called(object).Error(0)
}

func handler() (*mockRDS, *mockSDK, *Handler) {
	sdk := &mockSDK{}
	rds := &mockRDS{}
	h := &Handler{sdk: sdk, rds: rds}
	return rds, sdk, h
}

func TestHandler_Run(t *testing.T) {
	r, s, h := handler()

	s.On("Update", mock.Anything).Return(nil)
	s.On("Create", mock.Anything).Return(nil)
	r.On("DescribeDBInstances", mock.Anything).Return(
		&rds.DescribeDBInstancesOutput{},
		errors.New("exists"),
	)
	r.On("CreateDBInstance", mock.Anything).Return(&rds.CreateDBInstanceOutput{
		DBInstance: &rds.DBInstance{
			Endpoint: &rds.Endpoint{
				Address: aws.String("test"),
				Port:    aws.Int64(10),
			},
		},
	}, nil)

	h.Handle(context.Background(), sdk.Event{
		Object: &v1alpha1.Database{},
	})

	s.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestHandler_AlreadySet(t *testing.T) {
	r, s, h := handler()

	h.Handle(context.Background(), sdk.Event{
		Object: &v1alpha1.Database{
			Status: v1alpha1.DatabaseStatus{State: v1alpha1.StateCreated},
		},
	})

	s.AssertNotCalled(t, "Create")
	s.AssertNotCalled(t, "Update")
	r.AssertNotCalled(t, "DescribeDBInstances")

	s.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestHandler_StatusFail(t *testing.T) {
	r, s, h := handler()

	s.On("Update", mock.Anything).Return(errors.New("failure"))

	h.Handle(context.Background(), sdk.Event{
		Object: &v1alpha1.Database{},
	})

	r.AssertNotCalled(t, "DescribeDBInstances")

	s.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestHandler_AlreadyExists(t *testing.T) {
	r, s, h := handler()

	s.On("Update", mock.Anything).Return(nil)
	r.On("DescribeDBInstances", mock.Anything).Return(
		&rds.DescribeDBInstancesOutput{},
		nil,
	)

	h.Handle(context.Background(), sdk.Event{
		Object: &v1alpha1.Database{},
	})

	require.Equal(t, v1alpha1.StateCreated, s.obj.(*v1alpha1.Database).Status.State)

	s.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestHandler_CreationFailure(t *testing.T) {
	r, s, h := handler()

	s.On("Update", mock.Anything).Return(nil)
	r.On("DescribeDBInstances", mock.Anything).Return(
		&rds.DescribeDBInstancesOutput{},
		errors.New("exists"),
	)
	r.On("CreateDBInstance", mock.Anything).Return(&rds.CreateDBInstanceOutput{
		DBInstance: &rds.DBInstance{
			Endpoint: &rds.Endpoint{
				Address: aws.String("test"),
				Port:    aws.Int64(10),
			},
		},
	}, errors.New("test-error"))

	h.Handle(context.Background(), sdk.Event{
		Object: &v1alpha1.Database{},
	})

	require.Equal(t, v1alpha1.StateFailure, s.obj.(*v1alpha1.Database).Status.State)
	require.Equal(t, "test-error", s.obj.(*v1alpha1.Database).Status.Error)

	s.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestHandler_CreateSecretFailure(t *testing.T) {
	r, s, h := handler()

	s.On("Update", mock.Anything).Return(nil)
	s.On("Create", mock.Anything).Return(errors.New("test-error"))

	r.On("DescribeDBInstances", mock.Anything).Return(
		&rds.DescribeDBInstancesOutput{},
		errors.New("exists"),
	)
	r.On("CreateDBInstance", mock.Anything).Return(&rds.CreateDBInstanceOutput{
		DBInstance: &rds.DBInstance{
			Endpoint: &rds.Endpoint{
				Address: aws.String("test"),
				Port:    aws.Int64(10),
			},
		},
	}, nil)

	h.Handle(context.Background(), sdk.Event{
		Object: &v1alpha1.Database{},
	})

	require.Equal(t, v1alpha1.StateFailure, s.obj.(*v1alpha1.Database).Status.State)
	require.Equal(t, "test-error", s.obj.(*v1alpha1.Database).Status.Error)

	s.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestHandler_CreateSecretExisting(t *testing.T) {
	r, s, h := handler()

	s.On("Update", mock.Anything).Return(nil)
	s.On("Create", mock.Anything).Return(
		k8errors.NewAlreadyExists(schema.GroupResource{}, ""),
	)

	r.On("DescribeDBInstances", mock.Anything).Return(
		&rds.DescribeDBInstancesOutput{},
		errors.New("exists"),
	)
	r.On("CreateDBInstance", mock.Anything).Return(&rds.CreateDBInstanceOutput{
		DBInstance: &rds.DBInstance{
			Endpoint: &rds.Endpoint{
				Address: aws.String("test"),
				Port:    aws.Int64(10),
			},
		},
	}, nil)

	h.Handle(context.Background(), sdk.Event{
		Object: &v1alpha1.Database{},
	})

	require.Equal(t, v1alpha1.StateCreated, s.obj.(*v1alpha1.Database).Status.State)

	s.AssertExpectations(t)
	r.AssertExpectations(t)
}
