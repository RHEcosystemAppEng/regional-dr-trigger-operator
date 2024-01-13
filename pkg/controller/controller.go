// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts functions and types for instantiating the controller as part of the Addon Manager on the Hub cluster.

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Controller is a receiver representing the Addon controller. It encapsulates the Controller Options which will be used
// to configure the controller run. Use NewControllerWithOptions for instantiation.
type Controller struct {
	Options *Options
}

// Options is used for encapsulating the various options for configuring the controller run.
type Options struct {
	MetricAddr     string
	LeaderElection bool
	ProbeAddr      string
}

// NewControllerWithOptions is used as a factory for creating a Controller instance with a given Options instance.
func NewControllerWithOptions(options *Options) Controller {
	return Controller{Options: options}
}

// Run is used for running the Addon controller. It takes a context and the kubeconfig for the Hub it runs on. This
// function blocks while running the controller's manager.
func (c *Controller) Run(ctx context.Context, kubeConfig *rest.Config) error {
	logger := log.FromContext(ctx)

	// create and configure the scheme
	scheme := runtime.NewScheme()
	if err := installTypes(scheme); err != nil {
		logger.Error(err, "failed installing types")
		return err
	}

	// create a manager for the controller
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme:                 scheme,
		Logger:                 logger,
		LeaderElection:         c.Options.LeaderElection,
		LeaderElectionID:       "multicluster-resiliency-addon.appeng.ecosystem.redhat.com",
		Metrics:                server.Options{BindAddress: c.Options.MetricAddr},
		HealthProbeBindAddress: c.Options.ProbeAddr,
		BaseContext:            func() context.Context { return ctx },
	})
	if err != nil {
		logger.Error(err, "failed creating k8s manager")
		return err
	}

	reconciler := &AddonReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
	if err = reconciler.SetupWithManager(ctx, mgr); err != nil {
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
	// required for ManagedClusterAddon and AddonDeploymentConfig
	if err := addonv1alpha1.Install(scheme); err != nil {
		return fmt.Errorf("failed installing the framework types into the addon's scheme, %v", err)
	}
	// required for ManagedCluster
	if err := clusterv1.Install(scheme); err != nil {
		return fmt.Errorf("failed installing ocm's types into the addon's scheme, %v", err)
	}
	return nil
}
