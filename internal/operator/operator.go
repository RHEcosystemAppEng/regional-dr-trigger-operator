// Copyright (c) 2023 Red Hat, Inc.

package operator

import (
	"crypto/tls"
	"fmt"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"regional-dr-trigger-operator/internal/controller"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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
	EnableHttp2    bool
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
	logger := ctrl.Log.WithName("rdrtrigger-operator")

	// create the scheme and install the required types
	scheme := runtime.NewScheme()
	if err := installTypes(scheme); err != nil {
		logger.Error(err, "failed installing scheme")
		return err
	}

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		logger.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	var tlsOps []func(*tls.Config)
	if !c.Options.EnableHttp2 {
		tlsOps = append(tlsOps, disableHTTP2)
	}

	// configure metrics
	metricsOpts := server.Options{
		BindAddress: c.Options.MetricAddr, SecureServing: c.Options.MetricsSecure, TLSOpts: tlsOps}
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
		Cache: cache.Options{ByObject: map[client.Object]cache.ByObject{
			&ramenv1alpha1.DRPlacementControl{}: {Label: labels.Everything()},
		}},
	})
	if err != nil {
		logger.Error(err, "failed creating k8s manager")
		return err
	}

	// set up the controller
	controller := &controller.DRTriggerController{Client: mgr.GetClient(), Scheme: scheme}
	if err = controller.SetupWithManager(ctx, mgr); err != nil {
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

	logger.Info("stating manager")
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
