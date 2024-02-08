package workspace

import (
	"context"

	"github.com/cucumber/godog"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"
)

func givenAPrivateWorkspaceExists(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveTestNamespace(ctx)

	u, w, err := createUserAndWaitForWorkspace(ctx, cli, ns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	ctx = tcontext.InjectUser(ctx, *u)
	ctx = tcontext.InjectWorkspace(ctx, *w)
	return ctx, nil
}

func givenACommunityWorkspaceExists(ctx context.Context) error {
	return godog.ErrPending
}
