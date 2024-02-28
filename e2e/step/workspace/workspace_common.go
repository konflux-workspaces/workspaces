package workspace

import (
	"context"
	"fmt"
	"time"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func defaultWorkspaceIsCreatedForThem(ctx context.Context) (context.Context, error) {
	u := tcontext.RetrieveUser(ctx)
	w, err := getWorkspaceFromWorkspacesNamespace(ctx, u.Status.CompliantUsername)
	if err != nil {
		return ctx, err
	}

	return tcontext.InjectWorkspace(ctx, *w), nil
}

func createUserSignupAndWaitForWorkspace(
	ctx context.Context,
	cli cli.Cli,
	namespace, name string,
) (*toolchainv1alpha1.UserSignup, *workspacesv1alpha1.Workspace, error) {
	u, err := user.OnboardUser(ctx, cli, namespace, user.DefaultUserName)
	if err != nil {
		return nil, nil, err
	}

	w, err := getWorkspaceFromWorkspacesNamespace(ctx, u.Status.CompliantUsername)
	if err != nil {
		return nil, nil, err
	}

	return u, w, nil
}

func getWorkspaceFromWorkspacesNamespace(ctx context.Context, name string) (*workspacesv1alpha1.Workspace, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveWorkspacesNamespace(ctx)

	return getWorkspace(ctx, cli, ns, name)
}

func getWorkspace(ctx context.Context, cli cli.Cli, ns, name string) (*workspacesv1alpha1.Workspace, error) {
	w := workspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	if err := wait.PollUntilContextTimeout(ctx, time.Second*5, time.Minute, true, func(ctx context.Context) (done bool, err error) {
		if err := cli.Get(ctx, client.ObjectKeyFromObject(&w), &w); err != nil {
			return false, client.IgnoreNotFound(err)
		}
		return true, nil
	}); err != nil {
		return nil, fmt.Errorf("error retrieving workspace %s/%s: :%w", w.Namespace, w.Name, err)
	}
	return &w, nil
}

func checkWorkspaceVisibility(ctx context.Context, name string, visibility workspacesv1alpha1.WorkspaceVisibility) error {
	w, err := getWorkspaceFromWorkspacesNamespace(ctx, name)
	if err != nil {
		return err
	}

	if w.Spec.Visibility != visibility {
		return fmt.Errorf("expected %v visibility, found %v", visibility, w.Spec.Visibility)
	}
	return nil
}
