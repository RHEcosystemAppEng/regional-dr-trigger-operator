// Copyright (c) 2023 Red Hat, Inc.

package actions

import (
	"context"
	addonv1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// This file contains the action for deleting the ResilientCluster from the OLD spoke.

// deleteOldResilientCluster is used for deleting Hive's ClusterDeployment from the OLD spoke.
func deleteOldResilientCluster(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)
	logger.Info("deleting old resilient cluster", "old-spoke", options.OldSpoke)

	// the ResilientCluster resides in the cluster-namespace with a matching name
	oldRCSubject := types.NamespacedName{
		Namespace: options.OldSpoke,
		Name:      options.OldSpoke,
	}

	// fetch ResilientCluster from previous OLD cluster and delete it if exists
	oldRC := &addonv1.ResilientCluster{}
	if err := options.Client.Get(ctx, oldRCSubject, oldRC); err != nil {
		logger.Info("no ResilientCluster found", "old-spoke", options.OldSpoke)
	} else {
		if err = options.Client.Delete(ctx, oldRC); err != nil {
			logger.Error(err, "failed deleting ResilientCluster", "old-spoke", options.OldSpoke)
		}
	}
}

// init is registering deleteOldResilientCluster for running.
func init() {
	actionFuncs = append(actionFuncs, deleteOldResilientCluster)
}
