package cache

import (
	"context"
	"errors"
	"slices"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/filariow/workspaces/server/core/workspace"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

var (
	_ workspace.WorkspaceLister = &Cache{}
	_ workspace.WorkspaceReader = &Cache{}
)

type Cache struct {
	c                   cache.Cache
	kubesawNamespace    string
	workspacesNamespace string
}

// New creates a new Cache that caches Workspaces and SpaceBindings. The cache
// provides methods to retrieve the workspaces the user is allowed to access
func New(cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (*Cache, error) {
	s := runtime.NewScheme()
	s.AddKnownTypes(toolchainv1alpha1.GroupVersion, &toolchainv1alpha1.SpaceBinding{}, &toolchainv1alpha1.SpaceBindingList{})
	s.AddKnownTypes(workspacesv1alpha1.GroupVersion, &workspacesv1alpha1.Workspace{}, &workspacesv1alpha1.WorkspaceList{})

	c, err := cache.New(cfg, cache.Options{
		Scheme:                      s,
		ReaderFailOnMissingInformer: true,
		DefaultNamespaces: map[string]cache.Config{
			workspacesNamespace: {},
			kubesawNamespace:    {},
		},
		// look into DefaultTransform to add some labels and/or remove unwanted/internals properties
	})
	if err != nil {
		return nil, err
	}

	return &Cache{
		c:                   c,
		kubesawNamespace:    kubesawNamespace,
		workspacesNamespace: workspacesNamespace,
	}, nil
}

// ListUserWorkspaces Returns all the workspaces the user has access to
func (c *Cache) ListUserWorkspaces(ctx context.Context, user string, objs *workspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error {
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.c.List(ctx, &sbb, &client.ListOptions{}); err != nil {
		return err
	}
	if len(sbb.Items) == 0 {
		return nil
	}

	errs := []error{}
	for _, sb := range sbb.Items {
		if sb.Spec.MasterUserRecord != user {
			continue
		}

		k := c.workspaceNamespacedName(sb.Spec.Space)
		w := workspacesv1alpha1.Workspace{}
		if err := c.c.Get(ctx, k, &w, &client.GetOptions{}); err != nil {
			errs = append(errs, err)
			continue
		}

		objs.Items = append(objs.Items, w)
	}
	return errors.Join(errs...)
}

// ReadUserWorkspace Returns the Workspace details only if the user has access to it
func (c *Cache) ReadUserWorkspace(ctx context.Context, user string, space string, obj *workspacesv1alpha1.Workspace, opts ...client.GetOption) error {
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.c.List(ctx, &sbb, &client.ListOptions{}); err != nil {
		return err
	}
	if len(sbb.Items) == 0 {
		return nil
	}

	if !slices.ContainsFunc(sbb.Items, func(sb toolchainv1alpha1.SpaceBinding) bool {
		return sb.Spec.MasterUserRecord == user
	}) {
		return kerrors.NewNotFound(workspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), space)
	}

	return c.c.Get(ctx, c.workspaceNamespacedName(space), obj, opts...)
}

// WaitForCacheSync Synchronizes the cache
func (c *Cache) WaitForCacheSync(ctx context.Context) bool {
	return c.c.WaitForCacheSync(ctx)
}

// Start starts the cache. It Blocks.
func (c *Cache) Start(ctx context.Context) error {
	return c.c.Start(ctx)
}

func (c *Cache) workspaceNamespacedName(space string) client.ObjectKey {
	return types.NamespacedName{Namespace: c.workspacesNamespace, Name: space}
}
