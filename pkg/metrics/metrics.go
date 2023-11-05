// Copyright (c) 2023 Red Hat, Inc.

package metrics

// This file contains various metrics for use throughout the project.

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// vector labels for the metrics
const (
	LabelSpokeName    = "spoke_name"
	LabelClaimName    = "claim_name"
	LabelPoolName     = "pool_name"
	LabelOldSpokeName = "old_spoke_name"
	LabelNewSpokeName = "new_spoke_name"
)

var ResilientSpokeNotAvailable = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "resilient_spoke_not_available_count",
	Help: "Count times the Resilient Spoke cluster was reported not available",
}, []string{LabelSpokeName})

var ResilientSpokeAvailable = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "resilient_spoke_available_count",
	Help: "Count times the Resilient Spoke cluster was reported available",
}, []string{LabelSpokeName})

var NewClusterClaimCreated = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "new_cluster_claim_created",
	Help: "Count the times we created a new ClusterClaim for Hive",
}, []string{LabelPoolName, LabelClaimName, LabelOldSpokeName})

var NewSpokeReady = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "new_spoke_ready",
	Help: "Count the time we got a new ready cluster",
}, []string{LabelOldSpokeName, LabelNewSpokeName})

// init is registering our metrics with K8S registry.
func init() {
	metrics.Registry.MustRegister(
		ResilientSpokeNotAvailable,
		ResilientSpokeAvailable,
		NewClusterClaimCreated,
		NewSpokeReady,
	)
}
