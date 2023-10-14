// Copyright (c) 2023 Red Hat, Inc.

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const LabelSpokeName = "spoke_name"

var ResilientSpokeNotAvailable = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "resilient_spoke_not_available_count",
	Help: "Count times the Resilient Spoke cluster was not available",
}, []string{LabelSpokeName})

var ResilientSpokeAvailable = *prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "resilient_spoke_available_count",
	Help: "Count times the Resilient Spoke cluster was available",
}, []string{LabelSpokeName})

func init() {
	metrics.Registry.MustRegister(ResilientSpokeNotAvailable, ResilientSpokeAvailable)
}
