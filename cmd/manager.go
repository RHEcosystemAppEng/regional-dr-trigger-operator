// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	"embed"
	"github.com/spf13/cobra"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
)

//go:embed manifests
var FS embed.FS

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Multicluster Resiliency Addon Manager",
	Long:  "multicluster-resiliency-addon manager TODO add description",
	RunE:  runManager,
}

func init() {
	mcraCmd.AddCommand(managerCmd)
}

func runManager(cmd *cobra.Command, args []string) error {
	klog.Info("running manager")
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		klog.Errorf("unable to get kubeconfig: %v", err)
		return err
	}

	addonMgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		klog.Errorf("unable to setup addon manager: %v", err)
		return err
	}

	agentAddon, err := addonfactory.NewAgentAddonFactory(addonName, FS, "manifests").BuildTemplateAgentAddon()
	if err != nil {
		klog.Errorf("failed to build agent addon %v", err)
		return err
	}

	if err = addonMgr.AddAgent(agentAddon); err != nil {
		klog.Errorf("failed to add addon agent: %v", err)
		return err
	}

	go func() {
		if err := addonMgr.Start(cmd.Context()); err != nil {
			klog.Errorf("failed to add start addon: %v", err)
		}
	}()

	<-cmd.Context().Done()

	klog.Info("manager done")
	return nil
}
