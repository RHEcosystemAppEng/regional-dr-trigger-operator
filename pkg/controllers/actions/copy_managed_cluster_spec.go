// Copyright (c) 2023 Red Hat, Inc.

package actions

// This file contains the action for copying a ManagedCluster Spec between spokes.

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// copyManagedClusterSpec is used for reconciling ManagedCluster specs. It will copy the spec from the ManagedCluster
// representing the OLD spoke into the MangedCluster representing the NEW one. It also deletes the OLD ManagedCluster.
func copyManagedClusterSpec(ctx context.Context, options Options) {
	logger := log.FromContext(ctx)

	// fetch the OLD ManagedCluster or break
	oldMc := &clusterv1.ManagedCluster{}
	if err := options.Client.Get(ctx, types.NamespacedName{Name: options.OldSpoke}, oldMc); err != nil {
		logger.Error(err, fmt.Sprintf("failed loading old ManagedCluster %s", options.OldSpoke))
		return
	}

	// fetch the NEW ManagedCluster or break
	newMc := &clusterv1.ManagedCluster{}
	if err := options.Client.Get(ctx, types.NamespacedName{Name: options.NewSpoke}, newMc); err != nil {
		logger.Error(err, fmt.Sprintf("failed loading new ManagedCluster %s", options.NewSpoke))
		return
	}

	// update the NEW ManagedCluster with the OLD spec or break
	newMc.Spec = oldMc.Spec
	if err := options.Client.Update(ctx, newMc); err != nil {
		logger.Error(err, fmt.Sprintf("failed updating new ManagedCluster %s", options.NewSpoke))
		return
	}

	// delete OLD ManagedCluster
	if err := options.Client.Delete(ctx, oldMc); err != nil {
		logger.Error(err, fmt.Sprintf("failed deleting old ManagedCluster %s", options.OldSpoke))
	}
}

// init is registering copyManagedClusterSpec for running.
func init() {
	actionFuncs = append(actionFuncs, copyManagedClusterSpec)
}
