package iwclient

import (
	"context"
	"fmt"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/log"

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
	w, err := c.fetchInternalWorkspaceByLabel(ctx, key.Owner, key.Name, nil)
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
	owner string,
	space string,
	_ ...client.GetOption,
) (*workspacesv1alpha1.InternalWorkspace, error) {
	ww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := c.backend.List(ctx, &ww); err != nil {
		return nil, err
	}
	if len(ww.Items) == 0 {
		return nil, ErrWorkspaceNotFound
	}

	u, err := c.fetchUserSignupByComplaintName(ctx, owner)
	if err != nil {
		return nil, err
	}

	if i := slices.IndexFunc(ww.Items, func(w workspacesv1alpha1.InternalWorkspace) bool {
		return w.Spec.DisplayName == space &&
			w.Status.Owner.Username == u.Status.CompliantUsername
	}); i != -1 {
		return &ww.Items[i], nil
	}

	return nil, ErrWorkspaceNotFound
}

func (c *Client) fetchUserSignupByComplaintName(
	ctx context.Context,
	complaintName string,
) (*toolchainv1alpha1.UserSignup, error) {
	uu := toolchainv1alpha1.UserSignupList{}
	if err := c.backend.List(ctx, &uu); err != nil {
		return nil, err
	}

	i := slices.IndexFunc(uu.Items, func(u toolchainv1alpha1.UserSignup) bool {
		return u.Status.CompliantUsername == complaintName
	})

	switch i {
	case -1:
		return nil, ErrWorkspaceNotFound
	default:
		return &uu.Items[i], nil
	}
}
