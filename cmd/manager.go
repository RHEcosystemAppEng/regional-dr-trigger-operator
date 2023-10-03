// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/manager"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"open-cluster-management.io/addon-framework/pkg/cmd/factory"
)

func init() {
	mgr := manager.NewManager()

	mgrCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-manager", version.Get(), mgr.Run).
		NewCommand()
	mgrCmd.Use = "manager"
	mgrCmd.Short = "Multicluster Resiliency Addon Manager"

	mgrCmd.Flags().IntVar(&mgr.Options.AgentReplicas, "agent-replicas", 1, "TODO")

	mcraCmd.AddCommand(mgrCmd)
}
