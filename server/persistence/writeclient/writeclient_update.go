package writeclient

import (
	"context"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"

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
	key := iwclient.SpaceKey{Owner: workspace.Namespace, Name: workspace.Name}
	if err := c.workspacesReader.GetAsUser(ctx, user, key, &ciw); err != nil {
		return kerrors.NewNotFound(
			restworkspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(),
			workspace.Name)
	}

	// check Generation matching
	if iw.Generation != ciw.Generation {
		return kerrors.NewResourceExpired("workspace version changed")
	}

	// update the InternalWorkspace
	ciw.Spec.Visibility = iw.Spec.Visibility
	log.FromContext(ctx).Debug("updating user workspace", "workspace", iw, "user", user)
	return cli.Update(ctx, &ciw, opts...)
}
