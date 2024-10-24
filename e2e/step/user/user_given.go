package user

import (
	"context"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

func givenAnUserIsOnboarded(ctx context.Context) (context.Context, error) {
	u, err := OnBoardUserInKubespaceNamespace(ctx, DefaultUserName)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectUser(ctx, *u), nil
}

func givenUserIsOnboarded(ctx context.Context, name string) (context.Context, error) {
	u, err := OnBoardUserInKubespaceNamespace(ctx, name)
	if err != nil {
		return ctx, err
	}
	return tcontext.InjectCustomUser(ctx, name, *u), nil
}
