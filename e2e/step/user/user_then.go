package user

import (
	"context"
	"errors"
	"fmt"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

func thenTheUserRetrievesTheirDefaultWorkspace(ctx context.Context) (context.Context, error) {
	w := tcontext.RetrieveInternalWorkspace(ctx)
	u := tcontext.RetrieveUser(ctx)

	errs := []error{}
	if w.Namespace != u.Name {
		errs = append(errs, fmt.Errorf("expected workspace namespace to be %s, found %s", u.Name, w.Namespace))
	}
	if w.Name != u.Name {
		errs = append(errs, fmt.Errorf("expected workspace name to be %s, found %s", u.Name, w.Namespace))
	}
	return ctx, errors.Join(errs...)
}

func thenTheUserRetrievesAListOfWorkspacesContainingJustTheDefaultOne(ctx context.Context) (context.Context, error) {
	expected := 1
	u := tcontext.RetrieveUser(ctx)
	ww := tcontext.RetrieveUserWorkspaces(ctx)

	if n := len(ww.Items); n != expected {
		return ctx, fmt.Errorf("expected %d workspace, found %d", expected, n)
	}

	if wn := ww.Items[0].Name; wn != u.Name {
		return ctx, fmt.Errorf("expected workspace name to be %s, found %s", u.Name, wn)
	}

	return ctx, nil
}
