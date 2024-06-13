package workspace

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

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
	ctx = tcontext.InjectInternalWorkspace(ctx, *w)
	return ctx, nil
}

func whenOwnerChangesVisibilityToCommunity(ctx context.Context) (context.Context, error) {
	return ownerChangesVisibilityTo(ctx, workspacesv1alpha1.InternalWorkspaceVisibilityCommunity)
}

func whenOwnerChangesVisibilityToPrivate(ctx context.Context) (context.Context, error) {
	return ownerChangesVisibilityTo(ctx, workspacesv1alpha1.InternalWorkspaceVisibilityPrivate)
}

func ownerChangesVisibilityTo(ctx context.Context, visibility workspacesv1alpha1.InternalWorkspaceVisibility) (context.Context, error) {
	w := tcontext.RetrieveInternalWorkspace(ctx)
	cli := tcontext.RetrieveHostClient(ctx)

	_, err := controllerutil.CreateOrUpdate(ctx, &cli, &w, func() error {
		if w.Spec.Visibility == visibility {
			return fmt.Errorf("Visibility already set to %v", visibility)
		}
		w.Spec.Visibility = visibility
		return nil
	})
	if err != nil {
		return ctx, err
	}

	return tcontext.InjectInternalWorkspace(ctx, w), nil
}
