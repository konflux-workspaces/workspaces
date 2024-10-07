package user

import (
	"context"
	"errors"
	"fmt"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func thenTheUserRetrievesTheirDefaultWorkspace(ctx context.Context) (context.Context, error) {
	w := tcontext.RetrieveUserWorkspace(ctx)
	u := tcontext.RetrieveUser(ctx)

	errs := []error{}
	if w.Namespace != u.Status.CompliantUsername {
		errs = append(errs, fmt.Errorf("expected workspace namespace to be %s, found %s", u.Status.CompliantUsername, w.Namespace))
	}
	if en := workspacesv1alpha1.DisplayNameDefaultWorkspace; w.Name != en {
		errs = append(errs, fmt.Errorf("expected workspace name to be %s, found %s", en, w.Name))
	}
	return ctx, errors.Join(errs...)
}

func thenTheUserRetrievesAListOfWorkspacesContainingJustTheDefaultOne(ctx context.Context) (context.Context, error) {
	ww := tcontext.RetrieveUserWorkspaces(ctx)

	if ew, n := 1, len(ww.Items); n != ew {
		return ctx, fmt.Errorf("expected %d workspace, found %d: %v", ew, n, ww)
	}
	if en, wn := workspacesv1alpha1.DisplayNameDefaultWorkspace, ww.Items[0].Name; wn != en {
		return ctx, fmt.Errorf("expected workspace name to be %s, found %s", en, wn)
	}
	return ctx, nil
}
