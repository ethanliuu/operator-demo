/*
Copyright 2022 ethanliu.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperatorDemoSpec defines the desired state of OperatorDemo
type OperatorDemoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
	// omitempty 非必选
	Port int32 `json:"port,omitempty"`
}

// OperatorDemoStatus defines the observed state of OperatorDemo
type OperatorDemoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Replicas int32 `json:"replicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OperatorDemo is the Schema for the operatordemoes API
type OperatorDemo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorDemoSpec   `json:"spec,omitempty"`
	Status OperatorDemoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OperatorDemoList contains a list of OperatorDemo
type OperatorDemoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatorDemo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OperatorDemo{}, &OperatorDemoList{})
}
