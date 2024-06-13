package workspace

import (
	"context"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
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
	ctx = tcontext.InjectInternalWorkspace(ctx, *w)
	return ctx, nil
}

func givenACommunityWorkspaceExists(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	kns := tcontext.RetrieveKubespaceNamespace(ctx)

	u, w, err := createUserSignupAndWaitForWorkspace(ctx, cli, kns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	ctx = tcontext.InjectUser(ctx, *u)
	ctx = tcontext.InjectInternalWorkspace(ctx, *w)

	w.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityCommunity
	if err := cli.Update(ctx, w); err != nil {
		return ctx, err
	}

	if err := workspaceIsReadableForEveryone(ctx, cli, kns, w.Name); err != nil {
		return ctx, err
	}

	return ctx, nil
}
