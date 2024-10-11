package user

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

func whenUserRequestsTheListOfWorkspaces(ctx context.Context) (context.Context, error) {
	c, err := wrest.BuildWorkspacesClient(ctx)
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

func whenUserRequestsTheirDefaultWorkspace(ctx context.Context) (context.Context, error) {
	c, err := wrest.BuildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}

	u := tcontext.RetrieveUser(ctx)
	w := restworkspacesv1alpha1.Workspace{}
	wk := types.NamespacedName{Namespace: u.Status.CompliantUsername, Name: workspacesv1alpha1.DisplayNameDefaultWorkspace}
	if err := c.Get(ctx, wk, &w, &client.GetOptions{}); err != nil {
		k := tcontext.RetrieveUnauthKubeconfig(ctx)
		return ctx, fmt.Errorf("error retrieving workspace %v from host %s as user %s: %w", wk, k.Host, u.Status.CompliantUsername, err)
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
	return tcontext.InjectUserWorkspace(ctx, *w), nil
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
