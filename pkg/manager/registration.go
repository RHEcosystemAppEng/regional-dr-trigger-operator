// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"text/template"

	v1 "k8s.io/api/rbac/v1"
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

const AddonName = "multicluster-resiliency-addon"

type templateValues struct {
	Group        string
	ResourceName string
}

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
// target cluster-namespace in the Hub modify the verbs if needed.
func getPermissionConfig(ctx context.Context, kubeConfig *rest.Config) agent.PermissionConfigFunc {
	return func(cluster *clusterv1.ManagedCluster, addon *addonv1alpha1.ManagedClusterAddOn) error {
		kubeClientSet, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}

		values := templateValues{
			Group:        agent.DefaultGroups(cluster.Name, addon.Name)[0],
			ResourceName: fmt.Sprintf("open-cluster-management:%s:agent", addon.Name),
		}

		// role with RBAC rules to access resources on hub
		var role v1.Role
		if err = executeFileTemplate("rbac/role.yaml", values, &role); err != nil {
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

		// rolebinding to bind the above role to a certain user group
		var binding v1.RoleBinding
		if err = executeFileTemplate("rbac/rolebinding.yaml", values, &binding); err != nil {
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

// executeFileTemplate is used to load a template file from fsys, execute ig against templateValues, and convert it
// into a structured generic object by reference target.
func executeFileTemplate[T any](file string, values templateValues, target *T) error {
	tmpl, err := template.ParseFS(fsys, file)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	if err = tmpl.Execute(&buff, values); err != nil {
		return err
	}

	manifest := make(map[string]interface{})
	if err = yaml.Unmarshal(buff.Bytes(), &manifest); err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(manifest, target); err != nil {
		return err
	}

	return nil
}
