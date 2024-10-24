package workspace

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/pkg/poll"
	wrest "github.com/konflux-workspaces/workspaces/e2e/pkg/rest"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

func thenDefaultWorkspaceHasNoClusterURLInStatus(ctx context.Context) (context.Context, error) {
	return thenDefaultWorkspaceCheckClusterURLInStatusIs(ctx, "")
}

func thenDefaultWorkspaceHasClusterURLInStatus(ctx context.Context) (context.Context, error) {
	return thenDefaultWorkspaceCheckClusterURLInStatusIs(ctx, defaultWorkspaceTargetClusterURL)
}

func thenDefaultWorkspaceCheckClusterURLInStatusIs(ctx context.Context, clusterURL string) (context.Context, error) {
	return ctx, poll.WaitForConditionImmediatelyJoiningErrors(ctx, func(ctx context.Context) (bool, error) {
		u := tcontext.RetrieveUser(ctx)
		cli, err := wrest.BuildWorkspacesClient(ctx)
		if err != nil {
			return false, err
		}

		w := restworkspacesv1alpha1.Workspace{}
		k := types.NamespacedName{
			Name:      workspacesv1alpha1.DisplayNameDefaultWorkspace,
			Namespace: u.Status.CompliantUsername,
		}
		if err := cli.Get(ctx, k, &w); err != nil {
			return false, err
		}

		return w.Status.Space.TargetCluster == clusterURL, nil
	})
}

func thenDefaultWorkspaceIsCreatedForThem(ctx context.Context) (context.Context, error) {
	return defaultWorkspaceIsCreatedForThem(ctx)
}

func thenTheWorkspaceIsReadableOnlyForGranted(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	u := tcontext.RetrieveUser(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)

	asbb := toolchainv1alpha1.SpaceBindingList{}
	if err := cli.Client.List(ctx, &asbb, client.InNamespace(ns), client.MatchingLabels{
		toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: u.Name,
	}); err != nil {
		return err
	}

	sbb := []toolchainv1alpha1.SpaceBinding{}
	for _, sb := range asbb.Items {
		if cli.HasScenarioPrefix(sb.Name) {
			sbb = append(sbb, sb)
		}
	}
	if len(sbb) != 1 {
		return fmt.Errorf("expected just one SpaceBinding, found %d", len(sbb))
	}

	sb := sbb[0]
	switch {
	case sb.Spec.MasterUserRecord != u.Status.CompliantUsername:
		return fmt.Errorf("expected SpaceBinding %s to have MUR %s, found %s", sb.Name, u.Status.CompliantUsername, sb.Spec.MasterUserRecord)
	case sb.Spec.SpaceRole != "admin":
		return fmt.Errorf("expected SpaceBinding %s to have SpaceRole %s, found %s", sb.Name, "admin", sb.Spec.SpaceRole)
	default:
		return nil
	}
}

func thenTheWorkspaceIsReadableForEveryone(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)
	w := tcontext.RetrieveInternalWorkspace(ctx)

	return workspaceIsReadableForEveryone(ctx, cli, ns, w.Name)
}

func thenACommunityWorkspaceIsCreated(ctx context.Context) error {
	w := tcontext.RetrieveInternalWorkspace(ctx)
	return checkWorkspaceVisibility(ctx, w.Name, workspacesv1alpha1.InternalWorkspaceVisibilityCommunity)
}

func thenAPrivateWorkspaceIsCreated(ctx context.Context) error {
	w := tcontext.RetrieveInternalWorkspace(ctx)
	return checkWorkspaceVisibility(ctx, w.Name, workspacesv1alpha1.InternalWorkspaceVisibilityPrivate)
}

func thenTheOwnerIsGrantedAdminAccessToTheWorkspace(ctx context.Context) error {
	cli := tcontext.RetrieveHostClient(ctx)
	w := tcontext.RetrieveInternalWorkspace(ctx)
	u := tcontext.RetrieveUser(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)

	return poll.WaitForConditionImmediately(ctx, func(ctx context.Context) (done bool, err error) {
		asbb := toolchainv1alpha1.SpaceBindingList{}
		if err := cli.Client.List(ctx, &asbb, client.InNamespace(ns)); err != nil {
			return false, client.IgnoreNotFound(err)
		}

		sbb := []toolchainv1alpha1.SpaceBinding{}
		for _, sb := range asbb.Items {
			if cli.HasScenarioPrefix(sb.Name) {
				sbb = append(sbb, sb)
			}
		}
		if len(sbb) == 0 {
			return false, nil
		}

		for _, sb := range sbb {
			if sb.Spec.MasterUserRecord == u.Status.CompliantUsername && sb.Spec.Space == w.Name && sb.Spec.SpaceRole == "admin" {
				return true, nil
			}
		}
		return false, nil
	})
}

