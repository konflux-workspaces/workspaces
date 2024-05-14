package dump

import (
	"context"
	"errors"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	yaml "sigs.k8s.io/yaml/goyaml.v3"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

var resourcesToDump = func() []client.ObjectList {
	return []client.ObjectList{
		&workspacesv1alpha1.InternalWorkspaceList{},
		&toolchainv1alpha1.UserSignupList{},
		&toolchainv1alpha1.MasterUserRecordList{},
		&toolchainv1alpha1.SpaceList{},
		&toolchainv1alpha1.SpaceBindingList{},
	}
}

func DumpAll(ctx context.Context) error {
	rr := resourcesToDump()

	errs := []error{}
	for _, r := range rr {
		err := dumpResourceInAllNamespaces(ctx, r)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func dumpResourceInAllNamespaces(ctx context.Context, resource client.ObjectList) error {
	// retrieve host client
	cli := tcontext.RetrieveHostClient(ctx)

	// list resources
	if err := cli.Client.List(ctx, resource, client.InNamespace(metav1.NamespaceAll)); err != nil {
		return err
	}

	// marshal to yaml
	o, err := yaml.Marshal(resource)
	if err != nil {
		return err
	}

	// print on stderr
	if _, err := fmt.Fprintln(os.Stderr, string(o)); err != nil {
		return err
	}

	return nil
}
