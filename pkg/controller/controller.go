// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts functions and types for instantiating our controller as part of our Addon Manager on the Hub cluster.

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Controller is a receiver representing the Addon controller.
// It encapsulates the Controller Options which will be used to configure the controller run.
// Use NewControllerWithOptions for instantiation.
type Controller struct {
	Options *Options
}

// Options is used for encapsulating the various options for configuring the controller run.
type Options struct {
	MetricAddr     string
	LeaderElection bool
	ProbeAddr      string
}

// NewControllerWithOptions is used as a factory for creating a Controller instance.
func NewControllerWithOptions(options *Options) Controller {
	return Controller{Options: options}
}

// Run is used for running the Addon controller.
// It takes a context and the kubeconfig for the Hub it runs on.
// This function blocks while running the controller's manager.
func (c *Controller) Run(ctx context.Context, kubeConfig *rest.Config) error {
	logger := log.FromContext(ctx)

	scheme := runtime.NewScheme()
	if err := addonv1alpha1.Install(scheme); err != nil {
		return err
	}

	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme:                 scheme,
		Logger:                 logger,
		LeaderElection:         c.Options.LeaderElection,
		LeaderElectionID:       "todo-change-me",
		Metrics:                server.Options{BindAddress: c.Options.MetricAddr},
		HealthProbeBindAddress: c.Options.ProbeAddr,
		BaseContext:            func() context.Context { return ctx },
	})
	if err != nil {
		return err
	}

	reconciler := &McraReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
	if err = ctrl.NewControllerManagedBy(mgr).For(&addonv1alpha1.ManagedClusterAddOn{}).Complete(reconciler); err != nil {
		return err
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return err
	}

	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return err
	}

	// blocking
	if err = mgr.Start(ctx); err != nil {
		return err
	}

	return nil
}
