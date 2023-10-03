// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"context"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
)

// Manager is a receiver representing the Addon Manager.
// It encapsulates the Agent Options which will be used to configure the Agent run.
// Use NewManager for instantiation.
type Manager struct {
	Options *Options
}

// Options is used for encapsulating the various options for configuring the Manager Run.
type Options struct {
	AgentReplicas int
}

// NewManager is used as a factory for creating a Manager instance.
func NewManager() Manager {
	return Manager{Options: &Options{}}
}

// Run is used for running the Addon Manager.
// It takes a context and the kubeconfig for the Hub it runs on.
func (m *Manager) Run(ctx context.Context, kubeConfig *rest.Config) error {
	klog.Info("running addon manager")

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
			klog.Fatalf("failed to start the addon manager: %v", err)
		}
	}()

	<-ctx.Done()

	klog.Info("addon manager done")
	return nil
}
