// Copyright (c) 2023 Red Hat, Inc.

package cmd

import (
	goflag "flag"
	"github.com/rhecosystemappeng/multicluster-resiliency-addon/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
	cliflag "k8s.io/component-base/cli/flag"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Root MCRA Command.
var mcraCmd = &cobra.Command{
	Use:              "mcra",
	Short:            "Multicluster Resiliency Addon",
	SilenceUsage:     true,
	PersistentPreRun: configureLogging,
	Version:          version.Get().String(),
}

// Execute will execute the root MCRA Command.
// Note that sub-commands are added via the various init functions in this package.
func Execute() error {
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	return mcraCmd.Execute()
}

// configureLogging is used for configuring and setting our k8s environment logger persistently for every command.
func configureLogging(cmd *cobra.Command, args []string) {
	logOpts := &zap.Options{
		Encoder: zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			CallerKey:    "caller",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.FullCallerEncoder,
		}),
	}
	logOpts.BindFlags(goflag.CommandLine)

	logger := zap.New(zap.UseFlagOptions(logOpts))
	k8slog.SetLogger(logger)
}
