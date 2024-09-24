package writeclient

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
	"github.com/konflux-workspaces/workspaces/server/persistence/mutate"

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

	// apply the is-owner label
	mutate.ApplyIsOwnerLabel(w, user)
	// if a user is creating a workspace, then they must have direct access to it
	w.Labels[restworkspacesv1alpha1.LabelHasDirectAccess] = "true"

	// fill return value
	w.DeepCopyInto(workspace)
	return nil
}
