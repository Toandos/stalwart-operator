/*
Copyright 2026.

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
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// JMAPObjectSpec defines the desired state of JMAPObject
type JMAPObjectSpec struct {
	// The stalwart cluster the jmap object is applied to
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="cluster reference is immutable after creation"
	// +required
	ClusterRef corev1.LocalObjectReference `json:"clusterRef"`

	// The type of the jmap object
	// +required
	Type string `json:"type"`

	// The data of the jmap object
	// +required
	Data apiextensionsv1.JSON `json:"data"`
}

// JMAPObjectStatus defines the observed state of JMAPObject.
type JMAPObjectStatus struct {
	// The internal id of the created jmap object
	// +optional
	StalwartID string `json:"stalwartID,omitempty"`

	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// JMAPObject is the Schema for the jmapobjects API
type JMAPObject struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of JMAPObject
	// +required
	Spec JMAPObjectSpec `json:"spec"`

	// status defines the observed state of JMAPObject
	// +optional
	Status JMAPObjectStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// JMAPObjectList contains a list of JMAPObject
type JMAPObjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []JMAPObject `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &JMAPObject{}, &JMAPObjectList{})
		return nil
	})
}
