package service

const (
	tagPrefix            = "mlflowOperator-"
	cpuRequestTagName    = tagPrefix + "cpuRequest"
	cpuLimitTagName      = tagPrefix + "cpuLimit"
	memoryRequestTagName = tagPrefix + "memoryRequest"
	memoryLimitTagName   = tagPrefix + "memoryLimit"
)

type OperatorTags struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

// TODO: validate tags to fit kubernetes standards if validation fails update status of ml flow model via ml flow api

func (t Tags) GetOperatorTags() (mlFlowOperatorTags OperatorTags) {
	for _, tag := range t {
		switch tag.Key {
		case cpuRequestTagName:
			mlFlowOperatorTags.CPURequest = tag.Value
		case cpuLimitTagName:
			mlFlowOperatorTags.CPULimit = tag.Value
		case memoryRequestTagName:
			mlFlowOperatorTags.MemoryRequest = tag.Value
		case memoryLimitTagName:
			mlFlowOperatorTags.MemoryLimit = tag.Value
		}
	}
	return
}
