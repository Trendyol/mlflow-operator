package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	"github.com/Trendyol/mlflow-operator/internal/mlflow"
	"github.com/Trendyol/mlflow-operator/internal/mlflow/service"
	"github.com/Trendyol/mlflow-operator/internal/util"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	K8sClient           client.Client
	Scheme              *runtime.Scheme
	HTTPClient          util.HTTPClient
	MlflowClient        *service.Client
	MlflowObjectManager *mlflow.ObjectManager
	Debug               bool
	ModelSyncOnce       sync.Once
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

	existingDeployment, err := r.CreateOrUpdateDeployment(ctx, deployment)
	if err != nil {
		logger.Error(err, "unable to create Deployment for MlflowServerConfig")
		return reconcile.Result{}, err
	}

	r.UpdateStatus(ctx, &mlflowServerConfig, deployment, func(mlflowServerConfig *mlflowv1beta1.MLFlow, ref *corev1.ObjectReference) {
		if mlflowServerConfig.Status.ActiveModels == nil {
			mlflowServerConfig.Status.ActiveModels = make(map[string]corev1.ObjectReference)
		}
		mlflowServerConfig.Status.Active = *ref
	})

	svc, err := r.MlflowObjectManager.CreateMlflowServiceObject(req.Name, req.Namespace, &mlflowServerConfig)
	if err != nil {
		logger.Error(err, "unable to create Client for MlflowServerConfig when creating service")
		return reconcile.Result{}, err
	}

	_, err = r.CreateOrUpdateService(ctx, svc)
	if err != nil {
		logger.Error(err, "unable to create Client for MlflowServerConfig when pushing to k8s")
		return reconcile.Result{}, err
	}

	if r.DeploymentIsNotReady(existingDeployment) {
		logger.Info("Waiting for Deployment to be ready", deployment.Namespace, deployment.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	logger.Info("Deployment is ready")
	r.ModelSyncOnce.Do(func() {
		r.InitializeMlflowClient(&mlflowServerConfig)
		go r.StartMlFlowModelSync(deployment.Namespace, &mlflowServerConfig)
	})

	if r.Debug {
		if err := r.createTestModel(ctx, req, mlflowServerConfig); err != nil {
			logger.Error(err, "unable to create Job for MlflowServerConfig when creating job for test model")
			return reconcile.Result{}, err
		}
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

	r.MlflowClient = service.NewClient(mlflowServerCfg, r.HTTPClient, r.Debug)
}

func (r *MLFlowReconciler) CreateMLFlowWineQualityJob(ctx context.Context, job *batchv1.Job) error {
	if err := r.K8sClient.Create(ctx, job); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MLFlowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.ModelSyncOnce = sync.Once{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&mlflowv1beta1.MLFlow{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *MLFlowReconciler) MlFlowModelSync(namespace string, mlflowServerConfig *mlflowv1beta1.MLFlow) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	models, getModelsErr := r.MlflowClient.GetLatestModels()
	if getModelsErr != nil {
		logger.Error(getModelsErr, "unable to get latest models")
		return
	}

	for _, model := range models {
		modelDetails, modelDetailsErr := r.MlflowClient.GetModelVersionDetail(model.Name, model.Version)
		if modelDetailsErr != nil {
			logger.Error(modelDetailsErr, "failed to get model details")
			continue
		}

		mlFlowOperatorTags := modelDetails.Tags.GetOperatorTags()
		modelDeployment, modelDeploymentErr := r.MlflowObjectManager.CreateMlflowModelDeploymentObject(mlflow.ModelDeploymentObjectConfig{
			Name:               strings.ToLower(model.Name),
			Namespace:          namespace,
			MlFlowServerConfig: mlflowServerConfig,
			Model:              model,
			CPURequest:         mlFlowOperatorTags.CPURequest,
			CPULimit:           mlFlowOperatorTags.CPULimit,
			MemoryRequest:      mlFlowOperatorTags.MemoryRequest,
			MemoryLimit:        mlFlowOperatorTags.MemoryLimit,
			MlFlowTrackingURI:  fmt.Sprintf("http://%s:5000", mlflowServerConfig.Name),
			MlFlowModelImage:   mlflowServerConfig.Spec.ModelImage,
		})
		if modelDeploymentErr != nil {
			logger.Error(modelDeploymentErr, "unable to create Deployment for Model when creating model deployment")
		}

		existingDeployment, err := r.CreateOrUpdateDeployment(ctx, modelDeployment)
		if err != nil {
			logger.Error(err, "unable to create Deployment for Model when pushing to k8s")
			r.updateDescription(model.Name, "Your Mlflow deployment has been failed to deploy")
			return
		}

		r.UpdateStatus(ctx, mlflowServerConfig, existingDeployment, func(mlflowServerConfig *mlflowv1beta1.MLFlow, ref *corev1.ObjectReference) {
			if mlflowServerConfig.Status.ActiveModels == nil {
				mlflowServerConfig.Status.ActiveModels = make(map[string]corev1.ObjectReference)
			}
			mlflowServerConfig.Status.ActiveModels[existingDeployment.Name] = *ref
		})
		r.updateDescription(model.Name, "Your Mlflow deployment has been deployed")
	}
}

func (r *MLFlowReconciler) UpdateStatus(
	ctx context.Context,
	mlflowServerConfig *mlflowv1beta1.MLFlow,
	obj runtime.Object,
	callback func(mlflowServerConfig *mlflowv1beta1.MLFlow, ref *corev1.ObjectReference),
) {
	ref, err := reference.GetReference(r.Scheme, obj)
	if err != nil {
		return
	}

	callback(mlflowServerConfig, ref)

	if err = r.K8sClient.Status().Update(ctx, mlflowServerConfig); err != nil {
		return
	}
}

func (r *MLFlowReconciler) StartMlFlowModelSync(namespace string, mlflowServerConfig *mlflowv1beta1.MLFlow) {
	t := time.NewTicker(time.Minute * time.Duration(mlflowServerConfig.Spec.ModelSyncPeriodInMinutes))
	r.MlFlowModelSync(namespace, mlflowServerConfig)
	for range t.C {
		r.MlFlowModelSync(namespace, mlflowServerConfig)
	}
}

func (r *MLFlowReconciler) updateDescription(name string, message string) {
	updateTime := time.Now().Format("15:04:05 2006-01-02")
	msg := fmt.Sprintf("%s at %s", message, updateTime)
	err := r.MlflowClient.UpdateDescription(name, msg)
	if err != nil {
		return
	}
}
