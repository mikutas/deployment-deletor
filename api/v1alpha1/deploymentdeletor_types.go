/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type NamespacedName struct {
	//+optional
	Namespace string `json:"namespace"`

	// +optional
	Name string `json:"name"`
}

// DeploymentDeletorSpec defines the desired state of DeploymentDeletor
type DeploymentDeletorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Format=string
	MaxAge string `json:"maxAge"`

	//+optional
	Deployment NamespacedName `json:"deployment"`

	//+kubebuilder:validation:Required
	Selector *metav1.LabelSelector `json:"selector"`
}

// DeploymentDeletorStatus defines the observed state of DeploymentDeletor
type DeploymentDeletorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	LastDeletedDeployment metav1.ObjectMeta `json:"lastDeletedDeployment"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=dd;dds
//+kubebuilder:subresource:status

// DeploymentDeletor is the Schema for the deploymentdeletors API
type DeploymentDeletor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentDeletorSpec   `json:"spec,omitempty"`
	Status DeploymentDeletorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeploymentDeletorList contains a list of DeploymentDeletor
type DeploymentDeletorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeploymentDeletor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeploymentDeletor{}, &DeploymentDeletorList{})
}
