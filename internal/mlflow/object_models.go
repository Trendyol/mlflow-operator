package mlflow

import (
	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ModelDeploymentObjectConfig struct {
	Name               string
	Namespace          string
	MlFlowServerConfig *mlflowv1beta1.MLFlow
	Model              Model
	CPURequest         resource.Quantity
	CPULimit           resource.Quantity
	MemoryRequest      resource.Quantity
	MemoryLimit        resource.Quantity
	MlFlowTrackingURI  string
	MlFlowModelImage   string
}
