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

	// Image of the MLFlow model
	ModelImage string `json:"modelImage,omitempty"`

	// Name of the ConfigMap for MLFlowSpec's configuration
	// +kubebuilder:validation:MinLength=1
	ConfigMapName string `json:"configMapName"`

	// Image of the MLFlow model
	ModelSyncPeriodInMinutes int `json:"modelSyncPeriodInMinutes,omitempty"`

	// Quantity of instances
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`
}

// MLFlowStatus defines the observed state of MLFlow
type MLFlowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ActiveModels is the active instances of the MLflow model deployments
	ActiveModels map[string]corev1.ObjectReference `json:"activeModels,omitempty"`

	// Active is the active instance of the MLflow server deployment
	// +optional
	Active corev1.ObjectReference `json:"active,omitempty"`
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
