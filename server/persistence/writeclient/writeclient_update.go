package writeclient

import (
	"context"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/clientinterface"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
	"github.com/konflux-workspaces/workspaces/server/persistence/mutate"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ workspace.WorkspaceUpdater = &WriteClient{}

// UpdateUserWorkspace updates as `user` the InternalWorkspace representing the provided Workspace
func (c *WriteClient) UpdateUserWorkspace(ctx context.Context, user string, workspace *restworkspacesv1alpha1.Workspace, opts ...client.UpdateOption) error {
	// build client impersonating the user
	cli, err := c.buildClient(user)
	if err != nil {
		return err
	}

	// map to InternalWorkspace
	iw, err := mapper.Default.WorkspaceToInternalWorkspace(workspace)
	if err != nil {
		return kerrors.NewBadRequest("malformed workspace")
	}

	// get the InternalWorkspace as user
	ciw := workspacesv1alpha1.InternalWorkspace{}
	key := clientinterface.SpaceKey{Owner: workspace.Namespace, Name: workspace.Name}
	if err := c.workspacesReader.GetAsUser(ctx, user, key, &ciw); err != nil {
		return kerrors.NewNotFound(
			restworkspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(),
			workspace.Name)
	}

	// check Generation matching
	if iw.Generation != ciw.Generation {
		return kerrors.NewResourceExpired("workspace version changed")
	}
	if ciw.Status.Owner.Username != user {
		return kerrors.NewForbidden(
			workspacesv1alpha1.GroupVersion.WithResource("workspace").GroupResource(),
			"to update a workspace you need to be the owner", nil)
	}

	// update the InternalWorkspace
	ciw.Spec.Visibility = iw.Spec.Visibility
	log.FromContext(ctx).Debug("updating user workspace", "workspace", iw, "user", user)
	err = cli.Update(ctx, &ciw, opts...)
	if err != nil {
		return err
	}

	ws, err := mapper.Default.InternalWorkspaceToWorkspace(&ciw)
	if err != nil {
		return kerrors.NewInternalError(err)
	}

	mutate.ApplyIsOwnerLabel(ws, user)
	// If a user is updating a workspace, they have direct access
	// to the workspace.
	ws.Labels[restworkspacesv1alpha1.LabelHasDirectAccess] = "true"

	ws.DeepCopyInto(workspace)
	return nil
}
