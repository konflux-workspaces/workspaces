package iwclient

import (
	"context"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
)

func (c *Client) ListAsUser(ctx context.Context, user string, workspaces *workspacesv1alpha1.InternalWorkspaceList) error {
	// list community workspaces
	ww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := c.listCommunityWorkspaces(ctx, &ww); err != nil {
		return err
	}

	// fetch workspaces to which the user has direct access and that are visibile to the whole community
	if err := c.fetchMissingWorkspaces(ctx, user, &ww); err != nil {
		return err
	}

	// deepcopy result
	ww.DeepCopyInto(workspaces)
	return nil
}

func (c *Client) fetchMissingWorkspaces(ctx context.Context, user string, workspaces *workspacesv1alpha1.InternalWorkspaceList) error {
	// list user's space bindings
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.listUserSpaceBindings(ctx, user, &sbb); err != nil {
		return err
	}

	// filter already fetched Workspaces
	fsbb := make([]*toolchainv1alpha1.SpaceBinding, 0, len(sbb.Items))
	for i, sb := range sbb.Items {
		if slices.ContainsFunc(workspaces.Items, func(w workspacesv1alpha1.InternalWorkspace) bool {
			return w.Name == sb.Spec.Space
		}) {
			continue
		}

		fsbb = append(fsbb, &sbb.Items[i])
	}

	for _, sb := range fsbb {
		k := c.workspaceNamespacedName(sb.Spec.Space)
		w := workspacesv1alpha1.InternalWorkspace{}
		if err := c.backend.Get(ctx, k, &w, &client.GetOptions{}); err != nil {
			continue
		}

		workspaces.Items = append(workspaces.Items, w)
	}
	return nil
}

func (c *Client) listUserSpaceBindings(
	ctx context.Context,
	user string,
	spaceBindings *toolchainv1alpha1.SpaceBindingList,
) error {
	opts := []client.ListOption{
		client.MatchingLabels{toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: user},
	}
	return c.backend.List(ctx, spaceBindings, opts...)
}

func (c *Client) listCommunityWorkspaces(ctx context.Context, workspaces *workspacesv1alpha1.InternalWorkspaceList) error {
	opts := []client.ListOption{
		client.MatchingLabels{cache.LabelWorkspaceVisibility: string(workspacesv1alpha1.InternalWorkspaceVisibilityCommunity)},
	}
	return c.backend.List(ctx, workspaces, opts...)
}
