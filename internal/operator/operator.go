// Copyright (c) 2023 Red Hat, Inc.

package operator

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"regional-dr-trigger-operator/internal/controller"
	"regional-dr-trigger-operator/internal/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// DRTriggerOperator is the receiver for running the operator and binding the options
type DRTriggerOperator struct {
	Options *DRTriggerOperatorOptions
}

// DRTriggerOperatorOptions is used for encapsulating the operator options
type DRTriggerOperatorOptions struct {
	MetricAddr     string
	LeaderElection bool
	ProbeAddr      string
	Debug          bool
	MetricsSecure  bool
}

// NewDRTriggerOperator is a factory function for creating a regional dr trigger operator instance
func NewDRTriggerOperator() DRTriggerOperator {
	return DRTriggerOperator{Options: &DRTriggerOperatorOptions{}}
}

// Run is used for running the DRTriggerOperator. It takes a cobra.Command reference and string array of arguments.
func (c *DRTriggerOperator) Run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	// set logging and create initial logger
	ctrl.SetLogger(zap.New(zap.UseDevMode(c.Options.Debug)))
	logger := log.Log.WithName("rdrtrigger-operator")

	// create the scheme and install the required types
	scheme := runtime.NewScheme()
	if err := utils.InstallTypes(scheme); err != nil {
		logger.Error(err, "failed installing scheme")
		return err
	}

	// configure metrics
	metricsOpts := server.Options{BindAddress: c.Options.MetricAddr, SecureServing: c.Options.MetricsSecure}
	if c.Options.MetricsSecure {
		metricsOpts.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// create the manager
	kubeConfig := config.GetConfigOrDie()
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme:                 scheme,
		Logger:                 logger,
		LeaderElection:         c.Options.LeaderElection,
		LeaderElectionID:       "regional-dr-trigger-operator-leader-election-id",
		Metrics:                metricsOpts,
		HealthProbeBindAddress: c.Options.ProbeAddr,
	})
	if err != nil {
		logger.Error(err, "failed creating k8s manager")
		return err
	}

	// set up the controller
	controller := &controller.DRTriggerController{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
	if err = controller.SetupWithManager(mgr); err != nil {
		logger.Error(err, "failed setting up the controller")
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

	return mgr.Start(ctx)
}
