// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file contains predicates related functions for use as filters for controller events.

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// verifyManagedCluster takes a function and returns a predicate that verifies the event object for all events against
// the function.
// NOTE: BLOCKING ALL DELETE EVENTS!
func verifyManagedCluster(fn func(obj client.Object) bool) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return fn(createEvent.Object)
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			//return fn(deleteEvent.Object)
			return false
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return fn(updateEvent.ObjectOld)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return fn(genericEvent.Object)
		},
	}
}

// hasAddon is a utility function that takes an addon name and returns a function that takes a client.Object and
// returns true if we have the required ManagedClusterAddon in a namespace of the Object (expected to be ManagedCluster).
func hasAddon(ctx context.Context, clnt client.Client, addon string) func(obj client.Object) bool {
	return func(obj client.Object) bool {
		mcaSubject := types.NamespacedName{Name: addon, Namespace: obj.GetName()}
		mca := &addonv1alpha1.ManagedClusterAddOn{}
		// return whether the managedclusteraddon exists on the managedcluster cluster-namespace
		if err := clnt.Get(ctx, mcaSubject, mca); err != nil {
			if !errors.IsNotFound(err) {
				logger := log.FromContext(ctx)
				logger.Error(err, fmt.Sprintf("failed fetching %s ManagedClusterAddon for %s ManagedCluster", addon, obj.GetName()))
			}
			// the addon is NOT installed for the managed cluster
			return false
		}
		// the addon is installed for the managed cluster
		return true
	}
}

func acceptedByHub() func(obj client.Object) bool {
	return func(obj client.Object) bool {
		return obj.(*clusterv1.ManagedCluster).Spec.HubAcceptsClient
	}
}

func joinedHub() func(obj client.Object) bool {
	return func(obj client.Object) bool {
		mc := obj.(*clusterv1.ManagedCluster)
		return meta.IsStatusConditionTrue(mc.Status.Conditions, clusterv1.ManagedClusterConditionJoined)
	}
}

func notAvailable() func(obj client.Object) bool {
	return func(obj client.Object) bool {
		mc := obj.(*clusterv1.ManagedCluster)
		return meta.IsStatusConditionFalse(mc.Status.Conditions, clusterv1.ManagedClusterConditionAvailable)
	}
}
