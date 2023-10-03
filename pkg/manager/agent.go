// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"context"
	"embed"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/agent"
)

//go:embed agenttemplates
var FS embed.FS // Resource templates used for deploying the Addon Agent to Spokes.

// createAgent is used for creating the Addon Agent configuration for the Addon Manager.
func createAgent(ctx context.Context, kubeConfig *rest.Config, options *Options) (agent.AgentAddon, error) {
	return addonfactory.
		NewAgentAddonFactory(AddonName, FS, "agenttemplates").
		WithGetValuesFuncs(getTemplateValuesFunc(options)).
		WithAgentRegistrationOption(getRegistrationOptionFunc(ctx, kubeConfig)).
		BuildTemplateAgentAddon()
}
