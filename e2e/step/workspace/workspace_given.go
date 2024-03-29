package workspace

import (
	"context"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"
	workspacesiov1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func givenDefaultWorkspaceIsCreatedForThem(ctx context.Context) (context.Context, error) {
	return defaultWorkspaceIsCreatedForThem(ctx)
}

func givenAPrivateWorkspaceExists(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)

	u, w, err := createUserSignupAndWaitForWorkspace(ctx, cli, ns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	ctx = tcontext.InjectUser(ctx, *u)
	ctx = tcontext.InjectWorkspace(ctx, *w)
	return ctx, nil
}

func givenACommunityWorkspaceExists(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)

	u, err := user.OnboardUser(ctx, cli, ns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	w, err := createWorkspace(ctx, cli, ns, "new-community", u.Status.CompliantUsername, workspacesiov1alpha1.WorkspaceVisibilityCommunity)
	if err != nil {
		return ctx, err
	}

	if err := workspaceIsReadableForEveryone(ctx, cli, ns, w.Name); err != nil {
		return ctx, err
	}

	ctx = tcontext.InjectUser(ctx, *u)
	ctx = tcontext.InjectWorkspace(ctx, *w)
	return ctx, nil
}
