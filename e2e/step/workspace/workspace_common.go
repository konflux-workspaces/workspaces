package workspace

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/pkg/poll"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func defaultWorkspaceIsCreatedForThem(ctx context.Context) (context.Context, error) {
	u := tcontext.RetrieveUser(ctx)
	return defaultWorkspaceIsCreatedForUser(ctx, u.Status.CompliantUsername)
}

func defaultWorkspaceIsCreatedForCustomUser(ctx context.Context, name string) (context.Context, error) {
	u := tcontext.RetrieveCustomUser(ctx, name)
	return defaultWorkspaceIsCreatedForUser(ctx, u.Status.CompliantUsername)
}

func defaultWorkspaceIsCreatedForUser(ctx context.Context, compliantUsername string) (context.Context, error) {
	w, err := getWorkspaceFromWorkspacesNamespace(ctx, compliantUsername)
	if err != nil {
		return ctx, err
	}

	return tcontext.InjectInternalWorkspace(ctx, *w), nil
}

func createUserSignupAndWaitForWorkspace(
	ctx context.Context,
	cli cli.Cli,
	namespace, name string,
) (*toolchainv1alpha1.UserSignup, *workspacesv1alpha1.InternalWorkspace, error) {
	u, err := user.OnboardUser(ctx, cli, namespace, name)
	if err != nil {
		return nil, nil, err
	}

	w, err := getWorkspaceFromWorkspacesNamespace(ctx, u.Status.CompliantUsername)
	if err != nil {
		return nil, nil, err
	}

	return u, w, nil
}

func getWorkspaceFromWorkspacesNamespace(ctx context.Context, name string) (*workspacesv1alpha1.InternalWorkspace, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveWorkspacesNamespace(ctx)

	return getWorkspace(ctx, cli, ns, name)
}

func getWorkspace(ctx context.Context, cli cli.Cli, ns, name string) (*workspacesv1alpha1.InternalWorkspace, error) {
	w := workspacesv1alpha1.InternalWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	if err := poll.WaitForConditionImmediately(ctx, func(ctx context.Context) (done bool, err error) {
		if err := cli.Get(ctx, client.ObjectKeyFromObject(&w), &w); err != nil {
			return false, client.IgnoreNotFound(err)
		}

		return w.Status.Owner.Username != "", nil
	}); err != nil {
		return nil, fmt.Errorf("error retrieving workspace %s/%s: :%w", w.Namespace, w.Name, err)
	}
	return &w, nil
}

func checkWorkspaceVisibility(ctx context.Context, name string, visibility workspacesv1alpha1.InternalWorkspaceVisibility) error {
	w, err := getWorkspaceFromWorkspacesNamespace(ctx, name)
	if err != nil {
		return err
	}

	if w.Spec.Visibility != visibility {
		return fmt.Errorf("expected %v visibility, found %v", visibility, w.Spec.Visibility)
	}
	return nil
}
