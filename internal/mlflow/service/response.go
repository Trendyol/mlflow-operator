package service

type RegisteredModelsResponse struct {
	NextPageToken    *string           `json:"next_page_token,omitempty"`
	RegisteredModels []RegisteredModel `json:"registered_models"`
}

type RegisteredModel struct {
	Name           string          `json:"name"`
	LatestVersions []LatestVersion `json:"latest_versions"`
}

type LatestVersion struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	CurrentStage string `json:"current_stage"`
	Status       string `json:"status"`
}

type ModelVersionsResponse struct {
	NextPageToken *string        `json:"next_page_token,omitempty"`
	ModelVersions []ModelVersion `json:"model_versions"`
}

type ModelVersion struct {
	Version string `json:"version"`
}

type ModelVersionDetailResponse struct {
	ModelVersionDetail `json:"model_version"`
}

type ModelVersionDetail struct {
	Name                 string `json:"name"`
	Version              string `json:"version"`
	CreationTimestamp    int64  `json:"creation_timestamp"`
	LastUpdatedTimestamp int64  `json:"last_updated_timestamp"`
	CurrentStage         string `json:"current_stage"`
	Description          string `json:"description"`
	Source               string `json:"source"`
	RunId                string `json:"run_id"`
	Status               string `json:"status"`
	Tags                 Tags   `json:"tags"`
	RunLink              string `json:"run_link"`
}

type Tags []ModelVersionTag

type ModelVersionTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
