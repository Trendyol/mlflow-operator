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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MlflowServerConfigSpec defines the desired state of MlflowServerConfig
type MlflowServerConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image of the MLflow server
	Image string `json:"image,omitempty"`

	// Quantity of instances
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Name of the ConfigMap for MlflowServerConfigSpec's configuration
	// +kubebuilder:validation:MinLength=1
	ConfigMapName string `json:"configMapName"`
}

// MlflowServerConfigStatus defines the observed state of MlflowServerConfig
type MlflowServerConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Active is the active instance of the MLflow server deployment
	// +optional
	Active v1.ObjectReference `json:"active,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MlflowServerConfig is the Schema for the mlflowserverconfigs API
type MlflowServerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MlflowServerConfigSpec   `json:"spec,omitempty"`
	Status MlflowServerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MlflowServerConfigList contains a list of MlflowServerConfig
type MlflowServerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MlflowServerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MlflowServerConfig{}, &MlflowServerConfigList{})
}
