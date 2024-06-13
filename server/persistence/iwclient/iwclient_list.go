package iwclient

import (
	"context"
	"fmt"

	"k8s.io/utils/set"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// ListAsUser lists all the community workspaces together with the ones the user is allowed access to
func (c *Client) ListAsUser(ctx context.Context, user string, workspaces *workspacesv1alpha1.InternalWorkspaceList) error {
	// list community workspaces
	ww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := c.listCommunityWorkspaces(ctx, &ww); err != nil {
		return fmt.Errorf("error retrieving community workspaces: %w", err)
	}

	// fetch workspaces to which the user has direct access and that are visibile to the whole community
	if err := c.fetchMissingWorkspaces(ctx, user, &ww); err != nil {
		return fmt.Errorf("error fetching directly accessible workspaces: %w", err)
	}

	// deepcopy result
	ww.DeepCopyInto(workspaces)
	return nil
}

func (c *Client) fetchMissingWorkspaces(ctx context.Context, user string, workspaces *workspacesv1alpha1.InternalWorkspaceList) error {
	// retrieve names of missing workspaces
	nmww, err := c.calculateNamesOfMissingWorkspaces(ctx, user, workspaces)
	if err != nil {
		return err
	}

	// add workspaces to which the user has direct access to return list
	for _, s := range nmww {
		aww := workspacesv1alpha1.InternalWorkspaceList{}
		opt := client.MatchingFields{cache.IndexKeyInternalWorkspaceSpaceName: s}
		if err := c.backend.List(ctx, &aww, opt); err != nil {
			return err
		}
		workspaces.Items = append(workspaces.Items, aww.Items...)
	}
	return nil
}

func (c *Client) calculateNamesOfMissingWorkspaces(ctx context.Context, user string, workspaces *workspacesv1alpha1.InternalWorkspaceList) ([]string, error) {
	// list user's space bindings
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.listUserSpaceBindings(ctx, user, &sbb); err != nil {
		return nil, err
	}

	// retrieve all the spaces
	sbbn := set.New[string]()
	for _, sb := range sbb.Items {
		sbbn = sbbn.Insert(sb.Spec.Space)
	}

	// remove already fetched spaces
	for _, w := range workspaces.Items {
		sbbn = sbbn.Delete(w.Status.Space.Name)
	}

	// return missing spaces
	return sbbn.UnsortedList(), nil
}

func (c *Client) listUserSpaceBindings(
	ctx context.Context,
	user string,
	spaceBindings *toolchainv1alpha1.SpaceBindingList,
) error {
	opt := client.MatchingLabels{toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: user}
	return c.backend.List(ctx, spaceBindings, opt)
}

func (c *Client) listCommunityWorkspaces(ctx context.Context, workspaces *workspacesv1alpha1.InternalWorkspaceList) error {
	opt := client.MatchingFields{
		cache.IndexKeyInternalWorkspaceVisibility: string(workspacesv1alpha1.InternalWorkspaceVisibilityCommunity),
	}
	return c.backend.List(ctx, workspaces, opt)
}
