package controller

import (
	"context"

	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *MLFlowReconciler) CreateMLFlowDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	logger := log.FromContext(ctx)
	logger.Info("Creating MLFlow Deployment")
	return r.K8sClient.Create(ctx, deployment)
}

func (r *MLFlowReconciler) GetMLFlowDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	namespacedName := types.NamespacedName{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
	}
	return r.K8sClient.Get(ctx, namespacedName, deployment)
}

func (r *MLFlowReconciler) CreateMLFlowModelDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	if err := r.K8sClient.Create(ctx, deployment); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (r *MLFlowReconciler) UpdateDeploymentStatus(ctx context.Context, mlflowServerConfig *mlflowv1beta1.MLFlow, deployment *appsv1.Deployment) error {
	return r.updateStatus(ctx, mlflowServerConfig, deployment)
}

func (r *MLFlowReconciler) UpdateDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	return r.K8sClient.Update(ctx, deployment)
}

func (r *MLFlowReconciler) DeploymentIsNotReady(deployment *appsv1.Deployment) bool {
	return *deployment.Spec.Replicas != deployment.Status.ReadyReplicas
}

func (r *MLFlowReconciler) IsThereAnyChangeOnDeployment(oldDeployment *appsv1.Deployment, currentDeployment *appsv1.Deployment) bool {
	return equality.Semantic.DeepDerivative(oldDeployment.Spec, currentDeployment.Spec)
}
