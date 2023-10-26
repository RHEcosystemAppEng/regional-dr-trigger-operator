// Copyright (c) 2023 Red Hat, Inc.

package reconcilers

// This file contains options and functions for loading all the registered reconciler processes.

import "sigs.k8s.io/controller-runtime/pkg/manager"

type Options struct {
	ConfigMapName string
}

// reconcilerFuncs is used for registering reconciler funcs for setup.
var reconcilerFuncs []func(mgr manager.Manager, options Options) error

// Setup is used for setting up all the registered reconciler controllers with a manager.
func Setup(mgr manager.Manager, options Options) error {
	for _, f := range reconcilerFuncs {
		if err := f(mgr, options); err != nil {
			return err
		}
	}
	return nil
}
