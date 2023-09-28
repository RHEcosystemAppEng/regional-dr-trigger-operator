// Copyright (c) 2023 Red Hat, Inc.

package manager

import (
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/manager"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"github.com/spf13/cobra"
	"open-cluster-management.io/addon-framework/pkg/cmd/factory"
)

func GetCommand() *cobra.Command {
	managerCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-manager", version.Get(), manager.Run).
		NewCommand()
	managerCmd.Use = "manager"
	managerCmd.Short = "Multicluster Resiliency Addon Manager"

	return managerCmd
}
