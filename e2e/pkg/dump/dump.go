package dump

import (
	"context"
	"errors"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	yaml "sigs.k8s.io/yaml/goyaml.v3"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

var resourcesToDump = []client.ObjectList{
	&workspacesv1alpha1.InternalWorkspaceList{},
	&toolchainv1alpha1.UserSignupList{},
	&toolchainv1alpha1.MasterUserRecordList{},
	&toolchainv1alpha1.SpaceList{},
	&toolchainv1alpha1.SpaceBindingList{},
}

func DumpAll(ctx context.Context) error {
	// retrieve host client
	cli := tcontext.RetrieveHostClient(ctx)

	errs := []error{}
	for _, r := range resourcesToDump {
		// retrieve gvk for client.object
		gvk, err := cli.GroupVersionKindFor(r)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if err := dumpResourceInAllNamespaces(ctx, cli.Client, gvk); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func dumpResourceInAllNamespaces(ctx context.Context, cli client.Client, gvk schema.GroupVersionKind) error {
	fmt.Fprintf(os.Stderr, "*** Dump: %s\n", gvk.String())

	// list resource as UnstructuredList
	list, err := listAsUnstructuredList(ctx, cli, gvk)
	if err != nil {
		return err
	}

	// dump resources
	return dumpUnstructuredList(list)
}

func listAsUnstructuredList(ctx context.Context, cli client.Client, gvk schema.GroupVersionKind) (*unstructured.UnstructuredList, error) {
	// build UnstructuredList
	d := &unstructured.UnstructuredList{}
	d.SetGroupVersionKind(gvk)

	// list resources as UnstructuredList
	if err := cli.List(ctx, d, client.InNamespace(metav1.NamespaceAll)); err != nil {
		return nil, err
	}
	return d, nil
}

func dumpUnstructuredList(list *unstructured.UnstructuredList) error {
	// marshal to yaml
	o, err := yaml.Marshal(list)
	if err != nil {
		return err
	}

	// print on stderr
	if _, err := fmt.Fprintln(os.Stderr, string(o)); err != nil {
		return err
	}
	return nil
}
