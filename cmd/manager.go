// Copyright (c) 2023 Red Hat, Inc.

package cmd

// This file hosts the 'manager' command used for running the Addon Manager on a Hub cluster.

import (
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/manager"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"open-cluster-management.io/addon-framework/pkg/cmd/factory"
)

// init is used for creating the Manager Commend, incorporate its flags, and binding it to the root MCRA Command.
func init() {
	mgr := manager.NewManager()

	mgrCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-manager", version.Get(), mgr.Run).
		NewCommand()
	mgrCmd.Use = "manager"
	mgrCmd.Short = "MultiCluster Resiliency Addon Manager"

	mgrCmd.Flags().StringVar(&mgr.Options.ControllerMetricAddr, "controller-metric-address", ":8080", "TODO")
	mgrCmd.Flags().StringVar(&mgr.Options.ControllerProbeAddr, "controller-probe-address", ":8081", "TODO")
	mgrCmd.Flags().BoolVar(&mgr.Options.ControllerLeaderElection, "controller-leader-election", false, "TODO")
	mgrCmd.Flags().StringVar(&mgr.Options.AgentImage, "agent-image", "", "TODO")
	mgrCmd.Flags().StringVar(&mgr.Options.ServiceAccount, "service-account", "", "TODO")
	mgrCmd.Flags().StringVar(&mgr.Options.ConfigMapName, "configmap-name", "", "TODO")
	mgrCmd.Flags().BoolVar(&mgr.Options.EnableValidation, "enable-validation-webhook", false, "TODO")

	mcraCmd.AddCommand(mgrCmd)
}
