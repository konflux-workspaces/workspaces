package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// NewCache creates a controller-runtime cache.Cache instance configured to monitor
// spacebindings.toolchain.dev.openshift.com and workspaces.workspaces.io.
// IMPORTANT: returned cache needs to be started and initialized.
func NewCache(ctx context.Context, cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (cache.Cache, error) {
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := metav1.AddMetaToScheme(s); err != nil {
		return nil, err
	}
	if err := workspacesv1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := toolchainv1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}

	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}

	m, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return nil, err
	}

	c, err := cache.New(cfg, cache.Options{
		Scheme:                      s,
		Mapper:                      m,
		ReaderFailOnMissingInformer: true,
		ByObject: map[client.Object]cache.ByObject{
			&toolchainv1alpha1.SpaceBinding{}: {Namespaces: map[string]cache.Config{kubesawNamespace: {}}},
			&workspacesv1alpha1.Workspace{}:   {Namespaces: map[string]cache.Config{workspacesNamespace: {}}},
		},
		// look into DefaultTransform to add some labels and/or remove unwanted/internals properties
	})
	if err != nil {
		return nil, err
	}

	if _, err := c.GetInformer(ctx, &toolchainv1alpha1.SpaceBinding{}); err != nil {
		return nil, err
	}
	if _, err := c.GetInformer(ctx, &workspacesv1alpha1.Workspace{}); err != nil {
		return nil, err
	}

	return c, nil
}
