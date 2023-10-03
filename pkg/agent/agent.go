// Copyright (c) 2023 Red Hat, Inc.

package agent

import (
	"context"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/manager"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/lease"
)

// Agent is a receiver representing the Addon Agent.
// It encapsulates the Agent Options which will be used to configure the Agent run.
// Use NewAgent for instantiation.
type Agent struct {
	Options *Options
}

// Options is used for encapsulating the various options for configuring the Agent Run.
type Options struct {
	HubKubeConfigFile string
	SpokeName         string
	AgentNamespace    string
}

// NewAgent is used as a factory for creating an Agent instance.
func NewAgent() Agent {
	return Agent{Options: &Options{}}
}

// Run is used for running the Addon Agent.
// It takes a context and the kubeconfig for the Spoke it runs on.
func (a *Agent) Run(ctx context.Context, kubeConfig *rest.Config) error {
	klog.Info("running addon agent")

	spokeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	hubConfig, err := clientcmd.BuildConfigFromFlags("", a.Options.HubKubeConfigFile)
	if err != nil {
		return err
	}

	leaseUpdater := lease.
		NewLeaseUpdater(spokeClient, manager.AddonName, a.Options.AgentNamespace).
		WithHubLeaseConfig(hubConfig, a.Options.SpokeName)

	go func() {
		leaseUpdater.Start(ctx)
	}()

	<-ctx.Done()

	klog.Info("addon agent done")
	return nil
}
