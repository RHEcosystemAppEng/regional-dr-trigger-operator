// Copyright (c) 2023 Red Hat, Inc.

package actions

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Options struct {
	client.Client
	OldSpoke, NewSpoke, ConfigName string
}

// actionFuncs is used for registering actions to be performs when replacing clusters.
var actionFuncs []func(ctx context.Context, options Options)

// PerformReplace is used to perform all registered actions related to replacing the cluster.
func PerformReplace(ctx context.Context, options Options) {
	for _, f := range actionFuncs {
		f(ctx, options)
	}
}
