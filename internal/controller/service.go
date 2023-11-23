package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MLFlowReconciler) GetMLFlowService(ctx context.Context, name string, namespace string, service *corev1.Service) error {
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	return r.K8sClient.Get(ctx, namespacedName, service)
}

func (r *MLFlowReconciler) CreateMLFlowService(ctx context.Context, service *corev1.Service) (*corev1.Service, error) {
	logger := log.FromContext(ctx)
	existService := &corev1.Service{}
	err := r.GetMLFlowService(ctx, service.Name, service.Namespace, existService)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		} else {
			if err := r.K8sClient.Create(ctx, service); err != nil {
				return nil, err
			} else {
				return service, nil
			}
		}
	} else {
		if r.IsThereAnyChangeOnService(service, existService) {
			logger.Info("Updating Service")
			existService.Spec = service.Spec
			err := r.K8sClient.Update(ctx, existService)
			if err != nil {
				return nil, err
			} else {
				return service, nil
			}
		} else {
			return existService, nil
		}
	}
}

func (r *MLFlowReconciler) IsThereAnyChangeOnService(oldService *corev1.Service, currentService *corev1.Service) bool {
	return !equality.Semantic.DeepDerivative(oldService.Spec, currentService.Spec)
}
