package cache

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

const (
	LabelWorkspaceVisibility string = "workspaces.io/visibility"
)

// NewCache creates a controller-runtime cache.Cache instance configured to monitor
// spacebindings.toolchain.dev.openshift.com and workspaces.workspaces.io.
// IMPORTANT: returned cache needs to be started and initialized.
func NewCache(ctx context.Context, cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (cache.Cache, error) {
	s, err := createScheme()
	if err != nil {
		return nil, err
	}

	m, err := createMapper(cfg)
	if err != nil {
		return nil, err
	}

	c, err := newCache(cfg, s, m, workspacesNamespace, kubesawNamespace)
	if err != nil {
		return nil, err
	}

	if _, err := c.GetInformer(ctx, &toolchainv1alpha1.SpaceBinding{}); err != nil {
		return nil, err
	}
	if _, err := c.GetInformer(ctx, &toolchainv1alpha1.UserSignup{}); err != nil {
		return nil, err
	}
	if _, err := c.GetInformer(ctx, &workspacesv1alpha1.InternalWorkspace{}); err != nil {
		return nil, err
	}

	return c, nil
}

func newCache(cfg *rest.Config, scheme *runtime.Scheme, mapper meta.RESTMapper, workspacesNamespace, kubesawNamespace string) (cache.Cache, error) {
	return cache.New(cfg, cache.Options{
		Scheme:                      scheme,
		Mapper:                      mapper,
		ReaderFailOnMissingInformer: true,
		ByObject: map[client.Object]cache.ByObject{
			&toolchainv1alpha1.UserSignup{}:         {Namespaces: map[string]cache.Config{kubesawNamespace: {}}},
			&toolchainv1alpha1.SpaceBinding{}:       {Namespaces: map[string]cache.Config{kubesawNamespace: {}}},
			&workspacesv1alpha1.InternalWorkspace{}: {Namespaces: map[string]cache.Config{workspacesNamespace: {}}},
		},
		DefaultTransform: func(obj interface{}) (interface{}, error) {
			if ws, ok := obj.(*workspacesv1alpha1.InternalWorkspace); ok {
				if ws.Labels == nil {
					ws.Labels = map[string]string{}
				}
				ws.Labels[LabelWorkspaceVisibility] = string(ws.Spec.Visibility)
				return ws, nil
			}

			return obj, nil
		},
	})
}

func createMapper(cfg *rest.Config) (meta.RESTMapper, error) {
	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}

	return apiutil.NewDynamicRESTMapper(cfg, hc)
}

func createScheme() (*runtime.Scheme, error) {
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
	return s, nil
}
