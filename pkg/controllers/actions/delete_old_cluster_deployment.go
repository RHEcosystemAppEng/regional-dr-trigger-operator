// Copyright (c) 2023 Red Hat, Inc.

package actions

import (
	"context"
	"fmt"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func deleteOldClusterDeployment(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)

	// the ClusterDeployment resides in the cluster-namespace with a matching name
	oldDeploymentSubject := types.NamespacedName{
		Namespace: options.OldSpoke,
		Name:      options.OldSpoke,
	}

	// fetch ClusterDeployment from previous OLD cluster and delete it if exists
	oldDeployment := &hivev1.ClusterDeployment{}
	if err := options.Client.Get(ctx, oldDeploymentSubject, oldDeployment); err != nil {
		logger.Info(fmt.Sprintf("no ClusterDeployments found on %s", options.OldSpoke))
	} else {
		if err = options.Client.Delete(ctx, oldDeployment); err != nil {
			logger.Error(err, fmt.Sprintf("failed deleting ClusterDepolyment %v", oldDeploymentSubject))
		}
	}
}

func init() {
	actionFuncs = append(actionFuncs, deleteOldClusterDeployment)
}
