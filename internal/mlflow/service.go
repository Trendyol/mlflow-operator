package mlflow

import (
	"fmt"
	"net/url"

	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	"github.com/Trendyol/mlflow-operator/internal/util"
)

type Client struct {
	httpClient util.HTTPClient
	BaseURL    string
}

func NewClient(mlflowServerCfg *mlflowv1beta1.MLFlow, httpClient util.HTTPClient, debug bool) *Client {
	client := &Client{
		httpClient: httpClient,
	}

	if mlflowServerCfg.Namespace == "default" {
		client.BaseURL = fmt.Sprintf("http://%s:5000/api/2.0/mlflow", mlflowServerCfg.Name)
	} else {
		client.BaseURL = fmt.Sprintf("http://%s.%s:5000/api/2.0/mlflow", mlflowServerCfg.Name, mlflowServerCfg.Namespace)
	}

	if debug {
		client.BaseURL = "http://localhost:30099/api/2.0/mlflow"
	}

	return client
}

func (m *Client) GetLatestModels() (Models, error) {
	var models Models
	var nextPageToken *string

	for {
		var response RegisteredModelsResponse
		var err error
		if nextPageToken != nil {
			err = m.httpClient.GetJSON(fmt.Sprintf("%s/registered-models/search?page_token=%s", m.BaseURL, *nextPageToken), &response)
		} else {
			err = m.httpClient.GetJSON(fmt.Sprintf("%s/registered-models/search", m.BaseURL), &response)
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

func (m *Client) getModelVersions(name string) ([]ModelVersion, error) {
	var versions []ModelVersion
	var nextPageToken *string

	for {
		var response ModelVersionsResponse

		queryParams := url.Values{}
		queryParams.Add("filter", fmt.Sprintf("name='%s'", name))

		if nextPageToken != nil {
			queryParams.Add("page_token", *nextPageToken)
		}

		err := m.httpClient.GetJSON(fmt.Sprintf("%s/model-versions/search?%s", m.BaseURL, queryParams.Encode()), &response)
		if err != nil {
			return nil, err
		}

		versions = append(versions, response.ModelVersions...)

		if response.NextPageToken == nil {
			break
		}
	}

	return versions, nil
}
