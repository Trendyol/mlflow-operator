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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
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
	client.Client
	Scheme           *runtime.Scheme
	HttpClient       *util.HttpClient
	MlflowService    *mlflow.Service
	MlflowKubernetes *mlflow.Kubernetes
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
func (r *MLFlowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mlflowServerConfig := mlflowv1beta1.MLFlow{}

	if err := r.Get(ctx, req.NamespacedName, &mlflowServerConfig); err != nil {
		logger.Error(err, "unable to fetch MlflowServerConfig")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	} else {
		if r.MlflowService == nil {
			r.MlflowService = mlflow.NewService(mlflowServerConfig.Name, mlflowServerConfig.Namespace, r.HttpClient)
		}
	}

	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(req.Namespace), client.MatchingLabels{"app": "mlflow-server"}); err != nil {
		logger.Error(err, "unable to list deployments")
		return reconcile.Result{}, err
	}

	deployment, err := r.MlflowKubernetes.CreateMlflowDeployment(req.Name, req.Namespace, &mlflowServerConfig)

	existingDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, existingDeployment)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating Deployment", deployment.Namespace, deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			logger.Error(err, "unable to create Deployment for MlflowServerConfig")
			return reconcile.Result{}, err
		}

		err = r.updateStatus(ctx, &mlflowServerConfig, deployment)
		if err != nil {
			logger.Error(err, "unable to update status for MlflowServerConfig when creating Deployment")
			return reconcile.Result{}, err
		}

		logger.Info("Creating Service", deployment.Namespace, deployment.Name)
		service, err := r.MlflowKubernetes.CreateMlflowService(req.Name, req.Namespace, &mlflowServerConfig)
		if err != nil {
			logger.Error(err, "unable to create Service for MlflowServerConfig when creating service")
			return reconcile.Result{}, err
		}

		err = r.Create(ctx, service)
		if err != nil {
			logger.Error(err, "unable to create Service for MlflowServerConfig when pushing to k8s")
			return reconcile.Result{}, err
		}
	} else if err == nil {
		if *existingDeployment.Spec.Replicas != existingDeployment.Status.ReadyReplicas {
			logger.Info("Waiting for Deployment to be ready", deployment.Namespace, deployment.Name)
			return reconcile.Result{Requeue: true}, nil
		} else {
			logger.Info("Deployment is ready", deployment.Namespace, deployment.Name)
			models, err := r.MlflowService.GetLatestModelVersion()
			if err != nil {
				logger.Error(err, "unable to get latest model version")
				return reconcile.Result{}, err
			}

			for _, model := range models {
				logger.Info("Found model", model.Name, model.Version)
				modelDeployment, err := r.MlflowKubernetes.CreateMlflowModelDeployment(req.Name, req.Namespace, &mlflowServerConfig, model.Name, model.Version)
				if err != nil {
					logger.Error(err, "unable to create Deployment for Model when creating model deployment")
					return reconcile.Result{}, err
				}

				err = r.Create(ctx, modelDeployment)
				if err != nil {
					logger.Error(err, "unable to create Deployment for Model when pushing to k8s")
					return reconcile.Result{}, err
				}
			}
		}

		if !equality.Semantic.DeepDerivative(deployment.Spec, existingDeployment.Spec) {
			existingDeployment.Spec = deployment.Spec
			logger.Info("Updating Deployment", deployment.Namespace, deployment.Name)
			err = r.Update(ctx, existingDeployment)
			if err != nil {
				logger.Error(err, "unable to update Deployment for MlflowServerConfig")
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{}, err
}

func (r *MLFlowReconciler) updateStatus(ctx context.Context, mlflowServerConfig *mlflowv1beta1.MLFlow, obj runtime.Object) error {
	ref, err := reference.GetReference(r.Scheme, obj)
	if err != nil {
		return err
	}

	mlflowServerConfig.Status.Active = *ref

	if err = r.Status().Update(ctx, mlflowServerConfig); err != nil {
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
