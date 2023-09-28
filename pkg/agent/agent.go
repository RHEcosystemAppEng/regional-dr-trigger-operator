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

type Options struct {
	HubKubeConfigFile string
	SpokeName         string
	AddonName         string
	AgentNamespace    string
}

func Run(options *Options) func(ctx context.Context, kubeConfig *rest.Config) error {
	return func(ctx context.Context, kubeConfig *rest.Config) error {
		klog.Info("running agent")

		flag.Parse()

		klog.Info("tomer tomer debug 1")
		klog.Info(options)

		spokeClient, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}

		hubConfig, err := clientcmd.BuildConfigFromFlags("", options.HubKubeConfigFile)
		if err != nil {
			return err
		}

		leaseUpdater := lease.
			NewLeaseUpdater(spokeClient, options.AddonName, options.AgentNamespace).
			WithHubLeaseConfig(hubConfig, options.SpokeName)

		go func() {
			leaseUpdater.Start(ctx)
		}()

		klog.Info("tomer debug1")
		<-ctx.Done()

		klog.Info("agent done")
		return nil
	}
}
