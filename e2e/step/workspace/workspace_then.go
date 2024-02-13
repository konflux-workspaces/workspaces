package workspace

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	cli := tcontext.RetrieveHostClient(ctx)
	w := tcontext.RetrieveWorkspace(ctx)

	sbt := types.NamespacedName{Name: fmt.Sprintf("%s-community", w.Name), Namespace: w.Namespace}
	sb := toolchainv1alpha1.SpaceBinding{}
	if err := cli.Get(ctx, sbt, &sb); err != nil {
		return err
	}

	switch {
	case sb.Spec.MasterUserRecord != "public-viewer":
		return fmt.Errorf("expected SpaceBinding %s to have MUR %s, found %s", sb.Name, "public-viewer", sb.Spec.MasterUserRecord)
	case sb.Spec.SpaceRole != "viewer":
		return fmt.Errorf("expected SpaceBinding %s to have SpaceRole %s, found %s", sb.Name, "viewer", sb.Spec.SpaceRole)
	default:
		return nil
	}
}

func thenTheWorkspaceIsReadableForEveryone(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)
	w := tcontext.RetrieveWorkspace(ctx)

	return workspaceIsReadableForEveryone(ctx, cli, ns, w.Name)
}

func thenACommunityWorkspaceIsCreated(ctx context.Context) error {
	w := tcontext.RetrieveWorkspace(ctx)
	return checkWorkspaceVisibility(ctx, w.Name, workspacesv1alpha1.WorkspaceVisibilityCommunity)
}

func thenAPrivateWorkspaceIsCreated(ctx context.Context) error {
	w := tcontext.RetrieveWorkspace(ctx)
	return checkWorkspaceVisibility(ctx, w.Name, workspacesv1alpha1.WorkspaceVisibilityPrivate)
}

func thenTheOwnerIsGrantedAdminAccessToTheWorkspace(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	w := tcontext.RetrieveWorkspace(ctx)
	u := tcontext.RetrieveUser(ctx)
	ns := tcontext.RetrieveTestNamespace(ctx)

	return wait.PollUntilContextTimeout(ctx, 1*time.Second, 1*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		sbb := toolchainv1alpha1.SpaceBindingList{}
		if err := cli.List(ctx, &sbb, client.InNamespace(ns)); err != nil {
			return false, client.IgnoreNotFound(err)
		}

		if len(sbb.Items) == 0 {
			return false, nil
		}

		for _, sb := range sbb.Items {
			if sb.Spec.MasterUserRecord == u.Name && sb.Spec.Space == w.Name && sb.Spec.SpaceRole == "admin" {
				return true, nil
			}
		}
		return false, nil
	})
}

func thenTheWorkspaceVisibilityIsSetTo(ctx context.Context, visibility string) error {
	w := tcontext.RetrieveWorkspace(ctx)

	if w.Spec.Visibility != workspacesv1alpha1.WorkspaceVisibility(visibility) {
		return fmt.Errorf(`expected visibility "%s", found "%s"`, visibility, w.Spec.Visibility)
	}
	return nil
}

func workspaceIsReadableForEveryone(ctx context.Context, cli client.Client, namespace, name string) error {
	sb := &toolchainv1alpha1.SpaceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-community", name),
			Namespace: namespace,
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
	case "viewer" != sb.Spec.SpaceRole:
		return fmt.Errorf("expected SpaceBinding %s to have SpaceRole %s, found %s", sb.Name, "viewer", sb.Spec.SpaceRole)
	case name != sb.Spec.Space:
		return fmt.Errorf("expected referenced space to be %v, found %v", name, sb.Spec.Space)
	case "public-viewer" != sb.Spec.MasterUserRecord:
		return fmt.Errorf("expected referenced MUR to be %v, found %v", "public-viewer", sb.Spec.MasterUserRecord)
	default:
		return nil
	}
}
