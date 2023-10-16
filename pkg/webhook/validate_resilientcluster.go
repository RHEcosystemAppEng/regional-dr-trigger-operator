package webhook

import (
	"context"
	"errors"
	"fmt"
	v1 "github.com/rhecosystemappeng/multicluster-resiliency-addon/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
)

// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-appeng-ecosystem-redhat-com-v1-resilientcluster,mutating=false,failurePolicy=fail,groups=appeng.ecosystem.redhat.com,resources=resilientclusters,versions=v1,name=resilientcluster.appeng.ecosystem,sideEffects=None,admissionReviewVersions=v1

type ValidateResilientCluster struct {
	client.Client
	ServiceAccount string
}

func (v *ValidateResilientCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1.ResilientCluster{}).WithValidator(v).Complete()
}

func (v *ValidateResilientCluster) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if err := v.verifyUser(ctx); err != nil {
		return nil, err
	}
	return nil, v.verifyOnlyOneInNamespace(ctx)
}

func (v *ValidateResilientCluster) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	return nil, v.verifyUser(ctx)
}

func (v *ValidateResilientCluster) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, v.verifyUser(ctx)
}

func (v *ValidateResilientCluster) verifyUser(ctx context.Context) error {
	logger := log.FromContext(ctx)
	request, err := admission.RequestFromContext(ctx)
	if err != nil {
		logger.Error(err, "failed to parse admission request")
		return err
	}

	userArr := strings.Split(request.UserInfo.Username, ":")
	if userArr[0] == "system" &&
		userArr[1] == "serviceaccount" &&
		userArr[3] == v.ServiceAccount {
		return nil
	}

	return errors.New("user not allowed to control ResilientCluster")
}

func (v *ValidateResilientCluster) verifyOnlyOneInNamespace(ctx context.Context) error {
	rstcList := &v1.ResilientClusterList{}
	if err := v.Client.List(ctx, rstcList); err != nil {
		return err
	}
	request, err := admission.RequestFromContext(ctx)
	if err != nil {
		return err
	}

	for _, rstc := range rstcList.Items {
		if rstc.Namespace == request.Namespace {
			return fmt.Errorf("%s already has a ResilientCluster", rstc.Namespace)
		}
	}

	return nil
}
