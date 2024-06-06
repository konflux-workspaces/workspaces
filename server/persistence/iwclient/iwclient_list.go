package iwclient

import (
	"cmp"
	"context"
	"slices"
	"sort"

	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
)

// ListAsUser lists all the community workspaces together with the ones the user is allowed access to
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

	// fetch all workspaces
	aww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := c.backend.List(ctx, &aww); err != nil {
		return err
	}

	// filter already fetched Workspaces
	fsp := make([]string, 0, len(sbb.Items))
	for i, sb := range sbb.Items {
		if slices.ContainsFunc(workspaces.Items, func(w workspacesv1alpha1.InternalWorkspace) bool {
			return w.Status.Space.Name == sb.Spec.Space
		}) {
			continue
		}

		fsp = append(fsp, sbb.Items[i].Spec.Space)
	}
	sort.Strings(fsp)
	fsp = slices.CompactFunc(fsp, func(s1, s2 string) bool { return cmp.Compare(s1, s2) <= 0 })

	// add workspaces to which the user has direct access to return list
	for _, s := range fsp {
		if i := slices.IndexFunc(aww.Items, func(w workspacesv1alpha1.InternalWorkspace) bool {
			return w.Status.Space.Name == s
		}); i != -1 {
			workspaces.Items = append(workspaces.Items, aww.Items[i])
		}
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
