// Copyright (c) 2023 Red Hat, Inc.

package cmd

// This file hosts the root rdrtrigger command.

import (
	"context"
	goflag "flag"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"regional-dr-trigger-operator/pkg/controller"
	"regional-dr-trigger-operator/pkg/version"
)

// Execute will execute the root Regional DR Trigger (rdrtrigger) Command.
func Execute() error {
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	ctrl := controller.NewDRTriggerController()
	cmdConfig := controllercmd.NewControllerCommandConfig("regional-dr-trigger-operator", version.Get(), ctrl.Run)
	cmd := cmdConfig.NewCommandWithContext(context.Background())

	cmd.Use = "rdrtrigger"
	cmd.Short = "Regional DR Trigger Operator, ACM-based triggering"

	cmd.Flags().StringVar(&ctrl.Options.MetricAddr, "metric-address", ":8080", "TODO")
	cmd.Flags().StringVar(&ctrl.Options.ProbeAddr, "probe-address", ":8081", "TODO")
	cmd.Flags().BoolVar(&ctrl.Options.LeaderElection, "leader-election", false, "TODO")

	return cmd.Execute()
}
