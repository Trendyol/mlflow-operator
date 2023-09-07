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
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image of the MLFlow server
	Image string `json:"image,omitempty"`

	// Quantity of instances
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Name of the ConfigMap for MLFlowSpec's configuration
	// +kubebuilder:validation:MinLength=1
	ConfigMapName string `json:"configMapName"`
}

// MLFlowStatus defines the observed state of MLFlow
type MLFlowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Active is the active instance of the MLflow server deployment
	// +optional
	Active corev1.ObjectReference `json:"active,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MLFlow is the Schema for the mlflows API
type MLFlow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MLFlowSpec   `json:"spec,omitempty"`
	Status MLFlowStatus `json:"status,omitempty"`
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
