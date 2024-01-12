// Copyright (c) 2023 Red Hat, Inc.

package manager

// This file hosts functions and types for configuring the Addon Manager with instructions for creating Agent Addons.

import (
	"context"
	"fmt"
	mcra "github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg"
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

// deploymentValues is used for encapsulating template values extracted from the AddonDeploymentConfig.
type deploymentValues struct {
	AgentReplicas  int
	AgentNamespace string
}

// createAgent is used for creating the Addon Agent configuration for the Addon Manager.
func createAgent(ctx context.Context, kubeConfig *rest.Config, options *Options) (agent.AgentAddon, error) {
	client, err := addonv1alpha1client.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	getter := utils.NewAddOnDeploymentConfigGetter(client)

	agentAddon := addonfactory.
		NewAgentAddonFactory(mcra.AddonName, fsys, "templates/agent").
		WithConfigGVRs(utils.AddOnDeploymentConfigGVR).
		WithGetValuesFuncs(
			// keep following functions order to allow AddOnDeploymentConfig's AgentInstallNamespace to override
			// ManagedClusterAddOn's InstallNamespace
			getTemplateValuesFunc(options),
			addonfactory.GetAddOnDeploymentConfigValues(getter, loadDeploymentValuesFunc)).
		WithAgentRegistrationOption(getRegistrationOptionFunc(ctx, kubeConfig))

	// configurable support for installing the agent's ManagedClusterAddon automatically on all cluster namespaces
	if options.InstallAllStrategy {
		agentAddon.WithInstallStrategy(agent.InstallByFilterFunctionStrategy(options.InstallAllNamespace, targetMcPredicate))
	}

	return agentAddon.BuildTemplateAgentAddon()
}

// getTemplateValuesFunc is used for building a function for generating values to be used in the Addon Agent templates.
func getTemplateValuesFunc(options *Options) func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	return func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
		values := agentValues{
			KubeConfigSecret: fmt.Sprintf("%s-hub-kubeconfig", addon.Name),
			SpokeName:        cluster.Name,
			// namespace from ManagedClusterAddon defaults to 'open-cluster-management-agent-addon'
			// this function should be called before loadDeploymentValuesFunc to allow this to be overridden
			// TODO: UNCOMMENT THIS ONCE ACM SUPPORTS open-cluster-management.io/api v0.12.0
			//AgentNamespace: addon.Spec.InstallNamespace,
			AgentImage: options.AgentImage,
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
	// namespace from AddOnDeploymentConfig is set to its default open-cluster-management-agent-addon, we don't want it
	// to override the one set in ManagedClusterAddOn
	// TODO: UNCOMMENT THIS ONCE ACM SUPPORTS open-cluster-management.io/api v0.12.0
	//if config.Spec.AgentInstallNamespace != "open-cluster-management-agent-addon" {
	if true {
		// this function should be called after getTemplateValuesFunc for this to override the one set in ManagedClusterAddOn
		values.AgentNamespace = config.Spec.AgentInstallNamespace
	}
	return addonfactory.StructToValues(values), nil
}

// targetMcPredicate is used for filtering Managed Clusters as target for the Addon Agent. Currently, it returns false
// if the cluster name is 'local_cluster', this is done to prevent the Addon Agent to from being installed for the
// Standalone Cluster on the Hub.
func targetMcPredicate(cluster *clusterv1.ManagedCluster) bool {
	return cluster.Name != "local-cluster"
}
