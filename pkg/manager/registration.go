// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		CSRApproveCheck:   utils.DefaultCSRApprover(agentName),
		PermissionConfig:  getPermissionConfig(ctx, kubeConfig),
	}
}

// getPermissionConfig is used for creating a function that will create a role and a role binding in the
// cluster-namespace in the Hub cluster these resources will be used by the ManagedCluster to handle resources in its
// target cluster-namespace in the Hub modify the verbs if needed.
func getPermissionConfig(ctx context.Context, kubeConfig *rest.Config) agent.PermissionConfigFunc {
	return func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) error {
		kubeclient, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}

		groups := agent.DefaultGroups(cluster.Name, addon.Name)

		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("open-cluster-management:%s:agent", addon.Name),
				Namespace: cluster.Name,
			},
			Rules: []rbacv1.PolicyRule{
				{Verbs: []string{"get", "list", "watch"}, Resources: []string{"configmaps"}, APIGroups: []string{""}},
				{Verbs: []string{"get", "list", "watch"}, Resources: []string{"managedclusteraddons"}, APIGroups: []string{"addon.open-cluster-management.io"}},
			},
		}

		binding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("open-cluster-management:%s:agent", addon.Name),
				Namespace: cluster.Name,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     fmt.Sprintf("open-cluster-management:%s:agent", addon.Name),
			},
			Subjects: []rbacv1.Subject{
				{Kind: "Group", APIGroup: "rbac.authorization.k8s.io", Name: groups[0]},
			},
		}

		// create the role if not found
		_, err = kubeclient.RbacV1().Roles(cluster.Name).Get(ctx, role.Name, metav1.GetOptions{})
		switch {
		case errors.IsNotFound(err):
			_, createErr := kubeclient.RbacV1().Roles(cluster.Name).Create(ctx, role, metav1.CreateOptions{})
			if createErr != nil {
				return createErr
			}

		case err != nil:
			return err
		}

		// create the rolebinding if not found
		_, err = kubeclient.RbacV1().RoleBindings(cluster.Name).Get(ctx, binding.Name, metav1.GetOptions{})
		switch {
		case errors.IsNotFound(err):
			_, createErr := kubeclient.RbacV1().RoleBindings(cluster.Name).Create(ctx, binding, metav1.CreateOptions{})
			if createErr != nil {
				return createErr
			}
		case err != nil:
			return err
		}

		return nil
	}
}
