package cache

import (
	"context"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
)

var (
	_ workspace.WorkspaceLister = &Cache{}
	_ workspace.WorkspaceReader = &Cache{}
)

type Cache struct {
	c client.Reader

	kubesawNamespace    string
	workspacesNamespace string
}

func NewCRCache(ctx context.Context, cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (cache.Cache, error) {
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

// NewWithCRCache creates a new Cache that caches Workspaces and SpaceBindings. The cache
// provides methods to retrieve the workspaces the user is allowed to access
func NewWithCRCache(ctx context.Context, cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (*Cache, cache.Cache, error) {
	c, err := NewCRCache(ctx, cfg, workspacesNamespace, kubesawNamespace)
	if err != nil {
		return nil, nil, err
	}

	return NewWithReader(c, workspacesNamespace, kubesawNamespace), c, nil
}

// NewWithReader creates a new Cache that caches Workspaces and SpaceBindings.
func NewWithReader(reader client.Reader, workspacesNamespace, kubesawNamespace string) *Cache {
	return &Cache{
		c: reader,

		kubesawNamespace:    kubesawNamespace,
		workspacesNamespace: workspacesNamespace,
	}
}

// ListUserWorkspaces Returns all the workspaces the user has access to
func (c *Cache) ListUserWorkspaces(
	ctx context.Context,
	user string,
	objs *workspacesv1alpha1.WorkspaceList,
	opts ...client.ListOption,
) error {
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.c.List(ctx, &sbb, &client.ListOptions{}); err != nil {
		return err
	}

	if len(sbb.Items) == 0 {
		return nil
	}

	for _, sb := range sbb.Items {
		if sb.Spec.MasterUserRecord != user {
			continue
		}

		k := c.workspaceNamespacedName(sb.Spec.Space)
		w := workspacesv1alpha1.Workspace{}
		if err := c.c.Get(ctx, k, &w, &client.GetOptions{}); err != nil {
			continue
		}

		ow, err := getOwner(&w)
		if err != nil {
			continue
		}
		w.Namespace = *ow
		objs.Items = append(objs.Items, w)
	}
	return nil
}

var errWorkspaceWithoutOwner = fmt.Errorf("error workspace has no owner")

func getOwner(w *workspacesv1alpha1.Workspace) (*string, error) {
	ll := w.GetLabels()
	if len(ll) == 0 {
		return nil, fmt.Errorf("%w: %v", errWorkspaceWithoutOwner, client.ObjectKeyFromObject(w))
	}

	ol, ok := ll[workspacesv1alpha1.LabelWorkspaceOwner]
	if !ok || ol == "" {
		return nil, fmt.Errorf("%w: %v", errWorkspaceWithoutOwner, client.ObjectKeyFromObject(w))
	}
	return &ol, nil
}

// ReadUserWorkspace Returns the Workspace details only if the user has access to it
func (c *Cache) ReadUserWorkspace(
	ctx context.Context,
	user string,
	owner string,
	space string,
	obj *workspacesv1alpha1.Workspace,
	opts ...client.GetOption,
) error {
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

	w := &workspacesv1alpha1.Workspace{}
	err := c.c.Get(ctx, c.workspaceNamespacedName(space), w, opts...)
	if err != nil {
		return err
	}

	ll := w.GetLabels()
	if len(ll) == 0 {
		return kerrors.NewNotFound(workspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), space)
	}

	if ow, ok := ll[workspacesv1alpha1.LabelWorkspaceOwner]; !ok || ow != owner {
		return kerrors.NewNotFound(workspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), space)
	}
	w.SetNamespace(owner)

	w.DeepCopyInto(obj)
	return nil
}

func (c *Cache) workspaceNamespacedName(space string) client.ObjectKey {
	return types.NamespacedName{Namespace: c.workspacesNamespace, Name: space}
}
