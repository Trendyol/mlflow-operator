package mlflow

import (
	"fmt"
	"github.com/Trendyol/mlflow-operator/internal/util"
)

type Service struct {
	baseUrl    string
	httpClient *util.HttpClient
}

func (m *Service) GetLatestModelVersion() ([]LatestVersion, error) {
	var models []LatestVersion
	var nextPageToken *string

	for {
		var response RegisteredModelsResponse
		var err error
		if nextPageToken != nil {
			err = m.httpClient.GetJSON(fmt.Sprintf("%s?page_token=%s", m.baseUrl, *nextPageToken), &response)
		} else {
			err = m.httpClient.GetJSON(m.baseUrl, &response)
		}

		if err != nil {
			return nil, err
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

	return models, nil
}

func NewService(service string, namespace string, httpClient *util.HttpClient) *Service {
	var baseUrl = fmt.Sprintf("http://%s.%s:5000/api/2.0/mlflow/registered-models/search", service, namespace)

	return &Service{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}
