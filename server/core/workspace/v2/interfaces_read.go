package workspace

import (
	"context"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkspaceLister is the interface the data source needs to implement to allow the ListWorkspaceHandler to fetch data from it
type WorkspaceLister interface {
	ListUserWorkspaces(ctx context.Context, user string, objs *restworkspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error
}

// WorkspaceReader is the interface the data source needs to implement to allow the ReadWorkspaceHandler to fetch data from it
type WorkspaceReader interface {
	ReadUserWorkspace(ctx context.Context, user, owner, space string, obj *restworkspacesv1alpha1.Workspace, opts ...client.GetOption) error
}
