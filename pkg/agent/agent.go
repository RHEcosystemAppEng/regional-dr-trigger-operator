// Copyright (c) 2023 Red Hat, Inc.

package agent

import (
	"context"
	"flag"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/manager"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	klogv2 "k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/lease"
)

type Agent struct {
	Options *Options
}

type Options struct {
	HubKubeConfigFile string
	SpokeName         string
	AgentNamespace    string
}

// Run is used for running the Addon Agent
func (a *Agent) Run(ctx context.Context, kubeConfig *rest.Config) error {
	klogv2.Info("running agent")

	flag.Parse()

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

	return nil
}
