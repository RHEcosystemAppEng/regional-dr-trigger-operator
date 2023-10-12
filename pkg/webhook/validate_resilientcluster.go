package webhook

import (
	"context"
	v1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-resilientcluster,mutating=false,failurePolicy=fail,groups=appeng.ecosystem.redhat.com,resources=resilientclusters,versions=v1,name=resilientcluster.appeng.ecosystem,sideEffects=None,admissionReviewVersions=v1

type ValidateResilientCluster struct {
}

func (v *ValidateResilientCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1.ResilientCluster{}).WithValidator(v).Complete()
}

func (v *ValidateResilientCluster) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	logger := log.FromContext(ctx)
	logger.Info("TODO add ValidateCreate implementation")
	return nil, nil
}

func (v *ValidateResilientCluster) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	logger := log.FromContext(ctx)
	logger.Info("TODO add ValidateUpdate implementation")
	return nil, nil
}

func (v *ValidateResilientCluster) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	logger := log.FromContext(ctx)
	logger.Info("TODO add ValidateDelete implementation")
	return nil, nil
}
