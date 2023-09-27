// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/lease"
	ctrl "sigs.k8s.io/controller-runtime"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Multicluster Resiliency Addon Agent",
	Long:  "multicluster-resiliency-addon agent TODO add description",
	RunE:  runAgent,
}

func init() {
	mcraCmd.AddCommand(agentCmd)
}

func runAgent(cmd *cobra.Command, args []string) error {
	klog.Info("running agent")
	kubeConfig, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		klog.Errorf("unable to get kubeconfig: %v", err)
		return err
	}

	leaseUpdater := lease.NewLeaseUpdater(kubeConfig, addonName, agentNamespace)
	go func() {
		leaseUpdater.Start(cmd.Context())
	}()

	klog.Info("agent done")
	return nil
}
