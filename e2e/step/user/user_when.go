package user

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/pkg/rest"
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
	wk := types.NamespacedName{Namespace: u.Name, Name: workspacesv1alpha1.DisplayNameDefaultWorkspace}
	if err := c.Get(ctx, wk, &w, &client.GetOptions{}); err != nil {
		k := tcontext.RetrieveUnauthKubeconfig(ctx)
		return ctx, fmt.Errorf("error retrieving workspace %v from host %s as user %s: %w", wk, k.Host, u.Status.CompliantUsername, err)
	}
	log.Printf("retrieved workspace: %v", w)
	return tcontext.InjectUserWorkspace(ctx, w), nil
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

func whenUserRequestsANewPrivateWorkspace(ctx context.Context) (context.Context, error) {
	return createNewWorkspace(ctx, "new-private", restworkspacesv1alpha1.WorkspaceVisibilityPrivate)
}

func whenUserRequestsANewCommunityWorkspace(ctx context.Context) (context.Context, error) {
	return createNewWorkspace(ctx, "new-community", restworkspacesv1alpha1.WorkspaceVisibilityCommunity)
}

func createNewWorkspace(ctx context.Context, name string, visibility restworkspacesv1alpha1.WorkspaceVisibility) (context.Context, error) {
	u := tcontext.RetrieveUser(ctx)

	cli, err := rest.BuildWorkspacesClient(ctx)
	if err != nil {
		return ctx, err
	}

	w := restworkspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: u.Status.CompliantUsername,
		},
		Spec: restworkspacesv1alpha1.WorkspaceSpec{
			Visibility: visibility,
		},
	}

	if err := cli.Create(ctx, &w); err != nil {
		return nil, err
	}
	return tcontext.InjectUserWorkspace(ctx, w), nil
}
