package main

import (
	"context"
	"runtime"
	"time"

	"github.com/coldog/rds-operator/pkg/rds"
	"github.com/coldog/rds-operator/version"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func printVersion() {
	log.SetLevel(log.DebugLevel)
	log.WithFields(log.Fields{
		"goVersion":  runtime.Version(),
		"goOs":       runtime.GOOS,
		"goArch":     runtime.GOARCH,
		"sdkVersion": sdkVersion.Version,
		"rdsVersion": version.Version,
	}).Info("starting")
}

func main() {
	printVersion()

	sdk.ExposeMetricsPort()

	handler, err := rds.NewHandler()
	if err != nil {
		log.WithError(err).Fatal("failed init handler")
	}

	resource := "rds.aws.com/v1alpha1"
	kind := "Database"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.WithError(err).Fatal("failed watch namespace")
	}
	resyncPeriod := time.Duration(5) * time.Second

	log.WithFields(log.Fields{
		"resource":     resource,
		"kind":         kind,
		"namespace":    namespace,
		"resyncPeriod": resyncPeriod,
	}).Info("watching")

	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Handle(handler)
	sdk.Run(context.Background())
}
