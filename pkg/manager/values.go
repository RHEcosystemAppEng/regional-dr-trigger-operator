// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"fmt"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

type Values struct {
	KubeConfigSecret string
	SpokeName        string
	AgentNamespace   string
	AgentReplicas    int
}

// getTemplateValuesFunc is used to build a function for generating values to be used in the Addon Agent templates.
func getTemplateValuesFunc(options *Options) func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	return func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
		values := Values{
			KubeConfigSecret: fmt.Sprintf("%s-hub-kubeconfig", addon.Name),
			SpokeName:        cluster.Name,
			AgentNamespace:   addon.Spec.InstallNamespace,
			AgentReplicas:    options.AgentReplicas,
		}
		return addonfactory.StructToValues(values), nil
	}
}
