// Copyright (c) 2023 Red Hat, Inc.

package agent

import (
	"context"
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/lease"
)

type Agent struct {
	Options *Options
}

type Options struct {
	HubKubeConfigFile string
	SpokeName         string
	AddonName         string
	AgentNamespace    string
}

func (a *Agent) Run(ctx context.Context, kubeConfig *rest.Config) error {
	klog.Info("running agent")

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
		NewLeaseUpdater(spokeClient, a.Options.AddonName, a.Options.AgentNamespace).
		WithHubLeaseConfig(hubConfig, a.Options.SpokeName)

	go func() {
		leaseUpdater.Start(ctx)
	}()

	<-ctx.Done()

	return nil
}
