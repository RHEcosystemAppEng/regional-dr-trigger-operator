// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	goflag "flag"
	"time"

	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/rand"
	cliflag "k8s.io/component-base/cli/flag"
)

var mcraCmd = &cobra.Command{
	Use:          "mcra",
	Short:        "Multicluster Resiliency Addon",
	SilenceUsage: true,
	Version:      version.Get().String(),
}

func Execute() error {
	rand.Seed(uint64(time.Now().UTC().UnixNano()))

	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	return mcraCmd.Execute()
}
