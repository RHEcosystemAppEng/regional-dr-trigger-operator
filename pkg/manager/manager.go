// Copyright (c) 2023 Red Hat, Inc.

package manager

// This file hosts functions and types for running our Addon Manager on the Hub cluster.

import (
	"context"
	"embed"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/controllers"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed templates/agent templates/rbac
var fsys embed.FS

// Manager is a receiver representing the Addon manager. It encapsulates the Manager Options which will be used for
// configuring the manager run. Use NewManager for instantiation.
type Manager struct {
	Options *Options
}

// Options is used for encapsulating the various options for configuring the manager Run.
type Options struct {
	ControllerMetricAddr     string
	ControllerProbeAddr      string
	ControllerLeaderElection bool
	AgentImage               string
	ServiceAccount           string
	ConfigMapName            string
	EnableValidation         bool
}

// NewManager is used as a factory for creating a Manager instance with an Options instance.
func NewManager() Manager {
	return Manager{Options: &Options{}}
}

// Run is used for running the Addon manager. It takes a context and the kubeconfig for the Hub it runs on. This
// function blocks while running the controller (the Agent manager runs as a goroutine).
func (m *Manager) Run(ctx context.Context, kubeConfig *rest.Config) error {
	logger := log.FromContext(ctx)

	addonMgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		logger.Error(err, "failed creating the addon manager")
		return err
	}

	agentAddon, err := createAgent(ctx, kubeConfig, m.Options)
	if err != nil {
		logger.Error(err, "failed creating the addon agent")
		return err
	}

	if err = addonMgr.AddAgent(agentAddon); err != nil {
		logger.Error(err, "failed adding the addon agent to the addon manager")
		return err
	}

	go func() {
		if err = addonMgr.Start(ctx); err != nil {
			logger.Error(err, "failed to start the addon manager")
		}
	}()

	ctrl := controllers.NewControllersWithOptions(&controllers.Options{
		MetricAddr:       m.Options.ControllerMetricAddr,
		ProbeAddr:        m.Options.ControllerProbeAddr,
		LeaderElection:   m.Options.ControllerLeaderElection,
		ServiceAccount:   m.Options.ServiceAccount,
		ConfigMapName:    m.Options.ConfigMapName,
		EnableValidation: m.Options.EnableValidation,
	})

	// blocking
	return ctrl.Run(ctx, kubeConfig)
}
