package iwclient

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/log"
)

var (
	ErrWorkspaceNotFound error = fmt.Errorf("workspace not found")
	ErrUnauthorized      error = fmt.Errorf("user is not authorized to read the workspace")
	ErrMoreThanOneFound  error = fmt.Errorf("more than one workspace found")
)

func (c *Client) GetAsUser(
	ctx context.Context,
	user string,
	key SpaceKey,
	workspace *workspacesv1alpha1.InternalWorkspace,
	opts ...client.GetOption,
) error {
	l := log.FromContext(ctx).With("key", key, "user", user)
	l.Debug("retrieving InternalWorkspace")
	w, err := c.fetchInternalWorkspaceByLabel(ctx, user, key.Owner, key.Name, nil)
	if err != nil {
		l.Error("error retrieving InternalWorkspace", "error", err)
		return err
	}

	// if workspace visibility is community all users are allowed visibility
	if w.Spec.Visibility == workspacesv1alpha1.InternalWorkspaceVisibilityCommunity {
		l.Debug("InternalWorkspace has community visibility, returning it")
		w.DeepCopyInto(workspace)
		return nil
	}

	// check if user has direct visibility on the space
	l.Debug("InternalWorkspace is private, checking for a SpaceBinding for the user")
	ok, err := c.existsSpaceBindingForUserAndSpace(ctx, user, w.GetName())
	if err != nil {
		l.Error("error retrieving SpaceBindings for InternalWorkspace", "error", err)
		return err
	}
	if !ok {
		return ErrUnauthorized
	}

	w.DeepCopyInto(workspace)
	return nil
}

func (c *Client) fetchInternalWorkspaceByLabel(
	ctx context.Context,
	user string,
	owner string,
	space string,
	_ ...client.GetOption,
) (*workspacesv1alpha1.InternalWorkspace, error) {
	ww := workspacesv1alpha1.InternalWorkspaceList{}
	opts := []client.ListOption{
		client.MatchingLabels{
			workspacesv1alpha1.LabelDisplayName:    space,
			workspacesv1alpha1.LabelWorkspaceOwner: owner,
		},
	}
	if err := c.backend.List(ctx, &ww, opts...); err != nil {
		return nil, err
	}

	switch ni := len(ww.Items); ni {
	case 0:
		return nil, ErrWorkspaceNotFound
	case 1:
		return &ww.Items[0], nil
	default:
		return nil, ErrMoreThanOneFound
	}
}
