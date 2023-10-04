// Copyright (c) 2023 Red Hat, Inc.

package manager

// This file hosts functions and types for configuring our Addon Manager with instructions for creating Agent Addons.

import (
	"context"
	"fmt"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/agent"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

// agentValues is used to encapsulate template values for the Agent templates.
type agentValues struct {
	KubeConfigSecret string
	SpokeName        string
	AgentNamespace   string
	AgentReplicas    int
	AgentImage       string
}

// createAgent is used for creating the Addon Agent configuration for the Addon Manager.
func createAgent(ctx context.Context, kubeConfig *rest.Config, options *Options) (agent.AgentAddon, error) {
	return addonfactory.
		NewAgentAddonFactory(AddonName, fsys, "templates/agent").
		WithGetValuesFuncs(getTemplateValuesFunc(options)).
		WithAgentRegistrationOption(getRegistrationOptionFunc(ctx, kubeConfig)).
		BuildTemplateAgentAddon()
}

// getTemplateValuesFunc is used to build a function for generating values to be used in the Addon Agent templates.
func getTemplateValuesFunc(options *Options) func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	return func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
		values := agentValues{
			KubeConfigSecret: fmt.Sprintf("%s-hub-kubeconfig", addon.Name),
			SpokeName:        cluster.Name,
			AgentNamespace:   addon.Spec.InstallNamespace,
			AgentReplicas:    options.AgentReplicas,
			AgentImage:       options.AgentImage,
		}
		return addonfactory.StructToValues(values), nil
	}
}
