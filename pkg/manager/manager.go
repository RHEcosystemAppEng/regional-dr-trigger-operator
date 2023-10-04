// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"context"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/controller"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Manager is a receiver representing the Addon manager.
// It encapsulates the Manager Options which will be used to configure the manager run.
// Use NewManager for instantiation.
type Manager struct {
	Options *Options
}

// Options is used for encapsulating the various options for configuring the manager Run.
type Options struct {
	ControllerMetricAddr     string
	ControllerProbeAddr      string
	ControllerLeaderElection bool
	AgentReplicas            int
	AgentImage               string
}

// NewManager is used as a factory for creating a Manager instance.
func NewManager() Manager {
	return Manager{Options: &Options{}}
}

// Run is used for running the Addon Manager.
// It takes a context and the kubeconfig for the Hub it runs on.
func (m *Manager) Run(ctx context.Context, kubeConfig *rest.Config) error {
	logger := log.FromContext(ctx)

	addonMgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		return err
	}

	agentAddon, err := createAgent(ctx, kubeConfig, m.Options)
	if err != nil {
		return err
	}

	if err = addonMgr.AddAgent(agentAddon); err != nil {
		return err
	}

	go func() {
		if err = addonMgr.Start(ctx); err != nil {
			logger.Error(err, "failed to start the addon manager")
		}
	}()

	ctrl := controller.NewControllerWithOptions(&controller.Options{
		MetricAddr:     m.Options.ControllerMetricAddr,
		ProbeAddr:      m.Options.ControllerProbeAddr,
		LeaderElection: m.Options.ControllerLeaderElection,
	})

	// blocking
	if err = ctrl.Run(ctx, kubeConfig); err != nil {
		return err
	}

	return nil
}
