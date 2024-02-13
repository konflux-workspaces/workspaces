package user

import (
	"context"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

func whenAnUserOnboards(ctx context.Context) (context.Context, error) {
	u, err := OnBoardUserInKubespaceNamespace(ctx, DefaultUserName)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectUser(ctx, *u), nil
}
