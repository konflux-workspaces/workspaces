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
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkspaceVisibility string

const (
	// WorkspaceVisibilityCommunity Community value for Workspaces visibility
	WorkspaceVisibilityCommunity WorkspaceVisibility = "community"
	// WorkspaceVisibilityPrivate Private value for Workspaces visibility
	WorkspaceVisibilityPrivate WorkspaceVisibility = "private"

	// LabelIsOwner if the requesting user is the owner of the workspace
	LabelIsOwner string = workspacesv1alpha1.LabelInternalDomain + "is-owner"
	// LabelHasDirectAccess if the requesting user has access to the workspace
	// via a direct grant or via public-viewer
	LabelHasDirectAccess string = workspacesv1alpha1.LabelInternalDomain + "has-direct-access"
)

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	//+required
	//+kubebuilder:validation:Enum:=community;private
	Visibility WorkspaceVisibility `json:"visibility"`
}

// SpaceInfo Information about a Space
type SpaceInfo struct {
	//+required
	Name string `json:"name"`

	// TargetCluster contains the URL to the cluster where the workspace's namespaces live
	//+optional
	TargetCluster string `json:"targetCluster,omitempty"`
}

// UserInfoStatus User info stored in the status
type UserInfoStatus struct {
	//+required
	Email string `json:"email"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	//+optional
	Space *SpaceInfo `json:"space,omitempty"`
	//+optional
	Owner *UserInfoStatus `json:"owner,omitempty"`
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Visibility",type="string",JSONPath=`.spec.visibility`

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
