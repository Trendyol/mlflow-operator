/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	"github.com/Trendyol/mlflow-operator/internal/mlflow"
	"github.com/Trendyol/mlflow-operator/internal/util"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// MLFlowReconciler reconciles a MLFlow object
type MLFlowReconciler struct {
	K8sClient           client.Client
	Scheme              *runtime.Scheme
	HTTPClient          util.HTTPClient
	MlflowClient        *mlflow.Client
	MlflowObjectManager *mlflow.ObjectManager
	Debug               bool
}

//+kubebuilder:rbac:groups=mlflow.trendyol.com,resources=mlflows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlflow.trendyol.com,resources=mlflows/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlflow.trendyol.com,resources=mlflows/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// nolint:funlen
func (r *MLFlowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mlflowServerConfig := mlflowv1beta1.MLFlow{}

	if err := r.GetMlflowCRD(ctx, req.NamespacedName, &mlflowServerConfig); err != nil {
		logger.Error(err, "unable to fetch mlflow server config")
		return reconcile.Result{}, err
	}

	deployment, err := r.MlflowObjectManager.CreateMlflowDeploymentObject(req.Name, req.Namespace, &mlflowServerConfig)
	if err != nil {
		logger.Error(err, "unable to set ownership on deployment resource")
		return reconcile.Result{}, err
	}

	if r.Debug {
		if err := r.ConfigureVolumesAndEnvs(ctx, req, &mlflowServerConfig, deployment); err != nil {
			logger.Error(err, "unable to set volume related things")
			return reconcile.Result{}, err
		}
	}

	existingDeployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: deployment.Name, Namespace: deployment.Namespace}}
	if err = r.GetMLFlowDeployment(ctx, existingDeployment); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "unable to get Deployment")
			return reconcile.Result{}, err
		}

		if err := r.CreateMLFlowDeployment(ctx, deployment); err != nil {
			logger.Error(err, "unable to create Deployment for MlflowServerConfig")
			return reconcile.Result{}, err
		}

		if err = r.UpdateDeploymentStatus(ctx, &mlflowServerConfig, deployment); err != nil {
			logger.Error(err, "unable to update status for MlflowServerConfig when creating Deployment")
			return reconcile.Result{}, err
		}

		service, err := r.MlflowObjectManager.CreateMlflowServiceObject(req.Name, req.Namespace, &mlflowServerConfig)
		if err != nil {
			logger.Error(err, "unable to create Client for MlflowServerConfig when creating service")
			return reconcile.Result{}, err
		}

		if err = r.CreateMLFlowService(ctx, service); err != nil {
			logger.Error(err, "unable to create Client for MlflowServerConfig when pushing to k8s")
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	if r.DeploymentIsNotReady(existingDeployment) {
		logger.Info("Waiting for Deployment to be ready", deployment.Namespace, deployment.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	logger.Info("Deployment is ready")

	if r.Debug {
		if err := r.createTestModel(ctx, req, mlflowServerConfig); err != nil {
			logger.Error(err, "unable to create Job for MlflowServerConfig when creating job for test model")
			return reconcile.Result{}, err
		}
	}

	r.InitializeMlflowClient(&mlflowServerConfig)

	models, err := r.MlflowClient.GetLatestModels()
	if err != nil {
		logger.Error(err, "unable to get latest models")
		return reconcile.Result{}, err
	}

	for i := range models {
		modelDeployment, err := r.MlflowObjectManager.CreateMlflowModelDeploymentObject(req.Name, req.Namespace, &mlflowServerConfig, models[i])
		if err != nil {
			logger.Error(err, "unable to create Deployment for Model when creating model deployment")
			return reconcile.Result{}, err
		}

		if err := r.CreateMLFlowModelDeployment(ctx, modelDeployment); err != nil {
			logger.Error(err, "unable to create Deployment for Model when pushing to k8s")
			return reconcile.Result{}, err
		}
	}

	if r.IsThereAnyChangeOnDeployment(deployment, existingDeployment) {
		return reconcile.Result{}, nil
	}

	existingDeployment.Spec = deployment.Spec
	logger.Info("Updating Deployment")

	if err := r.UpdateDeployment(ctx, existingDeployment); err != nil {
		logger.Error(err, "unable to update Deployment for MlflowServerConfig")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *MLFlowReconciler) createTestModel(ctx context.Context, req ctrl.Request, mlflowServerConfig mlflowv1beta1.MLFlow) error {
	job, err := r.MlflowObjectManager.CreateMlflowWineQualityJobObject(req.Name, req.Namespace, &mlflowServerConfig)
	if err != nil {
		return err
	}

	if err := r.CreateMLFlowWineQualityJob(ctx, job); err != nil {
		return err
	}
	return nil
}

func (r *MLFlowReconciler) GetMlflowCRD(ctx context.Context, namespace types.NamespacedName, mlflowServerCfg *mlflowv1beta1.MLFlow) error {
	return r.K8sClient.Get(ctx, namespace, mlflowServerCfg)
}

func (r *MLFlowReconciler) InitializeMlflowClient(mlflowServerCfg *mlflowv1beta1.MLFlow) {
	if r.MlflowClient != nil {
		return
	}

	r.MlflowClient = mlflow.NewClient(mlflowServerCfg, r.HTTPClient, r.Debug)
}

func (r *MLFlowReconciler) CreateMLFlowService(ctx context.Context, service *corev1.Service) error {
	logger := log.FromContext(ctx)
	logger.Info("Creating Service")
	return r.K8sClient.Create(ctx, service)
}

func (r *MLFlowReconciler) CreateMLFlowWineQualityJob(ctx context.Context, job *batchv1.Job) error {
	if err := r.K8sClient.Create(ctx, job); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (r *MLFlowReconciler) updateStatus(ctx context.Context, mlflowServerConfig *mlflowv1beta1.MLFlow, obj runtime.Object) error {
	ref, err := reference.GetReference(r.Scheme, obj)
	if err != nil {
		return err
	}

	mlflowServerConfig.Status.Active = *ref

	if err = r.K8sClient.Status().Update(ctx, mlflowServerConfig); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MLFlowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlflowv1beta1.MLFlow{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
