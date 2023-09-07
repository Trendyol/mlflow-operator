package mlflow

import (
	"fmt"
	"github.com/Trendyol/mlflow-operator/internal/util"
	"net/url"
)

type Service struct {
	baseUrl    string
	httpClient *util.HttpClient
}

func (m *Service) getModelVersions(name string) ([]ModelVersion, error) {
	var versions []ModelVersion
	var nextPageToken *string

	for {
		var response ModelVersionsResponse

		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("name='%s'", name))

		if nextPageToken != nil {
			queryParams.Add("page_token", *nextPageToken)
		}

		err := m.httpClient.GetJSON(fmt.Sprintf("%s/model-versions/search?%s", m.baseUrl, queryParams.Encode()), &response)

		if err != nil {
			return nil, err
		}

		for _, version := range response.ModelVersions {
			versions = append(versions, version)
		}

		if response.NextPageToken == nil {
			break
		}
	}

	return versions, nil
}

func (m *Service) GetLatestModels() ([]Model, error) {
	var models []Model
	var nextPageToken *string

	for {
		var response RegisteredModelsResponse
		var err error
		if nextPageToken != nil {
			err = m.httpClient.GetJSON(fmt.Sprintf("%s/registered-models/search?page_token=%s", m.baseUrl, *nextPageToken), &response)
		} else {
			err = m.httpClient.GetJSON(fmt.Sprintf("%s/registered-models/search", m.baseUrl), &response)
		}

		if err != nil {
			return nil, err
		}

		for _, model := range response.RegisteredModels {
			versions, err := m.getModelVersions(model.Name)
			if err != nil {
				return nil, err
			}

			for _, version := range versions {
				models = append(models, Model{
					Name:    model.Name,
					Version: version.Version,
				})
			}
		}

		if response.NextPageToken == nil {
			break
		}
	}

	return models, nil
}

func NewService(service string, namespace string, httpClient *util.HttpClient, debug bool) *Service {
	var baseUrl string

	if debug {
		baseUrl = "http://localhost:30099/api/2.0/mlflow"
	} else {
		if namespace != "default" {
			baseUrl = fmt.Sprintf("http://%s.%s:5000/api/2.0/mlflow", service, namespace)
		} else {
			baseUrl = fmt.Sprintf("http://%s:5000/api/2.0/mlflow", service)
		}
	}

	return &Service{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}
