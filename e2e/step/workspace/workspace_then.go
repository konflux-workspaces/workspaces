package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func thenDefaultWorkspaceIsCreatedForThem(ctx context.Context) (context.Context, error) {
	u := tcontext.RetrieveUser(ctx)
	w, err := getWorkspaceFromTestNamespace(ctx, u.Name)
	if err != nil {
		return ctx, err
	}

	return tcontext.InjectWorkspace(ctx, *w), nil
}

func thenTheWorkspaceIsReadableOnlyForGranted(ctx context.Context) error {
	return godog.ErrPending
}

func thenTheWorkspaceIsReadableForEveryone(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveTestNamespace(ctx)
	w := tcontext.RetrieveWorkspace(ctx)

	sb := &toolchainv1alpha1.SpaceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-community", w.Name),
			Namespace: ns,
		},
	}
	if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
		if err := cli.Get(ctx, client.ObjectKeyFromObject(sb), sb); err != nil {
			return false, client.IgnoreNotFound(err)
		}
		return true, nil
	}); err != nil {
		return err
	}

	switch {
	case w.Name != sb.Spec.Space:
		return fmt.Errorf("expected referenced space to be %v, found %v", w.Name, sb.Spec.Space)
	case "public-viewer" != sb.Spec.MasterUserRecord:
		return fmt.Errorf("expected referenced MUR to be %v, found %v", "public-viewer", sb.Spec.MasterUserRecord)
	default:
		return nil
	}
}

func thenACommunityWorkspaceIsCreated(ctx context.Context) error {
	return checkWorkspaceVisibility(ctx, "new-community", workspacesv1alpha1.WorkspaceVisibilityCommunity)
}

func thenAPrivateWorkspaceIsCreated(ctx context.Context) error {
	return checkWorkspaceVisibility(ctx, "new-private", workspacesv1alpha1.WorkspaceVisibilityPrivate)
}

func thenAnUserOnboards() error {
	return godog.ErrPending
}

func thenTheOwnerIsGrantedAdminAccessToTheWorkspace() error {
	return godog.ErrPending
}

func thenTheWorkspaceVisibilityIsSetTo(ctx context.Context, visibility string) error {
	w := tcontext.RetrieveWorkspace(ctx)

	if w.Spec.Visibility != workspacesv1alpha1.WorkspaceVisibility(visibility) {
		return fmt.Errorf(`expected visibility "%s", found "%s"`, visibility, w.Spec.Visibility)
	}
	return nil
}
