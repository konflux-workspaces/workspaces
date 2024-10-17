package workspace

import (
	"context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultWorkspaceTargetClusterURL = "https://default-cluster-url.io/"

func givenWorkspaceHasClusterURLSet(ctx context.Context) (context.Context, error) {
	return givenWorkspaceHasClusterURLSetTo(ctx, defaultWorkspaceTargetClusterURL)
}

func givenWorkspaceHasNoClusterURLSet(ctx context.Context) (context.Context, error) {
	return givenWorkspaceHasClusterURLSetTo(ctx, "")
}

func givenWorkspaceHasClusterURLSetTo(ctx context.Context, clusterURL string) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	kns := tcontext.RetrieveKubespaceNamespace(ctx)
	ws := tcontext.RetrieveInternalWorkspace(ctx)

	// retrieve space
	s := toolchainv1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ws.Name,
			Namespace: kns,
		},
	}
	if err := cli.Get(ctx, client.ObjectKeyFromObject(&s), &s); err != nil {
		return ctx, err
	}

	// update target cluster with default URL
	s.Status.TargetCluster = clusterURL
	if err := cli.Status().Update(ctx, &s); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func givenDefaultWorkspaceIsCreatedForThem(ctx context.Context) (context.Context, error) {
	return defaultWorkspaceIsCreatedForThem(ctx)
}

func givenDefaultWorkspaceIsCreatedForCustomUser(ctx context.Context, name string) (context.Context, error) {
	u := tcontext.RetrieveCustomUser(ctx, name)
	return defaultWorkspaceIsCreatedForUser(ctx, u.Status.CompliantUsername)
}

func givenAPrivateWorkspaceExists(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)

	u, w, err := createUserSignupAndWaitForWorkspace(ctx, cli, ns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	ctx = tcontext.InjectUser(ctx, *u)
	ctx = tcontext.InjectInternalWorkspace(ctx, *w)
	return ctx, nil
}

func givenACommunityWorkspaceExists(ctx context.Context) (context.Context, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	kns := tcontext.RetrieveKubespaceNamespace(ctx)

	u, w, err := createUserSignupAndWaitForWorkspace(ctx, cli, kns, user.DefaultUserName)
	if err != nil {
		return ctx, err
	}

	ctx = tcontext.InjectUser(ctx, *u)
	ctx = tcontext.InjectInternalWorkspace(ctx, *w)

	w.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityCommunity
	if err := cli.Update(ctx, w); err != nil {
		return ctx, err
	}

	if err := workspaceIsReadableForEveryone(ctx, cli, kns, w.Name); err != nil {
		return ctx, err
	}

	return ctx, nil
}
