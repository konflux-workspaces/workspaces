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

type InternalWorkspaceVisibility string

const (
	DisplayNameDefaultWorkspace string = "default"

	InternalWorkspaceVisibilityCommunity InternalWorkspaceVisibility = "community"
	InternalWorkspaceVisibilityPrivate   InternalWorkspaceVisibility = "private"

	LabelInternalDomain string = "internal.workspaces.konflux.io/"
	LabelHomeWorkspace  string = LabelInternalDomain + "home-workspace"
	LabelDisplayName    string = LabelInternalDomain + "display-name"

	PublicViewerName string = "kubesaw-authenticated"
)

// UserInfo collects information for a logged user
type UserInfo struct {
	//+required
	JWTInfo JwtInfo `json:"jwtInfo"`
}

// JwtInfo collects information from the User JSON Web Token
type JwtInfo struct {
	//+required
	Sub string `json:"sub"`

	//+required
	Username string `json:"username"`

	//+optional
	Email string `json:"email"`
}

// InternalWorkspaceSpec defines the desired state of Workspace
type InternalWorkspaceSpec struct {
	//+required
	Visibility InternalWorkspaceVisibility `json:"visibility"`

	//+required
	Owner UserInfo `json:"owner"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	Space string `json:"space"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Visibility",type="string",JSONPath=`.spec.visibility`

// InternalWorkspace is the Schema for the workspaces API
type InternalWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InternalWorkspaceSpec `json:"spec,omitempty"`
	Status WorkspaceStatus       `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InternalWorkspaceList contains a list of Workspace
type InternalWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InternalWorkspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InternalWorkspace{}, &InternalWorkspaceList{})
}
