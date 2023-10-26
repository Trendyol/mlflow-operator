package service

import "k8s.io/apimachinery/pkg/api/resource"

const (
	tagPrefix            = "mlflowOperator-"
	cpuRequestTagName    = tagPrefix + "cpuRequest"
	cpuLimitTagName      = tagPrefix + "cpuLimit"
	memoryRequestTagName = tagPrefix + "memoryRequest"
	memoryLimitTagName   = tagPrefix + "memoryLimit"
)

type OperatorTags struct {
	CPURequest    resource.Quantity
	CPULimit      resource.Quantity
	MemoryRequest resource.Quantity
	MemoryLimit   resource.Quantity
}

// TODO: validate tags to fit kubernetes standards if validation fails update status of ml flow model via ml flow api
// TODO: set default cpu and mem resources if related tag not found

func (t Tags) GetOperatorTags() (mlFlowOperatorTags OperatorTags) {
	for _, tag := range t {
		switch tag.Key {
		case cpuRequestTagName:
			quantity, err := resource.ParseQuantity(tag.Value)
			if err == nil {
				mlFlowOperatorTags.CPURequest = quantity
			}
		case cpuLimitTagName:
			quantity, err := resource.ParseQuantity(tag.Value)
			if err == nil {
				mlFlowOperatorTags.CPULimit = quantity
			}
		case memoryRequestTagName:
			quantity, err := resource.ParseQuantity(tag.Value)
			if err == nil {
				mlFlowOperatorTags.MemoryRequest = quantity
			}
		case memoryLimitTagName:
			quantity, err := resource.ParseQuantity(tag.Value)
			if err == nil {
				mlFlowOperatorTags.MemoryLimit = quantity
			}
		}
	}
	return
}
