package workspace

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
)

//go:generate mockgen -destination=mocks/mocks_list.go -package=mocks . WorkspaceLister

// ListWorkspaceQuery contains the information needed to retrieve all the workspaces the user has access to from the data source
type ListWorkspaceQuery struct {
	Namespace string
}

// ListWorkspaceResponse contains all the workspaces the user can access
type ListWorkspaceResponse struct {
	Workspaces restworkspacesv1alpha1.WorkspaceList
}

// WorkspaceLister is the interface the data source needs to implement to allow the ListWorkspaceHandler to fetch data from it
type WorkspaceLister interface {
	ListUserWorkspaces(ctx context.Context, user string, objs *restworkspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error
}

// ListWorkspaceHandler process ListWorkspaceQuery and returns a ListWorkspaceResponse fetching data from a WorkspaceLister
type ListWorkspaceHandler struct {
	lister WorkspaceLister
}

// NewListWorkspaceHandler creates a new ListWorkspaceHandler that uses a specified WorkspaceLister
func NewListWorkspaceHandler(lister WorkspaceLister) *ListWorkspaceHandler {
	return &ListWorkspaceHandler{lister: lister}
}

// Handle handles a ListWorkspaceQuery abd returns a ListWorkspaceResponse or an error
func (h *ListWorkspaceHandler) Handle(ctx context.Context, query ListWorkspaceQuery) (*ListWorkspaceResponse, error) {
	// authorization
	// If required, implement here complex logic like multiple-domains filtering, etc
	u, ok := ctx.Value(ccontext.UserSignupComplaintNameKey).(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// validate query
	// TODO: sanitize input, block reserved labels, etc

	// data access
	ww := restworkspacesv1alpha1.WorkspaceList{}
	opts := &client.ListOptions{Namespace: query.Namespace}
	if err := h.lister.ListUserWorkspaces(ctx, u, &ww, opts); err != nil {
		return nil, err
	}

	// reply
	return &ListWorkspaceResponse{Workspaces: ww}, nil
}
