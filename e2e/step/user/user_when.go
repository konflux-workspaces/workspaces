package user

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	wrest "github.com/konflux-workspaces/workspaces/e2e/pkg/rest"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
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
	ww := workspacesv1alpha1.InternalWorkspaceList{}
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
	w := workspacesv1alpha1.InternalWorkspace{}
	wk := types.NamespacedName{Namespace: u.Name, Name: u.Name}
	if err := c.Get(ctx, wk, &w, &client.GetOptions{}); err != nil {
		k := tcontext.RetrieveUnauthKubeconfig(ctx)
		return ctx, fmt.Errorf("error retrieving workspace %v from host %s as user %s: %w", wk, k.Host, u.Status.CompliantUsername, err)
	}
	log.Printf("retrieved workspace: %v", w)
	return tcontext.InjectInternalWorkspace(ctx, w), nil
}

func whenTheUserChangesWorkspaceVisibilityTo(ctx context.Context, visibility string) (context.Context, error) {
	w := tcontext.RetrieveInternalWorkspace(ctx)

	cli, err := wrest.BuildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}

	w.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibility(visibility)
	if err := cli.Update(ctx, &w, &client.UpdateOptions{}); err != nil {
		return ctx, err
	}
	return tcontext.InjectInternalWorkspace(ctx, w), nil
}
