// Copyright (c) 2023 Red Hat, Inc.

package manager

// This file hosts functions and types for configuring our Addon Manager with instructions for creating Agent Addons.

import (
	"context"
	"fmt"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/mcra"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	addonv1alpha1client "open-cluster-management.io/api/client/addon/clientset/versioned"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"strconv"
)

// agentValues is used for encapsulating template values for the Agent templates.
type agentValues struct {
	KubeConfigSecret string
	SpokeName        string
	AgentNamespace   string
	AgentImage       string
}

// deploymentValues i used for encapsulating template values extracted from the AddonDeploymentConfig.
type deploymentValues struct {
	AgentReplicas int
}

// createAgent is used for creating the Addon Agent configuration for the Addon Manager.
func createAgent(ctx context.Context, kubeConfig *rest.Config, options *Options) (agent.AgentAddon, error) {
	client, err := addonv1alpha1client.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	getter := utils.NewAddOnDeploymentConfigGetter(client)

	return addonfactory.
		NewAgentAddonFactory(mcra.AddonName, fsys, "templates/agent").
		WithConfigGVRs(utils.AddOnDeploymentConfigGVR).
		WithGetValuesFuncs(
			addonfactory.GetAddOnDeploymentConfigValues(getter, loadDeploymentValuesFunc),
			getTemplateValuesFunc(options)).
		WithAgentRegistrationOption(getRegistrationOptionFunc(ctx, kubeConfig)).
		BuildTemplateAgentAddon()
}

// getTemplateValuesFunc is used for building a function for generating values to be used in the Addon Agent templates.
func getTemplateValuesFunc(options *Options) func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	return func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
		values := agentValues{
			KubeConfigSecret: fmt.Sprintf("%s-hub-kubeconfig", addon.Name),
			SpokeName:        cluster.Name,
			AgentNamespace:   addon.Spec.InstallNamespace,
			AgentImage:       options.AgentImage,
		}

		return addonfactory.StructToValues(values), nil
	}
}

// loadDeploymentValuesFunc is a function for instructing template values from an injected AddOnDeploymentConfig.
func loadDeploymentValuesFunc(config addonv1alpha1.AddOnDeploymentConfig) (addonfactory.Values, error) {
	values := deploymentValues{}
	for _, variable := range config.Spec.CustomizedVariables {
		if variable.Name == "AgentReplicas" {
			replicas, err := strconv.Atoi(variable.Value)
			if err != nil {
				return nil, err
			}

			values.AgentReplicas = replicas
		}
	}
	return addonfactory.StructToValues(values), nil
}
