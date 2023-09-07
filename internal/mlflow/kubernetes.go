package mlflow

import (
	"fmt"
	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

type Kubernetes struct {
	Scheme *runtime.Scheme
	Debug  bool
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
							Image:           "erayarslan/mlflow_serve:v2.6.0-conda",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "MLFLOW_TRACKING_URI",
									Value: "http://mlflow-sample:5000",
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
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Port:     5000,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5000,
					},
					NodePort: 30099,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, service, k.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (k *Kubernetes) CreateMlflowDeployment(name string, namespace string, config *mlflowv1beta1.MLFlow, volumes []string) (*appsv1.Deployment, error) {
	var env []corev1.EnvVar
	var args = []string{
		"server",
		"--serve-artifacts",
		"--host",
		"0.0.0.0",
	}

	var volumeList []corev1.Volume
	var volumeMountList []corev1.VolumeMount

	if !k.Debug {
		env = []corev1.EnvVar{
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
		}
		args = append(args,
			"--artifacts-destination",
			"s3://mlflow",
			"--backend-store-uri",
			"postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)",
		)
	}

	if k.Debug {
		for _, volume := range volumes {
			volumeList = append(volumeList, corev1.Volume{
				Name: volume,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: volume,
					},
				},
			})
		}
		for _, volume := range volumes {
			volumeMountList = append(volumeMountList, corev1.VolumeMount{
				Name:      volume,
				MountPath: fmt.Sprintf("/%s", strings.Split(volume, "-")[2]),
			})
		}
	}

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
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             env,
							Command:         []string{"mlflow"},
							Args:            args,
							VolumeMounts:    volumeMountList,
						},
					},
					Volumes: volumeList,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, deployment, k.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (k *Kubernetes) CreateMlflowPersistence(name string, namespace string, folder string, config *mlflowv1beta1.MLFlow) (*corev1.PersistentVolumeClaim, error) {
	storageClassName := "hostpath"
	volumeMode := corev1.PersistentVolumeFilesystem

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + folder,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			VolumeMode:       &volumeMode,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": resource.MustParse("1Gi"),
				},
			},
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase:       corev1.ClaimBound,
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse("1Gi"),
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, pvc, k.Scheme); err != nil {
		return nil, err
	}

	return pvc, nil
}

func (k *Kubernetes) CreateMlflowWineQualityJob(name string, namespace string, config *mlflowv1beta1.MLFlow) (*batchv1.Job, error) {
	backoffLimit := int32(4)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           "erayarslan/mlflow_wine_quality_example:v2.6.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "TRACKING_URL",
									Value: "http://mlflow-sample:5000",
								},
							},
						},
					},
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}

	if err := controllerutil.SetControllerReference(config, job, k.Scheme); err != nil {
		return nil, err
	}

	return job, nil
}
