package mlflow

import mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"

type MlflowModelDeploymentObjectConfig struct {
	Name               string
	Namespace          string
	MlFlowServerConfig *mlflowv1beta1.MLFlow
	Model              Model
	CPURequest         string
	CPULimit           string
	MemoryRequest      string
	MemoryLimit        string
	MlFlowTrackingUri  string
	MlFlowModelImage   string
}
