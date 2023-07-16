/*
Copyright 2022.

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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScalyrAlertFileSpec defines the desired state of ScalyrAlertFile
type ScalyrAlertFileSpec struct {
	// +kubebuilder:validation:Required
	// +required
	AlertAddress string                `json:"alertAddress,omitempty"`
	Alerts       []ScalyrAlertRuleSpec `json:"alerts,omitempty"`
}

// ScalyrAlertFileStatus defines the observed state of ScalyrAlertFile
type ScalyrAlertFileStatus struct {
	Synchronised bool `json:"Synchronised,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ScalyrAlertFile is the Schema for the scalyralertfiles API
type ScalyrAlertFile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScalyrAlertFileSpec   `json:"spec,omitempty"`
	Status ScalyrAlertFileStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScalyrAlertFileList contains a list of ScalyrAlertFile
type ScalyrAlertFileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalyrAlertFile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalyrAlertFile{}, &ScalyrAlertFileList{})
}