func thenUserCanNotChangeVisibilityTo(ctx context.Context, user, visibility string) error {
	iw := tcontext.RetrieveInternalWorkspace(ctx)
	u := tcontext.RetrieveCustomUser(ctx, user)
	cli, err := wrest.BuildWorkspacesClientForUser(ctx, u)
	if err != nil {
		return err
	}

	w := restworkspacesv1alpha1.Workspace{}
	wk := types.NamespacedName{Name: iw.Spec.DisplayName, Namespace: iw.Status.Owner.Username}
	if err := cli.Get(ctx, wk, &w); err != nil {
		return err
	}
	w.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibility(visibility)
	err = cli.Update(ctx, &w)
	switch {
	case err == nil:
		return fmt.Errorf("expected forbidden error updating the workspace, got no error")
	case errors.IsForbidden(err):
		return nil
	default:
		return fmt.Errorf("expected forbidden error updating the workspace, got %v", err)
	}
}

func thenUserCanNotPatchVisibilityTo(ctx context.Context, user, visibility string) error {
	iw := tcontext.RetrieveInternalWorkspace(ctx)
	u := tcontext.RetrieveCustomUser(ctx, user)
	cli, err := wrest.BuildWorkspacesClientForUser(ctx, u)
	if err != nil {
		return err
	}

	w := restworkspacesv1alpha1.Workspace{}
	wk := types.NamespacedName{Name: iw.Spec.DisplayName, Namespace: iw.Status.Owner.Username}
	if err := cli.Get(ctx, wk, &w); err != nil {
		return err
	}

	pw := w.DeepCopy()
	pw.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibility(visibility)
	p := client.StrategicMergeFrom(pw)
	err = cli.Patch(ctx, &w, p)
	switch {
	case err == nil:
		return fmt.Errorf("expected forbidden error updating the workspace, got no error")
	case errors.IsForbidden(err):
		return nil
	default:
		return fmt.Errorf("expected forbidden error updating the workspace, got %v", err)
	}
}

func thenTheWorkspaceVisibilityIsSetTo(ctx context.Context, visibility string) error {
	w := tcontext.RetrieveInternalWorkspace(ctx)

	if w.Spec.Visibility != workspacesv1alpha1.InternalWorkspaceVisibility(visibility) {
		return fmt.Errorf(`expected visibility "%s", found "%s"`, visibility, w.Spec.Visibility)
	}
	return nil
}

func workspaceIsReadableForEveryone(ctx context.Context, cli cli.Cli, namespace, name string) error {
	sb := &toolchainv1alpha1.SpaceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-community", name),
			Namespace: namespace,
		},
	}
	if err := poll.WaitForConditionImmediately(ctx, func(ctx context.Context) (done bool, err error) {
		if err := cli.Get(ctx, client.ObjectKeyFromObject(sb), sb); err != nil {
			return false, client.IgnoreNotFound(err)
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("error waiting for space binding %s/%s to be created: %w", sb.Namespace, sb.Name, err)
	}

	switch {
	case "viewer" != sb.Spec.SpaceRole:
		return fmt.Errorf("expected SpaceBinding %s to have SpaceRole %s, found %s", sb.Name, "viewer", sb.Spec.SpaceRole)
	case name != sb.Spec.Space:
		return fmt.Errorf("expected referenced space to be %v, found %v", name, sb.Spec.Space)
	case workspacesv1alpha1.PublicViewerName != sb.Spec.MasterUserRecord:
		return fmt.Errorf("expected referenced MUR to be %v, found %v", workspacesv1alpha1.PublicViewerName, sb.Spec.MasterUserRecord)
	default:
		return nil
	}
}

func thenTheWorkspaceVisibilityIsUpdatedTo(ctx context.Context, visibility string) error {
	w := tcontext.RetrieveInternalWorkspace(ctx)
	cli := tcontext.RetrieveHostClient(ctx)
	wk := client.ObjectKeyFromObject(&w)
	return poll.WaitForConditionImmediately(ctx, func(ctx context.Context) (done bool, err error) {
		if err := cli.Get(ctx, wk, &w); err != nil {
			return false, err
		}

		if w.Spec.Visibility != workspacesv1alpha1.InternalWorkspaceVisibility(visibility) {
			return false, nil
		}

		return true, nil
	})
}
