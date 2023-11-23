/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MLFlowSpec defines the desired state of MLFlow
type MLFlowSpec struct {
	Image                    string `json:"image,omitempty"`
	ModelImage               string `json:"modelImage,omitempty"`
	ConfigMapName            string `json:"configMapName"`
	ModelSyncPeriodInMinutes int    `json:"modelSyncPeriodInMinutes,omitempty"`
	Replicas                 int32  `json:"replicas,omitempty"`
}

// MLFlowStatus defines the observed state of MLFlow
type MLFlowStatus struct {
	ActiveModels map[string]corev1.ObjectReference `json:"activeModels,omitempty"`
	Active       corev1.ObjectReference            `json:"active,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MLFlow is the Schema for the mlflows API
type MLFlow struct {
	Status            MLFlowStatus `json:"status,omitempty"`
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MLFlowSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// MLFlowList contains a list of MLFlow
type MLFlowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MLFlow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MLFlow{}, &MLFlowList{})
}
