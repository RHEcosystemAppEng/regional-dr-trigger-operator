// Copyright (c) 2023 Red Hat, Inc.

package controller

// This file hosts our reconciler implementation for the controller run.

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type McraReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *McraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("TODO-ADD-RECONCILE-LOGIC")
	// TODO
	return ctrl.Result{Requeue: false}, nil
}
