// Copyright (c) 2023 Red Hat, Inc.

package main

import (
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
	"regional-dr-trigger-operator/pkg/operator"
	"regional-dr-trigger-operator/pkg/version"
)

// command used for running the operator
var cmd = &cobra.Command{
	Use:   "rdrtrigger",
	Short: "Regional DR Trigger Operator, ACM-based triggering",
}

// init is used for binding the flags and the controller to the command
func init() {
	oper := operator.NewDRTriggerOperator()

	cmd.Flags().StringVar(
		&oper.Options.MetricAddr,
		"metric-address",
		":8080",
		"The address the metric endpoint binds to.")
	cmd.Flags().StringVar(
		&oper.Options.ProbeAddr,
		"probe-address",
		":8081",
		"The address the probe endpoint binds to.")
	cmd.Flags().BoolVar(
		&oper.Options.LeaderElection,
		"leader-election",
		false,
		"Enable leader election for controllers manager.")
	cmd.Flags().BoolVar(
		&oper.Options.Debug,
		"debug",
		false,
		"Enable debug logging")

	cmd.RunE = oper.Run
	cmd.Version = version.Get().String()
}

// main is used for running the regional dr trigger operator command
func main() {
	if err := cli.RunNoErrOutput(cmd); err != nil {
		panic(err)
	}
}
