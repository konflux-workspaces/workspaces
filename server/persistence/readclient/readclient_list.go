package readclient

import (
	"context"
	"fmt"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/persistence/mutate"
)

var _ workspace.WorkspaceLister = &ReadClient{}

// ListUserWorkspaces Returns all the workspaces the user has access to
func (c *ReadClient) ListUserWorkspaces(
	ctx context.Context,
	user string,
	objs *restworkspacesv1alpha1.WorkspaceList,
	opts ...client.ListOption,
) error {
	// retrieve workspaces visible to user
	iww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := c.internalClient.ListAsUser(ctx, user, &iww); err != nil {
		return kerrors.NewInternalError(fmt.Errorf("error retrieving the list of workspaces for user %v", user))
	}

	// map list options
	listOpts, err := mapListOptions(opts...)
	if err != nil {
		return err
	}

	// filter internal workspaces
	fiww, err := filterByLabels(&iww, listOpts)
	if err != nil {
		return err
	}

	// map back to Workspaces
	ww, err := c.mapper.InternalWorkspaceListToWorkspaceList(fiww)
	if err != nil {
		return kerrors.NewInternalError(fmt.Errorf("error retrieving the list of workspaces for user %v", user))
	}

	// filter by namespace
	filterByNamespace(ww, listOpts.Namespace)

	// apply is-owner label
	for i := range ww.Items {
		mutate.ApplyIsOwnerLabel(&ww.Items[i], user)
	}

	ww.DeepCopyInto(objs)
	return nil
}

func filterByNamespace(ww *restworkspacesv1alpha1.WorkspaceList, namespace string) {
	if namespace == "" {
		return
	}

	fww := []restworkspacesv1alpha1.Workspace{}
	for _, w := range ww.Items {
		if w.Namespace == namespace {
			fww = append(fww, w)
		}
	}
	ww.Items = fww
}

func filterByLabels(ww *workspacesv1alpha1.InternalWorkspaceList, listOpts *client.ListOptions) (*workspacesv1alpha1.InternalWorkspaceList, error) {
	rww := workspacesv1alpha1.InternalWorkspaceList{}
	for _, w := range ww.Items {
		// selection
		if !matchesListOpts(listOpts, w.GetLabels()) {
			continue
		}

		rww.Items = append(rww.Items, w)
	}
	return &rww, nil
}

func mapListOptions(opts ...client.ListOption) (*client.ListOptions, error) {
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	if listOpts.LabelSelector == nil {
		return &listOpts, nil
	}

	rr, _ := listOpts.LabelSelector.Requirements()
	for _, ls := range rr {
		if strings.HasPrefix(ls.Key(), workspacesv1alpha1.LabelInternalDomain) {
			return nil, fmt.Errorf("invalid label selector: key '%s' is reserved", ls.Key())
		}
	}

	return &listOpts, nil
}

func matchesListOpts(
	listOpts *client.ListOptions,
	objLabels map[string]string,
) bool {
	return objLabels == nil || listOpts == nil || listOpts.LabelSelector == nil ||
		listOpts.LabelSelector.Matches(labels.Set(objLabels))
}
