package service

const (
	prefixMLFlowTag      = "mlflowOperator-"
	cpuRequestTagName    = prefixMLFlowTag + "cpuRequest"
	cpuLimitTagName      = prefixMLFlowTag + "cpuLimit"
	memoryRequestTagName = prefixMLFlowTag + "memoryRequest"
	memoryLimitTagName   = prefixMLFlowTag + "memoryLimit"
)

type MlFlowOperatorTags struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

// TODO: validate tags to fit kubernetes standards if validation fails update status of ml flow model via ml flow api

func (t Tags) GetMlFlowOperatorTags() (mlFlowOperatorTags MlFlowOperatorTags) {
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
