package user

import (
	"context"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

func givenAnUserIsOnboarded(ctx context.Context) (context.Context, error) {
	return givenUserIsOnboarded(ctx, DefaultUserName)
}

func givenUserIsOnboarded(ctx context.Context, name string) (context.Context, error) {
	u, err := OnBoardUserInKubespaceNamespace(ctx, name)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectUser(ctx, *u), nil
}
