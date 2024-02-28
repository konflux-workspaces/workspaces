package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

func thenTheUserRetrievesTheirDefaultWorkspace(ctx context.Context) (context.Context, error) {
	return ctx, godog.ErrPending
}

func thenTheUserRetrievesAListOfWorkspacesContainingJustTheDefaultOne(ctx context.Context) (context.Context, error) {
	u := tcontext.RetrieveUser(ctx)
	ww := tcontext.RetrieveUserWorkspaces(ctx)

	errs := []error{}
	if n := len(ww.Items); n != 1 {
		errs = append(errs, fmt.Errorf("expected 1 workspace, found %d", n))
	}

	if wn := ww.Items[0].Name; wn != u.Name {
		errs = append(errs, fmt.Errorf("expected workspace name to be %s, found %s", u.Name, wn))
	}

	return ctx, errors.Join(errs...)
}
