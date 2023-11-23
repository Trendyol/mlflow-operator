package service

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTags_GetOperatorTags(t *testing.T) {
	tests := []struct {
		wantMlFlowOperatorTags OperatorTags
		name                   string
		t                      Tags
	}{
		{
			name: "should return all resource tags",
			t: []ModelVersionTag{
				{
					Key:   "mlflowOperator-cpuRequest",
					Value: "100m",
				},
				{
					Key:   "mlflowOperator-cpuLimit",
					Value: "100m",
				},
				{
					Key:   "mlflowOperator-memoryRequest",
					Value: "1G",
				},
				{
					Key:   "mlflowOperator-memoryLimit",
					Value: "1G",
				},
			},
			wantMlFlowOperatorTags: OperatorTags{
				CPURequest:    resource.MustParse("100m"),
				CPULimit:      resource.MustParse("100m"),
				MemoryRequest: resource.MustParse("1G"),
				MemoryLimit:   resource.MustParse("1G"),
			},
		},
		{
			name: "should return empty resources if all values are wrong",
			t: []ModelVersionTag{
				{
					Key:   "mlflowOperator-cpuRequest",
					Value: "hundred",
				},
				{
					Key:   "mlflowOperator-cpuLimit",
					Value: "hundred",
				},
				{
					Key:   "mlflowOperator-memoryRequest",
					Value: "one",
				},
				{
					Key:   "mlflowOperator-memoryLimit",
					Value: "one",
				},
			},
			wantMlFlowOperatorTags: OperatorTags{},
		},
		{
			name: "should return empty resources if all keys are wrong",
			t: []ModelVersionTag{
				{
					Key:   "mlflowOperator-cpuRequest1",
					Value: "100m",
				},
				{
					Key:   "mlflowOperator-cpuLimit1",
					Value: "100m",
				},
				{
					Key:   "mlflowOperator-memoryRequest1",
					Value: "1G",
				},
				{
					Key:   "mlflowOperator-memoryLimit1",
					Value: "1G",
				},
			},
			wantMlFlowOperatorTags: OperatorTags{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMlFlowOperatorTags := tt.t.GetOperatorTags(); !reflect.DeepEqual(gotMlFlowOperatorTags, tt.wantMlFlowOperatorTags) {
				t.Errorf("GetOperatorTags() = %v, want %v", gotMlFlowOperatorTags, tt.wantMlFlowOperatorTags)
			}
		})
	}
}
