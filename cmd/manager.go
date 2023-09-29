// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/manager"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"open-cluster-management.io/addon-framework/pkg/cmd/factory"
)

func init() {
	managerCmd := factory.
		NewControllerCommandConfig("multicluster-resiliency-addon-manager", version.Get(), manager.Run).
		NewCommand()
	managerCmd.Use = "manager"
	managerCmd.Short = "Multicluster Resiliency Addon Manager"

	mcraCmd.AddCommand(managerCmd)
}
