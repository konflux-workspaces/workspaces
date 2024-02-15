package workspace

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func whenUserRequestsANewPrivateWorkspace(ctx context.Context) (context.Context, error) {
	return createNewWorkspace(ctx, "new-private", workspacesv1alpha1.WorkspaceVisibilityPrivate)
}

func whenUserRequestsANewCommunityWorkspace(ctx context.Context) (context.Context, error) {
	return createNewWorkspace(ctx, "new-community", workspacesv1alpha1.WorkspaceVisibilityCommunity)
}

func createNewWorkspace(ctx context.Context, name string, visibility workspacesv1alpha1.WorkspaceVisibility) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveWorkspacesNamespace(ctx)

	w, err := createWorkspace(ctx, cli, ns, name, visibility)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectWorkspace(ctx, *w), nil
}

func whenAWorkspaceIsCreatedForUser(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)

	u, err := user.OnboardUser(ctx, cli, ns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	w, err := getWorkspaceFromWorkspacesNamespace(ctx, u.Status.CompliantUsername)
	if err != nil {
		return ctx, err
	}

	ctx = tcontext.InjectUser(ctx, *u)
	ctx = tcontext.InjectWorkspace(ctx, *w)
	return ctx, nil
}

func whenOwnerChangesVisibilityToCommunity(ctx context.Context) (context.Context, error) {
	return ownerChangesVisibilityTo(ctx, workspacesv1alpha1.WorkspaceVisibilityCommunity)
}

func whenOwnerChangesVisibilityToPrivate(ctx context.Context) (context.Context, error) {
	return ownerChangesVisibilityTo(ctx, workspacesv1alpha1.WorkspaceVisibilityPrivate)
}

func ownerChangesVisibilityTo(ctx context.Context, visibility workspacesv1alpha1.WorkspaceVisibility) (context.Context, error) {
	w := tcontext.RetrieveWorkspace(ctx)
	cli := tcontext.RetrieveHostClient(ctx)

	_, err := controllerutil.CreateOrUpdate(ctx, &cli, &w, func() error {
		if w.Spec.Visibility == visibility {
			return fmt.Errorf("Visibility already set to %v", visibility)
		}
		w.Spec.Visibility = visibility
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tcontext.InjectWorkspace(ctx, w), nil
}

func createWorkspace(ctx context.Context, cli cli.Cli, namespace, name string, visibility workspacesv1alpha1.WorkspaceVisibility) (*workspacesv1alpha1.Workspace, error) {
	w := workspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: workspacesv1alpha1.WorkspaceSpec{
			Visibility: visibility,
		},
	}

	if err := cli.Create(ctx, &w); err != nil {
		return nil, err
	}
	return &w, nil
}
