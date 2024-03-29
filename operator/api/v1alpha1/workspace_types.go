/*
Copyright 2024 The Workspaces Authors.

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

type WorkspaceVisibility string

const (
	WorkspaceVisibilityCommunity WorkspaceVisibility = "community"
	WorkspaceVisibilityPrivate   WorkspaceVisibility = "private"
)

const (
	LabelHomeWorkspace  string = "workspaces.io/home-workspace"
	LabelWorkspaceOwner string = "workspaces.io/owner"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Owner struct {
	// +required
	// Name string `json:"name"`

	// +required
	Id string `json:"id"`
}

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	// +required
	Visibility WorkspaceVisibility `json:"visibility"`
	// +required
	Owner Owner `json:"owner"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	Space string `json:"space"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Visibility",type="string",JSONPath=`.spec.visibility`

// Workspace is the Schema for the workspaces API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
