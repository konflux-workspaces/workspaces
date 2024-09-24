package readclient

import (
	"context"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/clientinterface"
	"github.com/konflux-workspaces/workspaces/server/persistence/mutate"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ workspace.WorkspaceReader = &ReadClient{}

// ReadUserWorkspace returns the Workspace details only if the user has access to it
func (c *ReadClient) ReadUserWorkspace(
	ctx context.Context,
	user string,
	owner string,
	space string,
	obj *restworkspacesv1alpha1.Workspace,
	_ ...client.GetOption,
) error {
	l := log.FromContext(ctx).With("user", user, "owner", owner, "space", space)
	var w workspacesv1alpha1.InternalWorkspace
	key := clientinterface.SpaceKey{Owner: owner, Name: space}
	if err := c.internalClient.GetAsUser(ctx, user, key, &w); err != nil {
		l.Error("error retrieving Workspace", "error", err)
		return kerrors.NewNotFound(restworkspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), space)
	}

	r, err := c.mapper.InternalWorkspaceToWorkspace(&w)
	if err != nil {
		l.Error("error mapping internal workspace to workspace as user", "error", err)
		return kerrors.NewInternalError(err)
	}

	// apply is-owner label
	mutate.ApplyIsOwnerLabel(r, user)

	// apply has-direct-access label
	err = mutate.ApplyHasDirectAccessLabel(ctx, c.internalClient, r, user)
	if err != nil {
		l.Error("error checking user access to workspace", "error", err)
		return kerrors.NewInternalError(err)
	}

	// return workspace
	r.DeepCopyInto(obj)
	return nil
}
