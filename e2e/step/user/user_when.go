package user

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainapiv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	wrest "github.com/konflux-workspaces/workspaces/e2e/pkg/rest"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

func whenAnUserOnboards(ctx context.Context) (context.Context, error) {
	u, err := OnBoardUserInKubespaceNamespace(ctx, DefaultUserName)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectUser(ctx, *u), nil
}

func whenCustomUserRequestsTheListOfWorkspaces(ctx context.Context, user string) (context.Context, error) {
	u := tcontext.RetrieveCustomUser(ctx, user)
	return userSignupRequestsTheListOfWorkspaces(ctx, u)
}

func whenUserRequestsTheListOfWorkspaces(ctx context.Context) (context.Context, error) {
	u := tcontext.RetrieveUser(ctx)
	return userSignupRequestsTheListOfWorkspaces(ctx, u)
}

func userSignupRequestsTheListOfWorkspaces(ctx context.Context, user toolchainapiv1alpha1.UserSignup) (context.Context, error) {
	c, err := wrest.BuildWorkspacesClientForUser(ctx, user)
	if err != nil {
		return ctx, err
	}

	ww := restworkspacesv1alpha1.WorkspaceList{}
	if err := c.List(ctx, &ww, &client.ListOptions{}); err != nil {
		u := tcontext.RetrieveUser(ctx)
		k := tcontext.RetrieveUnauthKubeconfig(ctx)
		return ctx, fmt.Errorf("error retrieving workspaces from host %s as user %s: %w", k.Host, u.Status.CompliantUsername, err)
	}

	return tcontext.InjectUserWorkspaces(ctx, ww), nil
}

func whenCustomUserRequestsTheirDefaultWorkspace(ctx context.Context, name string) (context.Context, error) {
	u := tcontext.RetrieveCustomUser(ctx, name)
	return userSignupRequestsTheirDefaultWorkspace(ctx, u)
}

func whenUserRequestsTheirDefaultWorkspace(ctx context.Context) (context.Context, error) {
	u := tcontext.RetrieveUser(ctx)
	return userSignupRequestsTheirDefaultWorkspace(ctx, u)
}

func userSignupRequestsTheirDefaultWorkspace(ctx context.Context, user toolchainapiv1alpha1.UserSignup) (context.Context, error) {
	c, err := wrest.BuildWorkspacesClientForUser(ctx, user)
	if err != nil {
		return ctx, err
	}

	w := restworkspacesv1alpha1.Workspace{}
	wk := types.NamespacedName{Namespace: user.Status.CompliantUsername, Name: workspacesv1alpha1.DisplayNameDefaultWorkspace}
	if err := c.Get(ctx, wk, &w, &client.GetOptions{}); err != nil {
		k := tcontext.RetrieveUnauthKubeconfig(ctx)
		return ctx, fmt.Errorf("error retrieving workspace %v from host %s as user %s: %w", wk, k.Host, user.Status.CompliantUsername, err)
	}
	return tcontext.InjectUserWorkspace(ctx, w), nil
}

func whenTheUserPatchesWorkspaceVisibilityTo(ctx context.Context, visibility string) (context.Context, error) {
	cli, err := wrest.BuildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}

	// retrieve user's Workspace from context
	w, err := func() (*restworkspacesv1alpha1.Workspace, error) {
		w, ok := tcontext.LookupUserWorkspace(ctx)
		if !ok {
			// fallback to InternalWorkspace
			iw := tcontext.RetrieveInternalWorkspace(ctx)
			return mapper.Default.InternalWorkspaceToWorkspace(&iw)
		}
		return &w, nil
	}()
	if err != nil {
		return ctx, err
	}

	pw := w.DeepCopy()
	pw.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibility(visibility)
	if err := cli.Patch(ctx, pw, client.MergeFrom(w)); err != nil {
		return ctx, err
	}
	return tcontext.InjectUserWorkspace(ctx, *pw), nil
}

func whenTheUserChangesWorkspaceVisibilityTo(ctx context.Context, visibility string) (context.Context, error) {
	cli, err := wrest.BuildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}

	// retrieve user's Workspace from context
	w, err := func() (*restworkspacesv1alpha1.Workspace, error) {
		w, ok := tcontext.LookupUserWorkspace(ctx)
		if !ok {
			// fallback to InternalWorkspace
			iw := tcontext.RetrieveInternalWorkspace(ctx)
			return mapper.Default.InternalWorkspaceToWorkspace(&iw)
		}
		return &w, nil
	}()
	if err != nil {
		return ctx, err
	}

	w.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibility(visibility)
	if err := cli.Update(ctx, w, &client.UpdateOptions{}); err != nil {
		return ctx, err
	}
	return tcontext.InjectUserWorkspace(ctx, *w), nil
}
