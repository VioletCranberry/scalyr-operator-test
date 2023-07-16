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

// ScalyrAlertRuleSpec defines the desired state of ScalyrAlertRule
type AlertRuleSilenceRule struct {
	SilenceUntil   string `json:"silenceUntil,omitempty"`
	SilenceComment string `json:"silenceComment,omitempty"`
	Condition      string `json:"condition,omitempty"`
}
type ScalyrAlertRuleSpec struct {
	// +kubebuilder:validation:Required
	// +required
	Trigger string `json:"trigger"`
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	AlertAddress string `json:"alertAddress,omitempty"`
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +optional
	GracePeriodMinutes *int32 `json:"gracePeriodMinutes,omitempty"`
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +optional
	RenotifyPeriodMinutes *int32 `json:"renotifyPeriodMinutes,omitempty"`
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=9999
	// +optional
	ResolutionDelayMinutes *int32 `json:"resolutionDelayMinutes,omitempty"`
	// +optional
	Silences []AlertRuleSilenceRule `json:"silences,omitempty"`
}

// ScalyrAlertRuleStatus defines the observed state of ScalyrAlertRule
type ScalyrAlertRuleStatus struct {
	Synchronised bool `json:"Synchronised,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ScalyrAlertRule is the Schema for the scalyralertrules API
type ScalyrAlertRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScalyrAlertRuleSpec   `json:"spec,omitempty"`
	Status ScalyrAlertRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScalyrAlertRuleList contains a list of ScalyrAlertRule
type ScalyrAlertRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalyrAlertRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalyrAlertRule{}, &ScalyrAlertRuleList{})
}
