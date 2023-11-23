package mlflow

import "strings"

type Model struct {
	Name    string
	Version string
}

func (m Model) ToLowerName() string {
	return strings.ToLower(m.Name)
}

func (m Model) GenerateDeploymentName(prefix string) string {
	return prefix + "-" + m.ToLowerName() + "-" + m.Version
}

type Models []Model
