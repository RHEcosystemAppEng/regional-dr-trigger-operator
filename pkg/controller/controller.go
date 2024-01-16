// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts functions and types for instantiating the DRTriggerController.

import (
	"context"
	"fmt"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// DRTriggerController is a receiver representing the controller. It encapsulates a DRTriggerControllerOptions which
// will be used to configure the controller run. Use NewDRTriggerController for instantiation.
type DRTriggerController struct {
	Options *DRTriggerControllerOptions
}

// DRTriggerControllerOptions is used for encapsulating the various options for configuring the controller run.
type DRTriggerControllerOptions struct {
	MetricAddr     string
	LeaderElection bool
	ProbeAddr      string
}

// NewDRTriggerController is used as a factory for creating a DRTriggerController instance.
func NewDRTriggerController() DRTriggerController {
	return DRTriggerController{Options: &DRTriggerControllerOptions{}}
}

// Run is used for running the DRTriggerController. It takes a context and a ControllerContext for the Hub it runs on.
func (c *DRTriggerController) Run(ctx context.Context, controllerContext *controllercmd.ControllerContext) error {
	logger := log.FromContext(ctx)

	// create and configure the scheme
	scheme := runtime.NewScheme()
	if err := installTypes(scheme); err != nil {
		logger.Error(err, "failed installing types")
		return err
	}

	// create a manager
	mgr, err := ctrl.NewManager(controllerContext.KubeConfig, ctrl.Options{
		Scheme:                 scheme,
		Logger:                 logger,
		LeaderElection:         c.Options.LeaderElection,
		LeaderElectionID:       "regional-dr-trigger-operator-leader-election-id",
		Metrics:                server.Options{BindAddress: c.Options.MetricAddr},
		HealthProbeBindAddress: c.Options.ProbeAddr,
		BaseContext:            func() context.Context { return ctx },
	})
	if err != nil {
		logger.Error(err, "failed creating k8s manager")
		return err
	}

	// set up the reconciler
	reconciler := &DRTriggerReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
	if err = reconciler.SetupWithManager(mgr); err != nil {
		logger.Error(err, "failed setting up the reconciler")
		return err
	}

	// configure health checks
	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "failed setting up health check")
		return err
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "failed setting up ready check")
		return err
	}

	// start the manager, blocking
	return mgr.Start(ctx)
}

// installTypes is used for installing all the required types with a scheme.
func installTypes(scheme *runtime.Scheme) error {
	// required for ManagedCluster
	if err := clusterv1.Install(scheme); err != nil {
		return fmt.Errorf("failed installing ocm's types into the scheme, %v", err)
	}
	// required for DRPlacementControl
	if err := ramenv1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed installing ramen's types into the scheme, %v", err)
	}
	return nil
}
