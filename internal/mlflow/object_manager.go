package mlflow

import (
	"fmt"
	"strings"

	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	appLabelKey = "app"
)

type ObjectManager struct {
	Scheme *runtime.Scheme
	Debug  bool
}

func (om *ObjectManager) CreateMlflowModelDeploymentObject(config ModelDeploymentObjectConfig) (*appsv1.Deployment, error) {
	var replicas int32 = 1

	depName := config.Model.GenerateDeploymentName(config.MlFlowServerConfig.Name)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: config.Namespace,
			Labels: map[string]string{
				appLabelKey: depName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					appLabelKey: depName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{appLabelKey: depName}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            depName,
							Image:           config.MlFlowModelImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "MLFLOW_TRACKING_URI",
									Value: config.MlFlowTrackingURI,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    config.CPULimit,
									corev1.ResourceMemory: config.MemoryLimit,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    config.CPURequest,
									corev1.ResourceMemory: config.MemoryRequest,
								},
							},
							Command: []string{"mlflow"},
							Args: []string{
								"models",
								"serve",
								"-m",
								fmt.Sprintf("models:/%s/%s", config.Model.Name, config.Model.Version),
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

	if err := controllerutil.SetControllerReference(config.MlFlowServerConfig, deployment, om.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (om *ObjectManager) CreateMlflowServiceObject(name string, namespace string, config *mlflowv1beta1.MLFlow) (*corev1.Service, error) {
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

	if err := controllerutil.SetControllerReference(config, service, om.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (om *ObjectManager) CreateVolumeObject(volumes []string) []corev1.Volume {
	volumeList := make([]corev1.Volume, 0, len(volumes))

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

	return volumeList
}

func (om *ObjectManager) CreateVolumeMountObject(volumes []string) []corev1.VolumeMount {
	volumeMountList := make([]corev1.VolumeMount, 0, len(volumes))

	for _, volume := range volumes {
		volumeMountList = append(volumeMountList, corev1.VolumeMount{
			Name:      volume,
			MountPath: fmt.Sprintf("/%s", strings.Split(volume, "-")[2]),
		})
	}

	return volumeMountList
}

func (om *ObjectManager) CreateMlflowDeploymentObject(name string, namespace string, config *mlflowv1beta1.MLFlow) (*appsv1.Deployment, error) {
	args := []string{
		"server",
		"--serve-artifacts",
		"--host",
		"0.0.0.0",
		"--artifacts-destination",
		"s3://mlflow",
		"--backend-store-uri",
		"postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)",
	}

	// TODO add this values to config map
	env := []corev1.EnvVar{
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
					ResourceClaims: []corev1.PodResourceClaim{},
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           config.Spec.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             env,
							Command:         []string{"mlflow"},
							Args:            args,
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(config, deployment, om.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (om *ObjectManager) CreateMlflowPVCObject(name string, namespace string, folder string, config *mlflowv1beta1.MLFlow) (*corev1.PersistentVolumeClaim, error) {
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
			Resources: corev1.VolumeResourceRequirements{
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

	if err := controllerutil.SetControllerReference(config, pvc, om.Scheme); err != nil {
		return nil, err
	}

	return pvc, nil
}

func (om *ObjectManager) CreateMlflowWineQualityJobObject(name string, namespace string, config *mlflowv1beta1.MLFlow) (*batchv1.Job, error) {
	var backoffLimit int32 = 4

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
									Value: fmt.Sprintf("http://%s:5000", config.Name),
								},
							},
						},
					},
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}

	if err := controllerutil.SetControllerReference(config, job, om.Scheme); err != nil {
		return nil, err
	}

	return job, nil
}
