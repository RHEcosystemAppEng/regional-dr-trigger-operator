// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"context"
	"embed"
	"k8s.io/client-go/rest"
	klogv2 "k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
)

//go:embed agenttemplates
var FS embed.FS

const AddonName = "multicluster-resiliency-addon"

type Manager struct {
	Options *Options
}

type Options struct {
	AgentReplicas int
}

// NewManager is used as a factory for creating a Manager instance
func NewManager() Manager {
	return Manager{Options: &Options{}}
}

// Run is used for creating the Addon and running the Addon Manager
func (m *Manager) Run(ctx context.Context, kubeConfig *rest.Config) error {
	klogv2.Info("running manager")

	addonMgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		return err
	}

	agentAddon, err := addonfactory.
		NewAgentAddonFactory(AddonName, FS, "agenttemplates").
		WithGetValuesFuncs(getTemplateValuesFunc(m.Options)).
		WithAgentRegistrationOption(getRegistrationOptionFunc(ctx, kubeConfig)).
		BuildTemplateAgentAddon()
	if err != nil {
		return err
	}

	if err = addonMgr.AddAgent(agentAddon); err != nil {
		return err
	}

	go func() {
		if err = addonMgr.Start(ctx); err != nil {
			klogv2.Fatalf("failed to add start addon: %v", err)
		}
	}()

	<-ctx.Done()

	return nil
}
