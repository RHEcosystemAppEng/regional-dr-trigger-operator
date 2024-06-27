// Copyright (c) 2023 Red Hat, Inc.

package utils

import (
	"fmt"
	ramenv1alpha1 "github.com/ramendr/ramen/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

// InstallTypes is used for installing all the required types with a scheme.
func InstallTypes(scheme *runtime.Scheme) error {
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
