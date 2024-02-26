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
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

type Cache struct {
	c                   cache.Cache
	ctx                 context.Context
	workspacesNamespace string
}

func NewCache(ctx context.Context, cfg *rest.Config, workspacesNamespace string) (*Cache, error) {
	s := runtime.NewScheme()
	s.AddKnownTypes(toolchainv1alpha1.GroupVersion, &toolchainv1alpha1.SpaceBinding{})
	s.AddKnownTypes(workspacesv1alpha1.GroupVersion, &workspacesv1alpha1.Workspace{})

	c, err := cache.New(cfg, cache.Options{
		Scheme:                      s,
		ReaderFailOnMissingInformer: true,
		// look into DefaultTransform to add some labels and/or remove unwanted/internals properties
	})
	if err != nil {
		return nil, err
	}

	return &Cache{
		c:                   c,
		ctx:                 ctx,
		workspacesNamespace: workspacesNamespace,
	}, nil
}

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

func (c *Cache) workspaceNamespacedName(space string) client.ObjectKey {
	return types.NamespacedName{Namespace: c.workspacesNamespace, Name: space}
}
