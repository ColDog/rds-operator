package rds

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/coldog/rds-operator/pkg/apis/rds/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SDK Represents the operator SDK.
type SDK interface {
	Create(object sdk.Object) error
	Update(object sdk.Object) error
}

type sdkWrap struct{}

func (sdkWrap) Create(object sdk.Object) error { return sdk.Create(object) }
func (sdkWrap) Update(object sdk.Object) error { return sdk.Update(object) }

// NewHandler returns a new handler instantiating and AWS client.
func NewHandler() (sdk.Handler, error) {
	awsSession, err := session.NewSession(&aws.Config{
		Region: str(os.Getenv("AWS_REGION")),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return &Handler{rds: rds.New(awsSession), sdk: sdkWrap{}}, nil
}

// Handler will create RDS databases.
type Handler struct {
	rds rdsiface.RDSAPI
	sdk SDK
}

func dbName(o *v1alpha1.Database) string { return o.Namespace + "-" + o.Name }

func secretName(o *v1alpha1.Database) string { return o.Name + "-db-credentials" }

// Handle will handle a specific event.
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Database:
		if event.Deleted {
			return h.delete(o)
		}

		if o.Status.State == v1alpha1.StateCreated ||
			o.Status.State == v1alpha1.StateFailure {
			return nil
		}

		if err := h.setStatus(o, v1alpha1.StatePending, nil); err != nil {
			return err
		}

		v1alpha1.Defaults(o)

		if err := h.create(o); err != nil {
			return h.setStatus(o, v1alpha1.StateFailure, err)
		}

		return h.setStatus(o, v1alpha1.StateCreated, nil)
	}
	return nil
}

func (h *Handler) setStatus(o *v1alpha1.Database, status string, err error) error {
	log.WithField("db", dbName(o)).WithField("state", status).Debug("set status")

	copy := o.DeepCopy()
	copy.Status.State = status
	if err != nil {
		copy.Status.Error = err.Error()
	}
	return h.sdk.Update(copy)
}

func (h *Handler) delete(cr *v1alpha1.Database) error {
	log.WithField("db", dbName(cr)).Debug("deleteing db")

	_, err := h.rds.DeleteDBInstance(&rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: str(dbName(cr)),
	})
	if err != nil {
		log.WithError(err).WithField("db", dbName(cr)).Error("deletion failed")
	}
	return err
}

func (h *Handler) create(o *v1alpha1.Database) error {
	err := h.getDB(o)
	if err == nil {
		log.WithField("db", dbName(o)).Info("db already exists")
		return nil
	}

	db, err := h.createDB(o)
	if err != nil {
		log.WithField("db", dbName(o)).WithError(err).Error("db creation failed")
		return err
	}

	err = h.sdk.Create(h.createSecret(o, db))
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		log.WithField("db", dbName(o)).WithError(err).Error("secret creation failed")
		return err
	}

	return nil
}

func (h *Handler) createSecret(cr *v1alpha1.Database, db *rds.DBInstance) *corev1.Secret {
	log.WithField("db", dbName(cr)).Debug("creating secret")

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			ClusterName: cr.ObjectMeta.ClusterName,
			Namespace:   cr.ObjectMeta.Namespace,
			Name:        secretName(cr),
			Labels:      cr.Labels,
			Annotations: map[string]string{"rds.aws.com/database": cr.Name},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "Database",
				}),
			},
		},
		Data: map[string][]byte{
			"username": encStr(cr.Spec.Username),
			"password": encStr(cr.Spec.Password),
			"host":     encStr(*db.Endpoint.Address),
			"port":     encI64(*db.Endpoint.Port),
			"url": encStr(
				cr.Spec.Engine + "://" + cr.Spec.Username + ":" +
					cr.Spec.Password + "@" + *db.Endpoint.Address + ":" +
					strI64(*db.Endpoint.Port) + "/" + cr.Spec.Database,
			),
		},
	}
}

func (h *Handler) getDB(cr *v1alpha1.Database) error {
	log.WithField("db", dbName(cr)).Debug("fetching db")

	_, err := h.rds.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: str(dbName(cr)),
	})
	return err
}

func (h *Handler) createDB(cr *v1alpha1.Database) (*rds.DBInstance, error) {
	spec := cr.Spec
	req := &rds.CreateDBInstanceInput{
		DBInstanceIdentifier:    str(dbName(cr)),
		MasterUsername:          str(spec.Username),
		MasterUserPassword:      str(spec.Password),
		DBName:                  str(spec.Database),
		Engine:                  str(spec.Engine),
		AllocatedStorage:        i64(spec.Storage),
		AutoMinorVersionUpgrade: bo(spec.AutoMinorVersionUpgrade),
		AvailabilityZone:        str(spec.AvailabilityZone),
		BackupRetentionPeriod:   i64(spec.BackupRetentionPeriod),
		CharacterSetName:        str(spec.CharacterSetName),
		DBInstanceClass:         str(spec.InstanceClass),
		DBSubnetGroupName:       str(spec.SubnetGroup),
		EngineVersion:           str(spec.EngineVersion),
		Iops:                    i64(spec.Iops),
		StorageType:             str(spec.StorageType),
		MultiAZ:                 bo(spec.MultiAZ),
		StorageEncrypted:        bo(spec.Encrypted),
		VpcSecurityGroupIds:     strs(spec.SecurityGroups),
	}

	log.WithField("db", dbName(cr)).
		WithField("instance", req).
		WithField("resource", cr).
		Debug("creating db")

	out, err := h.rds.CreateDBInstance(req)
	return out.DBInstance, err
}
