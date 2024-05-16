package workspace

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

// WorkspaceCreator is the interface the data source needs to implement to allow the CreateWorkspaceHandler to properly create the workspace
type WorkspaceCreator interface {
	CreateUserWorkspace(ctx context.Context, user string, workspace *restworkspacesv1alpha1.Workspace, opts ...client.CreateOption) error
}

// WorkspaceUpdater is the interface the data source needs to implement to allow the UpdateWorkspaceHandler to update the workspace
type WorkspaceUpdater interface {
	UpdateUserWorkspace(ctx context.Context, user string, obj *restworkspacesv1alpha1.Workspace, opts ...client.UpdateOption) error
}
