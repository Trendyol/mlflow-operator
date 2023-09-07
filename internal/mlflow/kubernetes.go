package mlflow

import (
	"fmt"
	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Kubernetes struct {
	Scheme *runtime.Scheme
}

func (k *Kubernetes) CreateMlflowModelDeployment(name string, namespace string, config *mlflowv1beta1.MLFlow, modelName string, version string) (*appsv1.Deployment, error) {
	var replicas int32 = 1
	var _name = name + "-model"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      _name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": _name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": _name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": _name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            _name,
							Image:           "erayarslan/mlflow_serve:v2.4.1",
							ImagePullPolicy: corev1.PullNever,
							Env: []corev1.EnvVar{
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

	if err := controllerutil.SetControllerReference(config, deployment, k.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (k *Kubernetes) CreateMlflowService(name string, namespace string, config *mlflowv1beta1.MLFlow) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:     5000,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5000,
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, service, k.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (k *Kubernetes) CreateMlflowDeployment(name string, namespace string, config *mlflowv1beta1.MLFlow) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &config.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           config.Spec.Image,
							ImagePullPolicy: corev1.PullNever,
							Env: []corev1.EnvVar{
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

	if err := controllerutil.SetControllerReference(config, deployment, k.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}
