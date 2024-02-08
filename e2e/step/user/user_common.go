package user

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
)

const DefaultUserName string = "default-user"

func OnBoardUserInTestNamespace(ctx context.Context, name string) (*toolchainv1alpha1.MasterUserRecord, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveTestNamespace(ctx)

	u, err := OnboardUser(ctx, cli, ns, name)
	return u, err
}

func OnboardUser(ctx context.Context, cli client.Client, namespace, name string) (*toolchainv1alpha1.MasterUserRecord, error) {
	u := toolchainv1alpha1.MasterUserRecord{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: toolchainv1alpha1.MasterUserRecordSpec{},
	}
	if err := cli.Create(ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
