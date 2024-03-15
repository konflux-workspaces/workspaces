package kube

import (
	"context"
	"slices"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"

	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/workspacesutil"
)

var (
	_ workspace.WorkspaceLister = &ReadClient{}
	_ workspace.WorkspaceReader = &ReadClient{}
)

// ReadClient implements the WorkspaceLister and WorkspaceReader interfaces
// using a client.Reader as backend
type ReadClient struct {
	backend client.Reader

	kubesawNamespace    string
	workspacesNamespace string
}

// NewReadClientWithCache creates a controller-runtime cache and use it as KubeReadClient's backend.
func NewReadClientWithCache(ctx context.Context, cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (*ReadClient, cache.Cache, error) {
	c, err := NewCache(ctx, cfg, workspacesNamespace, kubesawNamespace)
	if err != nil {
		return nil, nil, err
	}

	return NewReadClientWithReader(c, workspacesNamespace, kubesawNamespace), c, nil
}

// NewReadClientWithReader creates a new KubeReadClient with the provided backend
func NewReadClientWithReader(backend client.Reader, workspacesNamespace, kubesawNamespace string) *ReadClient {
	return &ReadClient{
		backend: backend,

		kubesawNamespace:    kubesawNamespace,
		workspacesNamespace: workspacesNamespace,
	}
}

// ListUserWorkspaces Returns all the workspaces the user has access to
func (c *ReadClient) ListUserWorkspaces(
	ctx context.Context,
	user string,
	objs *workspacesv1alpha1.WorkspaceList,
	opts ...client.ListOption,
) error {
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.backend.List(ctx, &sbb, &client.ListOptions{}); err != nil {
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
		if err := c.backend.Get(ctx, k, &w, &client.GetOptions{}); err != nil {
			continue
		}

		ow, err := workspacesutil.GetOwner(&w)
		if err != nil {
			continue
		}
		w.Namespace = *ow
		objs.Items = append(objs.Items, w)
	}
	return nil
}

// ReadUserWorkspace Returns the Workspace details only if the user has access to it
func (c *ReadClient) ReadUserWorkspace(
	ctx context.Context,
	user string,
	owner string,
	space string,
	obj *workspacesv1alpha1.Workspace,
	opts ...client.GetOption,
) error {
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.backend.List(ctx, &sbb, &client.ListOptions{}); err != nil {
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
	err := c.backend.Get(ctx, c.workspaceNamespacedName(space), w, opts...)
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

func (c *ReadClient) workspaceNamespacedName(space string) client.ObjectKey {
	return types.NamespacedName{Namespace: c.workspacesNamespace, Name: space}
}
