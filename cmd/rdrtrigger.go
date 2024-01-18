// Copyright (c) 2023 Red Hat, Inc.

package cmd

// This file hosts the root rdrtrigger command.

import (
	"context"
	goflag "flag"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/util/rand"
	k8sflag "k8s.io/component-base/cli/flag"
	"regional-dr-trigger-operator/pkg/controller"
	"regional-dr-trigger-operator/pkg/version"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

// command used for running the Regional DR Trigger Operator
var cmd = &cobra.Command{
	Use:     "rdrtrigger",
	Short:   "Regional DR Trigger Operator, ACM-based triggering",
	PreRun:  configureLogging,
	Version: version.Get().String(),
}

// init is used for binding the flags and the controller to the command
func init() {
	ctrl := controller.NewDRTriggerController()

	cmd.Flags().StringVar(&ctrl.Options.MetricAddr, "metric-address", ":8080", "TODO")
	cmd.Flags().StringVar(&ctrl.Options.ProbeAddr, "probe-address", ":8081", "TODO")
	cmd.Flags().BoolVar(&ctrl.Options.LeaderElection, "leader-election", false, "TODO")

	cmd.RunE = ctrl.Run
}

// Execute will execute the root Regional DR Trigger (rdrtrigger) command which, in turn, will run the controller.
func Execute() error {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(k8sflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	return cmd.ExecuteContext(context.Background())
}

// configureLogging is used for configuring and setting the k8s environment logger.
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
