package kube

import (
	"context"
	"slices"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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
	// list community workspaces
	ww := workspacesv1alpha1.WorkspaceList{}
	if err := c.listCommunityWorkspaces(ctx, &ww); err != nil {
		return err
	}

	// fetch workspaces to which the user has direct access and that are visibile to the whole community
	if err := c.fetchMissingWorkspaces(ctx, user, &ww); err != nil {
		return err
	}

	// override workspaces namespace with owner
	for _, w := range ww.Items {
		ow, err := workspacesutil.GetOwner(&w)
		if err != nil {
			continue
		}
		w.Namespace = *ow
		objs.Items = append(objs.Items, w)
	}

	// TODO: apply label selection

	return nil
}

func (c *ReadClient) fetchMissingWorkspaces(ctx context.Context, user string, workspaces *workspacesv1alpha1.WorkspaceList) error {
	// list user's space bindings
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.listUserSpaceBindings(ctx, user, &sbb); err != nil {
		return err
	}

	// filter already fetched Workspaces
	fsbb := slices.DeleteFunc(sbb.Items, func(sb toolchainv1alpha1.SpaceBinding) bool {
		return slices.ContainsFunc(workspaces.Items, func(w workspacesv1alpha1.Workspace) bool {
			return w.Name == sb.Spec.Space
		})
	})

	for _, sb := range fsbb {
		k := c.workspaceNamespacedName(sb.Spec.Space)
		w := workspacesv1alpha1.Workspace{}
		if err := c.backend.Get(ctx, k, &w, &client.GetOptions{}); err != nil {
			continue
		}

		workspaces.Items = append(workspaces.Items, w)
	}
	return nil
}

func (c *ReadClient) listUserSpaceBindings(
	ctx context.Context,
	user string,
	spaceBindings *toolchainv1alpha1.SpaceBindingList,
) error {
	rmur, err := labels.NewRequirement(toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey, selection.Equals, []string{user})
	if err != nil {
		return err
	}

	opts := &client.ListOptions{LabelSelector: labels.NewSelector().Add(*rmur)}
	return c.backend.List(ctx, spaceBindings, opts)
}

func (c *ReadClient) listCommunityWorkspaces(ctx context.Context, workspaces *workspacesv1alpha1.WorkspaceList) error {
	r, err := labels.NewRequirement(LabelWorkspaceVisibility, selection.Equals, []string{string(workspacesv1alpha1.WorkspaceVisibilityCommunity)})
	if err != nil {
		return err
	}

	opts := &client.ListOptions{LabelSelector: labels.NewSelector().Add(*r)}
	return c.backend.List(ctx, workspaces, opts)
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
	w := &workspacesv1alpha1.Workspace{}
	err := c.backend.Get(ctx, c.workspaceNamespacedName(space), w, opts...)
	if err != nil {
		return err
	}
	w.SetNamespace(owner)

	// if workspace visibility is community all users are allowed visibility
	if w.Spec.Visibility == workspacesv1alpha1.WorkspaceVisibilityCommunity {
		w.DeepCopyInto(obj)
		return nil
	}

	// chek if user has direct visibility on the space
	ok, err := c.existsSpaceBindingForUserAndSpace(ctx, user, space)
	if err != nil {
		return err
	}
	if !ok {
		return kerrors.NewNotFound(workspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), space)
	}

	// overwrite namespace with owner's complaint username
	ll := w.GetLabels()
	if len(ll) == 0 {
		return kerrors.NewNotFound(workspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), space)
	}

	if ow, ok := ll[workspacesv1alpha1.LabelWorkspaceOwner]; !ok || ow != owner {
		return kerrors.NewNotFound(workspacesv1alpha1.GroupVersion.WithResource("workspaces").GroupResource(), space)
	}

	// return workspace
	w.DeepCopyInto(obj)
	return nil
}

func (c *ReadClient) existsSpaceBindingForUserAndSpace(ctx context.Context, user, space string) (bool, error) {
	rmur, err := labels.NewRequirement(toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey, selection.Equals, []string{user})
	if err != nil {
		return false, err
	}
	rspc, err := labels.NewRequirement(toolchainv1alpha1.SpaceBindingSpaceLabelKey, selection.Equals, []string{space})
	if err != nil {
		return false, err
	}
	ls := labels.NewSelector().Add(*rmur, *rspc)
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.backend.List(ctx, &sbb, &client.ListOptions{LabelSelector: ls}); err != nil {
		return false, err
	}

	return len(sbb.Items) > 0, nil
}

func (c *ReadClient) workspaceNamespacedName(space string) client.ObjectKey {
	return types.NamespacedName{Namespace: c.workspacesNamespace, Name: space}
}
