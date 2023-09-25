package mlflow

import "strings"

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
	NextPageToken    *string           `json:"next_page_token,omitempty"`
	RegisteredModels []RegisteredModel `json:"registered_models"`
}

type ModelVersion struct {
	Version string `json:"version"`
}

type ModelVersionsResponse struct {
	NextPageToken *string        `json:"next_page_token,omitempty"`
	ModelVersions []ModelVersion `json:"model_versions"`
}

type Model struct {
	Name    string
	Version string
}

func (m Model) ToLowerName() string {
	return strings.ToLower(m.Name)
}

func (m Model) GenerateDeploymentName(reqName string) string {
	return reqName + "-" + m.ToLowerName() + "-" + m.Version + "-" + "model"
}

type Models []Model
