package iwclient

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

var (
	ErrWorkspaceNotFound error = fmt.Errorf("workspace not found")
	ErrUnauthorized      error = fmt.Errorf("user is not authorized to read the workspace")
	ErrMoreThanOneFound  error = fmt.Errorf("more than one workspace found")
)

// GetAsUser retrieves the requested workspace if and only if it is community or `user` is allowed access to
func (c *Client) GetAsUser(
	ctx context.Context,
	user string,
	key SpaceKey,
	workspace *workspacesv1alpha1.InternalWorkspace,
	opts ...client.GetOption,
) error {
	l := log.FromContext(ctx).With("key", key, "user", user)
	l.Debug("retrieving InternalWorkspace")
	w, err := c.fetchInternalWorkspace(ctx, key.Owner, key.Name, nil)
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

func (c *Client) fetchInternalWorkspace(
	ctx context.Context,
	owner string,
	space string,
	_ ...client.GetOption,
) (*workspacesv1alpha1.InternalWorkspace, error) {
	u, err := c.fetchUserSignupByComplaintName(ctx, owner)
	if err != nil {
		return nil, err
	}

	ww := workspacesv1alpha1.InternalWorkspaceList{}
	opts := []client.ListOption{
		client.MatchingFields{
			cache.IndexKeyInternalWorkspaceDisplayName:   space,
			cache.IndexKeyInternalWorkspaceOwnerUsername: u.Status.CompliantUsername,
		},
	}
	if err := c.backend.List(ctx, &ww, opts...); err != nil {
		return nil, err
	}
	if len(ww.Items) == 0 {
		return nil, ErrWorkspaceNotFound
	}

	return &ww.Items[0], nil
}

func (c *Client) fetchUserSignupByComplaintName(
	ctx context.Context,
	complaintName string,
) (*toolchainv1alpha1.UserSignup, error) {
	uu := toolchainv1alpha1.UserSignupList{}
	opt := client.MatchingFields{cache.IndexKeyUserComplaintName: complaintName}
	if err := c.backend.List(ctx, &uu, opt); err != nil {
		return nil, err
	}

	if len(uu.Items) == 0 {
		return nil, ErrWorkspaceNotFound
	}
	return &uu.Items[0], nil
}
