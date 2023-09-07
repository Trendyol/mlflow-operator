package mlflow

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
