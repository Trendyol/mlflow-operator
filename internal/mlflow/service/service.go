package service

import (
	"context"
	"fmt"
	mlflowv1beta1 "github.com/Trendyol/mlflow-operator/api/v1beta1"
	"github.com/Trendyol/mlflow-operator/internal/mlflow"
	"github.com/Trendyol/mlflow-operator/internal/util"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultNamespace = "default"
)

type Client struct {
	httpClient util.HTTPClient
	BaseURL    string
}

func NewClient(mlflowServerCfg *mlflowv1beta1.MLFlow, httpClient util.HTTPClient, debug bool) *Client {
	client := &Client{
		httpClient: httpClient,
	}

	if mlflowServerCfg.Namespace == defaultNamespace {
		client.BaseURL = fmt.Sprintf("http://%s:5000/api/2.0/mlflow", mlflowServerCfg.Name)
	} else {
		client.BaseURL = fmt.Sprintf("http://%s.%s:5000/api/2.0/mlflow", mlflowServerCfg.Name, mlflowServerCfg.Namespace)
	}

	if debug {
		client.BaseURL = "http://localhost:30099/api/2.0/mlflow"
	}

	return client
}

func (m *Client) GetLatestModels() (mlflow.Models, error) {
	var models mlflow.Models
	var nextPageToken *string

	for {
		var response RegisteredModelsResponse
		var err error
		if nextPageToken != nil {
			err = m.httpClient.SendGetRequest(fmt.Sprintf("%s/registered-models/search?page_token=%s", m.BaseURL, *nextPageToken), &response)
		} else {
			err = m.httpClient.SendGetRequest(fmt.Sprintf("%s/registered-models/search", m.BaseURL), &response)
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
				models = append(models, mlflow.Model{
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

func (m *Client) UpdateDescription(name string, message string) error {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	req := map[string]interface{}{
		"name":        name,
		"description": message,
	}

	var r UpdateDescriptionResponse
	err := m.httpClient.SendPatchRequest(fmt.Sprintf("%s/registered-models/update", m.BaseURL), req, &r)
	if err != nil {
		logger.V(1).Error(err, "unable to update description")
		return err
	}

	logger.V(1).Info("Model description updated for model", "Name", r.RegisteredModel.Name)

	return nil
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

		err := m.httpClient.SendGetRequest(fmt.Sprintf("%s/model-versions/search?%s", m.BaseURL, queryParams.Encode()), &response)
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

func (m *Client) GetModelVersionDetail(name, version string) (*ModelVersionDetailResponse, error) {
	var response ModelVersionDetailResponse
	queryParams := url.Values{}
	queryParams.Add("name", name)
	queryParams.Add("version", version)
	err := m.httpClient.SendGetRequest(fmt.Sprintf("%s/model-versions/get?%s", m.BaseURL, queryParams.Encode()), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
