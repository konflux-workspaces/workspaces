package writeclient

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/core/workspace/v2"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ workspace.WorkspaceCreator = &WriteClient{}

// CreateUserWorkspace creates as `user` the InternalWorkspace representing the provided Workspace
func (c *WriteClient) CreateUserWorkspace(ctx context.Context, user string, workspace *restworkspacesv1alpha1.Workspace, opts ...client.CreateOption) error {
	cli, err := c.buildClient(user)
	if err != nil {
		return err
	}

	// map Workspace to InternalWorkspace
	iw, err := mapper.Default.WorkspaceToInternalWorkspace(workspace)
	if err != nil {
		return err
	}
	iw.SetNamespace(c.workspacesNamespace)
	iw.SetName("")
	iw.SetGenerateName(workspace.Name)

	// create InternalWorkspace
	log.FromContext(ctx).Debug("creating user workspace", "workspace", workspace, "user", user)
	if err := cli.Create(ctx, iw, opts...); err != nil {
		return err
	}

	// map InternalWorkspace to Workspace
	w, err := mapper.Default.InternalWorkspaceToWorkspace(iw)
	if err != nil {
		return err
	}

	// fill return value
	w.DeepCopyInto(workspace)
	return nil
}
