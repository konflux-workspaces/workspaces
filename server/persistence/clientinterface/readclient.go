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

package clientinterface

import (
	"context"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SpaceKey comprises a Space name, with a mandatory owner.
type SpaceKey struct {
	Owner string
	Name  string
}

// InternalWorkspacesReader is the definition for a InternalWorkspaces Read Client
type InternalWorkspacesReader interface {
	GetAsUser(context.Context, string, SpaceKey, *workspacesv1alpha1.InternalWorkspace, ...client.GetOption) error
	ListAsUser(context.Context, string, *workspacesv1alpha1.InternalWorkspaceList) error
}

// InternalWorkspacesMapper is the definition for a InternalWorkspaces/Workspaces Mapper
type InternalWorkspacesMapper interface {
	InternalWorkspaceListToWorkspaceList(*workspacesv1alpha1.InternalWorkspaceList) (*restworkspacesv1alpha1.WorkspaceList, error)
	InternalWorkspaceToWorkspace(*workspacesv1alpha1.InternalWorkspace) (*restworkspacesv1alpha1.Workspace, error)
	WorkspaceToInternalWorkspace(*restworkspacesv1alpha1.Workspace) (*workspacesv1alpha1.InternalWorkspace, error)
}

type DirectAccessChecker interface {
	UserHasDirectAccess(context.Context, string, string) (bool, error)
}

type InternalWorkspacesReadClient interface {
	InternalWorkspacesReader
	DirectAccessChecker
}
