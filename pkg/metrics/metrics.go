// Copyright (c) 2023 Red Hat, Inc.

package metrics

// This file contains various metrics for use throughout the project.

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	DRClusterName     = "dr_cluster_name"
	DRControlName     = "dr_control_name"
	DRApplicationName = "dr_application_name"
)

var DRClusterNotAvailable = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dr_cluster_not_available_count",
	Help: "Counter for DR clusters identified as not available",
}, []string{DRClusterName})

var DRApplicationFailover = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dr_application_failover_count",
	Help: "Counter for DR application failover performed",
}, []string{DRClusterName, DRControlName, DRApplicationName})

// init is registering the metrics with K8S registry.
func init() {
	metrics.Registry.MustRegister(
		DRClusterNotAvailable,
		DRApplicationFailover,
	)
}
