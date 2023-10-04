// Copyright (c) 2023 Red Hat, Inc.

package manager

// This file hosts functions and types for setting our Addon Manager registration process for Agent Addons.

import (
	"context"
	"fmt"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

// getRegistrationOptionFunc is used to create a function for creating a registry option for configuring the addon automated
// registration process. In runtime, when the addon is enabled on a Spoke, this will create a role and a binding on the
// Spoke's cluster-namespace on the Hub. This will allow the Spoke's Agent to access resource in its own namespace on
// the Hub.
// The kubeConfig if for the Hub's kubeconfig.
// Use the 'getPermissionConfig' function to modify the required resources.
func getRegistrationOptionFunc(ctx context.Context, kubeConfig *rest.Config) *agent.RegistrationOption {
	agentName := rand.String(5)
	return &agent.RegistrationOption{
		CSRConfigurations: agent.KubeClientSignerConfigurations(AddonName, agentName),
		CSRApproveCheck:   utils.DefaultCSRApprover(agentName), // this is the default auto-approving, consider replacing
		PermissionConfig:  getPermissionConfig(ctx, kubeConfig),
	}
}

// getPermissionConfig is used for creating a function that will create a role and a role binding in the
// cluster-namespace in the Hub cluster these resources will be used by the ManagedCluster to handle resources in its
// target cluster-namespace in the Hub modify the verbs if needed in role.yaml.
func getPermissionConfig(ctx context.Context, kubeConfig *rest.Config) agent.PermissionConfigFunc {
	return func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) error {
		kubeClientSet, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}

		templateValues := struct {
			Group        string
			ResourceName string
		}{
			Group:        agent.DefaultGroups(cluster.Name, addon.Name)[0],
			ResourceName: fmt.Sprintf("open-cluster-management:%s:agent", addon.Name),
		}

		// load role from template, the role specify what access will the Spoke have in its cluster-namespace on the Hub
		var role rbacv1.Role
		if err = loadTemplateFromFile("templates/rbac/role.yaml", templateValues, &role); err != nil {
			return err
		}

		// create the role if not found
		_, err = kubeClientSet.RbacV1().Roles(cluster.Name).Get(ctx, role.Name, metav1.GetOptions{})
		switch {
		case errors.IsNotFound(err):
			_, createErr := kubeClientSet.RbacV1().Roles(cluster.Name).Create(ctx, &role, metav1.CreateOptions{})
			if createErr != nil {
				return createErr
			}
		case err != nil:
			return err
		}

		// load rolebinding from template, the rolebinding binds the aforementioned role to the addon group
		var binding rbacv1.RoleBinding
		if err = loadTemplateFromFile("templates/rbac/rolebinding.yaml", templateValues, &binding); err != nil {
			return err
		}

		// create the rolebinding if not found
		_, err = kubeClientSet.RbacV1().RoleBindings(cluster.Name).Get(ctx, binding.Name, metav1.GetOptions{})
		switch {
		case errors.IsNotFound(err):
			_, createErr := kubeClientSet.RbacV1().RoleBindings(cluster.Name).Create(ctx, &binding, metav1.CreateOptions{})
			if createErr != nil {
				return createErr
			}
		case err != nil:
			return err
		}

		return nil
	}
}
