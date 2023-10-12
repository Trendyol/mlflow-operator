package mlflow

import "strings"

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
