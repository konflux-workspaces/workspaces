package hook

import (
	"context"
	"errors"
	"time"

	"github.com/cucumber/godog"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	// corev1 "k8s.io/api/core/v1"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

func deleteResources(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	// skip if error is not nil
	if err != nil {
		return ctx, err
	}

	cli := tcontext.RetrieveHostClient(ctx)

	errs := []error{}
	{
		usl := &toolchainv1alpha1.UserSignupList{}
		if err := cli.Client.List(ctx, usl, client.InNamespace(metav1.NamespaceAll)); err != nil {
			errs = append(errs, err)
		} else {
			for _, r := range usl.Items {
				if !cli.HasScenarioPrefix(r.Name) {
					continue
				}

				if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
					if err := cli.Delete(ctx, &r); err != nil {
						if kerrors.IsNotFound(err) {
							return true, nil
						}
						return false, err
					}
					return true, nil
				}); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	{
		murl := &toolchainv1alpha1.MasterUserRecordList{}
		if err := cli.Client.List(ctx, murl, client.InNamespace(metav1.NamespaceAll)); err != nil {
			errs = append(errs, err)
		} else {
			for _, r := range murl.Items {
				if !cli.HasScenarioPrefix(r.Name) {
					continue
				}

				if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
					if err := cli.Delete(ctx, &r); err != nil {
						if kerrors.IsNotFound(err) {
							return true, nil
						}
						return false, err
					}
					return true, nil
				}); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	{
		spl := &toolchainv1alpha1.SpaceList{}
		if err := cli.Client.List(ctx, spl, client.InNamespace(metav1.NamespaceAll)); err != nil {
			errs = append(errs, err)
		} else {
			for _, r := range spl.Items {
				if !cli.HasScenarioPrefix(r.Name) {
					continue
				}

				if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
					if err := cli.Delete(ctx, &r); err != nil {
						if kerrors.IsNotFound(err) {
							return true, nil
						}
						return false, err
					}
					return true, nil
				}); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	{
		sbl := &toolchainv1alpha1.SpaceBindingList{}
		if err := cli.Client.List(ctx, sbl, client.InNamespace(metav1.NamespaceAll)); err != nil {
			errs = append(errs, err)
		} else {
			for _, r := range sbl.Items {
				if !cli.HasScenarioPrefix(r.Name) {
					continue
				}

				if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
					if err := cli.Delete(ctx, &r); err != nil {
						if kerrors.IsNotFound(err) {
							return true, nil
						}
						return false, err
					}
					return true, nil
				}); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	{
		wl := &workspacesv1alpha1.InternalWorkspaceList{}
		if err := cli.Client.List(ctx, wl, client.InNamespace(metav1.NamespaceAll)); err != nil {
			errs = append(errs, err)
		} else {
			for _, r := range wl.Items {
				if !cli.HasScenarioPrefix(r.Name) {
					continue
				}

				if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
					if err := cli.Delete(ctx, &r); err != nil {
						if kerrors.IsNotFound(err) {
							return true, nil
						}
						return false, err
					}
					return true, nil
				}); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	return ctx, errors.Join(errs...)
}
