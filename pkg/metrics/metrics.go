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
	Name: "resilient_spoke_not_available_count",
	Help: "Count times the Resilient Spoke cluster was reported not available",
}, []string{DRClusterName})

var DRApplicationFailover = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "resilient_spoke_application_failover_count",
	Help: "Count times the Resilient Spoke Application failed over",
}, []string{DRClusterName, DRControlName, DRApplicationName})

// init is registering the metrics with K8S registry.
func init() {
	metrics.Registry.MustRegister(
		DRClusterNotAvailable,
		DRApplicationFailover,
	)
}
