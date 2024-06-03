package workspace

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
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

func createInternalWorkspace(ctx context.Context, cli cli.Cli, namespace, name, user string, visibility workspacesv1alpha1.InternalWorkspaceVisibility) (*workspacesv1alpha1.InternalWorkspace, error) {
	w := workspacesv1alpha1.InternalWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				workspacesv1alpha1.LabelDisplayName: name,
			},
		},
		Spec: workspacesv1alpha1.InternalWorkspaceSpec{
			Visibility: visibility,
			Owner: workspacesv1alpha1.UserInfo{
				JWTInfo: workspacesv1alpha1.JwtInfo{
					Username: user,
				},
			},
		},
	}

	if err := cli.Create(ctx, &w); err != nil {
		return nil, err
	}
	return &w, nil
}
