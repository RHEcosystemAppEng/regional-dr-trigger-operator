// Copyright (c) 2023 Red Hat, Inc.

package agent

// This file hosts functions and types for running our Addon Agent on the Spoke cluster.

import (
	"context"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/mcra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"open-cluster-management.io/addon-framework/pkg/lease"
)

// Agent is a receiver representing the Addon agent. It encapsulates the Agent Options used for configuring the agent run.
// Use NewAgent for instantiation.
type Agent struct {
	Options *Options
}

// Options is used for encapsulating the various options for configuring the agent Run.
type Options struct {
	HubKubeConfigFile string
	SpokeName         string
	AgentNamespace    string
}

// NewAgent is used as a factory for creating an Agent instance with an Options instance.
func NewAgent() Agent {
	return Agent{Options: &Options{}}
}

// Run is used for running the Addon agent. It takes a context and the kubeconfig for the Spoke it runs on. This
// function blocks while waiting for the context to be done.
func (a *Agent) Run(ctx context.Context, kubeConfig *rest.Config) error {
	spokeClientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	hubConfig, err := clientcmd.BuildConfigFromFlags("", a.Options.HubKubeConfigFile)
	if err != nil {
		return err
	}

	leaseUpdater := lease.
		NewLeaseUpdater(spokeClientSet, mcra.AddonName, a.Options.AgentNamespace).
		WithHubLeaseConfig(hubConfig, a.Options.SpokeName)

	go func() {
		leaseUpdater.Start(ctx)
	}()

	// blocking
	<-ctx.Done()

	return nil
}
