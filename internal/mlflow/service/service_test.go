package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Trendyol/mlflow-operator/internal/mlflow"
	"github.com/Trendyol/mlflow-operator/mock"
)

func TestGetLatestModels(t *testing.T) {
	// given
	responses := map[string]string{
		"http://example.com/registered-models/search":                         generateRegisteredModelsResponse(),
		"http://example.com/model-versions/search?filter=name%3D%27ModelA%27": generateModelVersionResponse(),
		"http://example.com/model-versions/search?filter=name%3D%27ModelB%27": generateModelVersionResponse(),
		"http://example.com/model-versions/search?filter=name%3D%27ModelC%27": generateModelVersionResponse(),
	}
	mockClient := &mock.MockHTTPClient{
		Responses: responses,
	}

	client := &Client{
		httpClient: mockClient,
		BaseURL:    "http://example.com",
	}

	// when
	models, err := client.GetLatestModels()
	// then
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if len(models) != 9 {
		t.Errorf("Expected 9 models, but got %d", len(models))
	}

	expectedModels := mlflow.Models{
		{Name: "ModelA", Version: "2"},
		{Name: "ModelB", Version: "1"},
		{Name: "ModelC", Version: "1"},
		{Name: "ModelA", Version: "1"},
		{Name: "ModelA", Version: "2"},
		{Name: "ModelB", Version: "1"},
		{Name: "ModelC", Version: "1"},
		{Name: "ModelA", Version: "1"},
		{Name: "ModelA", Version: "2"},
	}
	if len(models) != len(expectedModels) {
		t.Errorf("Expected %d models, but got %d", len(expectedModels), len(models))
	}

	for _, expectedModel := range expectedModels {
		found := false
		for _, model := range models {
			if model.Name == expectedModel.Name && model.Version == expectedModel.Version {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s version %s not found in the result", expectedModel.Name, expectedModel.Version)
		}
	}
}

func TestUpdateDescription(t *testing.T) {
	// given
	response, _ := json.Marshal(UpdateDescriptionResponse{RegisteredModel{
		Name: "ModelA",
		LatestVersions: []LatestVersion{{
			Name:         "Model1",
			Version:      "1",
			CurrentStage: "Production",
			Status:       "Active",
		}},
	}})
	responses := map[string]string{
		"http://example.com/registered-models/update": string(response),
	}
	mockClient := &mock.MockHTTPClient{
		Responses: responses,
	}

	client := &Client{
		httpClient: mockClient,
		BaseURL:    "http://example.com",
	}

	// when
	err := client.UpdateDescription("ModelA")

	// then
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func generateRegisteredModelsResponse() string {
	latestVersion1 := LatestVersion{
		Name:         "Model1",
		Version:      "1",
		CurrentStage: "Production",
		Status:       "Active",
	}

	latestVersion2 := LatestVersion{
		Name:         "Model1",
		Version:      "2",
		CurrentStage: "Staging",
		Status:       "Inactive",
	}

	registeredModel1 := RegisteredModel{
		Name:           "ModelA",
		LatestVersions: []LatestVersion{latestVersion1, latestVersion2},
	}

	latestVersion3 := LatestVersion{
		Name:         "Model2",
		Version:      "1",
		CurrentStage: "Production",
		Status:       "Active",
	}

	registeredModel2 := RegisteredModel{
		Name:           "ModelB",
		LatestVersions: []LatestVersion{latestVersion3},
	}

	latestVersion4 := LatestVersion{
		Name:         "Model3",
		Version:      "1",
		CurrentStage: "Staging",
		Status:       "Inactive",
	}

	registeredModel3 := RegisteredModel{
		Name:           "ModelC",
		LatestVersions: []LatestVersion{latestVersion4},
	}

	response := RegisteredModelsResponse{
		RegisteredModels: []RegisteredModel{registeredModel1, registeredModel2, registeredModel3},
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return string(jsonResponse)
}

func generateModelVersionResponse() string {
	versions := []ModelVersion{
		{Version: "1"},
		{Version: "2"},
		{Version: "3"},
	}

	response := ModelVersionsResponse{
		NextPageToken: nil,
		ModelVersions: versions,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return string(jsonResponse)
}
