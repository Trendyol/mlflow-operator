package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *MLFlowReconciler) GetMLFlowDeployment(ctx context.Context, name string, namespace string, deployment *appsv1.Deployment) error {
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	return r.K8sClient.Get(ctx, namespacedName, deployment)
}

func (r *MLFlowReconciler) CreateMLFlowDeployment(ctx context.Context, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	logger := log.FromContext(ctx)
	existDeployment := &appsv1.Deployment{}
	err := r.GetMLFlowDeployment(ctx, deployment.Name, deployment.Namespace, existDeployment)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		} else {
			if err := r.K8sClient.Create(ctx, deployment); err != nil {
				return nil, err
			} else {
				return deployment, nil
			}
		}
	} else {
		if r.IsThereAnyChangeOnDeployment(deployment, existDeployment) {
			logger.Info("Updating Deployment")
			existDeployment.Spec = deployment.Spec
			err := r.K8sClient.Update(ctx, existDeployment)
			if err != nil {
				return nil, err
			} else {
				return deployment, nil
			}
		} else {
			return existDeployment, nil
		}
	}
}

func (r *MLFlowReconciler) DeploymentIsNotReady(deployment *appsv1.Deployment) bool {
	return *deployment.Spec.Replicas != deployment.Status.ReadyReplicas
}

func (r *MLFlowReconciler) IsThereAnyChangeOnDeployment(oldDeployment *appsv1.Deployment, currentDeployment *appsv1.Deployment) bool {
	return !equality.Semantic.DeepDerivative(oldDeployment.Spec, currentDeployment.Spec)
}
