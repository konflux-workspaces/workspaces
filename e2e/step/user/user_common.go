package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/pkg/poll"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
)

const DefaultUserName string = "default-user"

func OnBoardUserInKubespaceNamespace(ctx context.Context, name string) (*toolchainv1alpha1.UserSignup, error) {
	cli := tcontext.RetrieveHostClient(ctx)
	ns := tcontext.RetrieveKubespaceNamespace(ctx)

	u, err := OnboardUser(ctx, cli, ns, name)
	return u, err
}

func OnboardUser(ctx context.Context, cli cli.Cli, namespace, name string) (*toolchainv1alpha1.UserSignup, error) {
	e := fmt.Sprintf("%s@test.test", name)
	h := md5.New()
	h.Write([]byte(e))
	eh := hex.EncodeToString(h.Sum(nil))

	u := toolchainv1alpha1.UserSignup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				toolchainv1alpha1.UserSignupUserEmailHashLabelKey: string(eh),
			},
		},
		Spec: toolchainv1alpha1.UserSignupSpec{
			IdentityClaims: toolchainv1alpha1.IdentityClaimsEmbedded{
				PropagatedClaims: toolchainv1alpha1.PropagatedClaims{
					Email: e,
				},
				PreferredUsername: cli.EnsurePrefix(name),
			},
			States: []toolchainv1alpha1.UserSignupState{toolchainv1alpha1.UserSignupStateApproved},
		},
	}
	if err := cli.Create(ctx, &u); err != nil {
		return nil, err
	}

	lu := toolchainv1alpha1.UserSignup{}
	if err := poll.WaitForConditionImmediately(ctx, func(ctx context.Context) (done bool, err error) {
		if err := cli.Get(ctx, client.ObjectKeyFromObject(&u), &lu); err != nil {
			return false, client.IgnoreNotFound(err)
		}

		if lu.Status.CompliantUsername == "" {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, fmt.Errorf("error waiting for CompliantUsername of user %s/%s: %w", u.Namespace, u.Name, err)
	}

	return &lu, nil
}
