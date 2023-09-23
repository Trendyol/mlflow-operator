package controller

import (
	"context"
	"fmt"

	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MLFlowReconciler) ConfigureVolumesAndEnvs(ctx context.Context, req ctrl.Request, mlflowServerConfig *mlflowv1beta1.MLFlow, deployment *appsv1.Deployment) error {
	deployment.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{}
	deployment.Spec.Template.Spec.Containers[0].Args = []string{
		"server",
		"--serve-artifacts",
		"--host",
		"0.0.0.0",
	}

	simulatedVolumes, err := r.GetSimulatedVolumes(ctx, req, mlflowServerConfig)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		return nil
	}

	deployment.Spec.Template.Spec.Volumes = r.MlflowResourceManager.CreateVolumeObject(simulatedVolumes)
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = r.MlflowResourceManager.CreateVolumeMountObject(simulatedVolumes)

	return nil
}

func (r *MLFlowReconciler) GetSimulatedVolumes(ctx context.Context, req ctrl.Request, mlflowServerConfig *mlflowv1beta1.MLFlow) ([]string, error) {
	mlartifactsPvc, err := r.CreateMlArtifactsPVC(ctx, req, mlflowServerConfig)
	if err != nil {
		return nil, err
	}

	mlrunsPvc, err := r.CreateMlrunsPVC(ctx, req, mlflowServerConfig)
	if err != nil {
		return nil, err
	}

	var volumes []string
	volumes = append(volumes, mlartifactsPvc.Name)
	volumes = append(volumes, mlrunsPvc.Name)

	return volumes, nil
}

func (r *MLFlowReconciler) CreateMlArtifactsPVC(ctx context.Context, req ctrl.Request, mlflowServerConfig *mlflowv1beta1.MLFlow) (*corev1.PersistentVolumeClaim, error) {
	mlartifactsPvc, err := r.MlflowResourceManager.CreateMlflowPVCObject(req.Name, req.Namespace, "mlartifacts", mlflowServerConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create MlflowPersistence for mlflowserverconfig %w", err)
	}

	err = r.K8sClient.Create(ctx, mlartifactsPvc)
	if err != nil {
		return nil, err
	}

	return mlartifactsPvc, nil
}

func (r *MLFlowReconciler) CreateMlrunsPVC(ctx context.Context, req ctrl.Request, mlflowServerConfig *mlflowv1beta1.MLFlow) (*corev1.PersistentVolumeClaim, error) {
	mlrunsPvc, err := r.MlflowResourceManager.CreateMlflowPVCObject(req.Name, req.Namespace, "mlruns", mlflowServerConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create MlflowPersistence for mlflowserverconfig %w", err)
	}

	err = r.K8sClient.Create(ctx, mlrunsPvc)
	if err != nil {
		return nil, err
	}

	return mlrunsPvc, nil
}
