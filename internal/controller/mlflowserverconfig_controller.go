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
	"encoding/json"
	"fmt"
	mlflowserverv1 "github.com/Trendyol/mlflow-operator/api/v1"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/reference"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// MlflowServerConfigReconciler reconciles a MlflowServerConfig object
type MlflowServerConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mlflow.server.trendyol.com,resources=mlflowserverconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlflow.server.trendyol.com,resources=mlflowserverconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlflow.server.trendyol.com,resources=mlflowserverconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MlflowServerConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *MlflowServerConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mlflowServerConfig := mlflowserverv1.MlflowServerConfig{}

	if err := r.Get(ctx, req.NamespacedName, &mlflowServerConfig); err != nil {
		logger.Error(err, "unable to fetch MlflowServerConfig")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	var deployments v1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(req.Namespace), client.MatchingLabels{"app": "mlflow-server"}); err != nil {
		logger.Error(err, "unable to list deployments")
		return reconcile.Result{}, err
	}

	deployment, err := r.createMlServerDeployment(req.Name, req.Namespace, &mlflowServerConfig)

	existingDeployment := &v1.Deployment{}
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
		service, err := r.createMlServerService(req.Name, req.Namespace, &mlflowServerConfig)
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
			for _, model := range r.getLatestModels() {
				logger.Info("Found model", model.Name, model.Version)
				modelDeployment, err := r.createMlModelDeployment(req.Name, req.Namespace, &mlflowServerConfig, model.Name, model.Version)
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

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func (r *MlflowServerConfigReconciler) createMlServerDeployment(name string, namespace string, config *mlflowserverv1.MlflowServerConfig) (*v1.Deployment, error) {
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &config.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Name:            name,
							Image:           config.Spec.Image,
							ImagePullPolicy: v12.PullNever,
							Env: []v12.EnvVar{
								{
									Name:  "AWS_ACCESS_KEY_ID",
									Value: "cAUnfrfQ2uGfeB9Y",
								},
								{
									Name:  "AWS_SECRET_ACCESS_KEY",
									Value: "8vjDVvwlDS07m4dOWw4vwUh9ReFFdCf1",
								},
								{
									Name:  "MLFLOW_S3_ENDPOINT_URL",
									Value: "https://minio.minio.svc.cluster.local",
								},
								{
									Name:  "MLFLOW_S3_IGNORE_TLS",
									Value: "true",
								},
								{
									Name:  "DB_USER",
									Value: "admin",
								},
								{
									Name:  "DB_PASSWORD",
									Value: "123456",
								},
								{
									Name:  "DB_HOST",
									Value: "postgres.postgres",
								},
								{
									Name:  "DB_PORT",
									Value: "5432",
								},
								{
									Name:  "DB_NAME",
									Value: "mlflow",
								},
							},
							Command: []string{"mlflow"},
							Args: []string{
								"server",
								"--serve-artifacts",
								"--artifacts-destination",
								"s3://mlflow",
								"--backend-store-uri",
								"postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)",
								"--host",
								"0.0.0.0",
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, deployment, r.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *MlflowServerConfigReconciler) createMlModelDeployment(name string, namespace string, config *mlflowserverv1.MlflowServerConfig, modelName string, version string) (*v1.Deployment, error) {
	var replicas int32 = 1
	var _name = name + "-model"

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      _name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": _name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": _name,
				},
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": _name}},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Name:            _name,
							Image:           "erayarslan/mlflow_serve:v2.4.1",
							ImagePullPolicy: v12.PullNever,
							Env: []v12.EnvVar{
								{
									Name:  "MLFLOW_TRACKING_URI",
									Value: "http://mlflowserverconfig-sample:5000",
								},
							},
							Command: []string{"mlflow"},
							Args: []string{
								"models",
								"serve",
								"-m",
								fmt.Sprintf("models:/%s/%s", modelName, version),
								"--host",
								"0.0.0.0",
								"--env-manager",
								"conda",
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, deployment, r.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

type LatestVersion struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	CurrentStage string `json:"current_stage"`
	Status       string `json:"status"`
}

type RegisteredModel struct {
	Name           string          `json:"name"`
	LatestVersions []LatestVersion `json:"latest_versions"`
}

type RegisteredModelsResponse struct {
	RegisteredModels []RegisteredModel `json:"registered_models"`
	NextPageToken    *string           `json:"next_page_token,omitempty"`
}

func (r *MlflowServerConfigReconciler) getLatestModels() []LatestVersion {
	var models []LatestVersion
	var nextPageToken *string

	for {
		var response RegisteredModelsResponse
		var err error
		if nextPageToken != nil {
			err = getJson(fmt.Sprintf("http://mlflowserverconfig-sample.mlflow-operator-system:5000/api/2.0/mlflow/registered-models/search?page_token=%s", *nextPageToken), &response)
		} else {
			err = getJson("http://mlflowserverconfig-sample.mlflow-operator-system:5000/api/2.0/mlflow/registered-models/search", &response)
		}

		if err != nil {
			return models
		}

		for _, model := range response.RegisteredModels {
			for _, version := range model.LatestVersions {
				if version.CurrentStage == "Production" {
					models = append(models, version)
				}
			}
		}

		if response.NextPageToken == nil {
			break
		}
	}

	return models
}

func (r *MlflowServerConfigReconciler) createMlServerService(name string, namespace string, config *mlflowserverv1.MlflowServerConfig) (*v12.Service, error) {
	service := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v12.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Type: v12.ServiceTypeClusterIP,
			Ports: []v12.ServicePort{
				{
					Port:     5000,
					Protocol: v12.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5000,
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *MlflowServerConfigReconciler) updateStatus(ctx context.Context, mlflowServerConfig *mlflowserverv1.MlflowServerConfig, obj runtime.Object) error {
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
func (r *MlflowServerConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlflowserverv1.MlflowServerConfig{}).
		Owns(&v1.Deployment{}).
		Owns(&v12.Service{}).
		Complete(r)
}
