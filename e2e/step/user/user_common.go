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
	email := fmt.Sprintf("%s@test.test", name)
	prefixedUsername := cli.EnsurePrefix(name)

	// apply mutators
	u := approvedUserSignup(name, namespace, email, prefixedUsername, prefixedUsername)

	if err := onboardUser(ctx, cli, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func OnboardUserWithSub(ctx context.Context, cli cli.Cli, namespace, name, sub string) (*toolchainv1alpha1.UserSignup, error) {
	email := fmt.Sprintf("%s@test.test", name)
	prefixedUsername := cli.EnsurePrefix(name)

	u := approvedUserSignup(name, namespace, email, sub, prefixedUsername)
	if err := onboardUser(ctx, cli, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func onboardUser(ctx context.Context, cli cli.Cli, userSignup *toolchainv1alpha1.UserSignup) error {
	// create the UserSignup
	if err := cli.Create(ctx, userSignup); err != nil {
		return err
	}

	// wait for UserSignup to be processed and for the CompliantUsername to be generated
	lu := userSignup.DeepCopy()
	if err := poll.WaitForConditionImmediately(ctx, func(ctx context.Context) (done bool, err error) {
		if err := cli.Get(ctx, client.ObjectKeyFromObject(userSignup), lu); err != nil {
			return false, client.IgnoreNotFound(err)
		}

		if lu.Status.CompliantUsername == "" {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("error waiting for CompliantUsername of user %s/%s: %w", userSignup.Namespace, userSignup.Name, err)
	}

	lu.DeepCopyInto(userSignup)
	return nil
}

func approvedUserSignup(name, namespace, email, sub, preferredUsername string) toolchainv1alpha1.UserSignup {
	h := md5.New()
	h.Write([]byte(email))
	eh := hex.EncodeToString(h.Sum(nil))

	return toolchainv1alpha1.UserSignup{
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
					Email: email,
					Sub:   sub,
				},
				PreferredUsername: preferredUsername,
			},
			States: []toolchainv1alpha1.UserSignupState{toolchainv1alpha1.UserSignupStateApproved},
		},
	}
}
