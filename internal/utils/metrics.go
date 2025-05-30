// Copyright (c) 2023 Red Hat, Inc.

package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	DRClusterName     = "dr_cluster_name"
	DRControlName     = "dr_control_name"
	DRApplicationName = "dr_application_name"
)

var DRApplicationFailoverMetric = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dr_application_failover_count",
	Help: "Counter for DR Applications failover initiated by the Regional DR Trigger Operator",
}, []string{DRClusterName, DRControlName, DRApplicationName})

func init() {
	metrics.Registry.MustRegister(DRApplicationFailoverMetric)
}
