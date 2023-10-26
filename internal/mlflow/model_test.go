package mlflow

import (
	"testing"
)

func TestModelToLowerName(t *testing.T) {
	model := Model{Name: "MyModel", Version: "v1.0"}
	expected := "mymodel"
	result := model.ToLowerName()

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestModelGenerateDeploymentName(t *testing.T) {
	model := Model{Name: "MyModel", Version: "v1.0"}
	reqName := "Deployment"
	expected := "Deployment-mymodel-v1.0-model"
	result := model.GenerateDeploymentName(reqName)

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}
