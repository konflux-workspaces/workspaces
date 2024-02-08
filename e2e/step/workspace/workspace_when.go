package workspace

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func whenUserRequestsANewPrivateWorkspace(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveTestNamespace(ctx)

	w := workspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-private",
			Namespace: ns,
		},
		Spec: workspacesv1alpha1.WorkspaceSpec{
			Visibility: workspacesv1alpha1.WorkspaceVisibilityPrivate,
		},
	}
	return cli.Create(ctx, &w)
}

func whenUserRequestsANewCommunityWorkspace(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveTestNamespace(ctx)

	w := workspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-community",
			Namespace: ns,
		},
		Spec: workspacesv1alpha1.WorkspaceSpec{
			Visibility: workspacesv1alpha1.WorkspaceVisibilityCommunity,
		},
	}
	return cli.Create(ctx, &w)
}

func whenAWorkspaceIsCreatedForUser(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveTestNamespace(ctx)

	_, err := user.OnboardUser(ctx, cli, ns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	if _, err := getWorkspaceFromTestNamespace(ctx, user.DefaultUserName); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func whenOwnerChangesVisibilityToCommunity(ctx context.Context) (context.Context, error) {
	return ownerChangesVisibilityTo(ctx, workspacesv1alpha1.WorkspaceVisibilityCommunity)
}

func whenOwnerChangesVisibilityToPrivate(ctx context.Context) (context.Context, error) {
	return ownerChangesVisibilityTo(ctx, workspacesv1alpha1.WorkspaceVisibilityPrivate)
}

func ownerChangesVisibilityTo(ctx context.Context, visibility workspacesv1alpha1.WorkspaceVisibility) (context.Context, error) {
	w := tcontext.RetrieveWorkspace(ctx)
	cli := tcontext.RetrieveHostClient(ctx)

	_, err := controllerutil.CreateOrUpdate(ctx, cli, &w, func() error {
		if w.Spec.Visibility == visibility {
			return fmt.Errorf("Visibility already set to %v", visibility)
		}
		w.Spec.Visibility = visibility
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tcontext.InjectWorkspace(ctx, w), nil
}
