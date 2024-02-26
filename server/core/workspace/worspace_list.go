package workspace

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// ListWorkspaceQuery contains the information needed to retrieve all the workspaces the user has access to from the data source
type ListWorkspaceQuery struct{}

// ListWorkspaceResponse contains all the workspaces the user can access
type ListWorkspaceResponse struct {
	Workspaces workspacesv1alpha1.WorkspaceList
}

// WorkspaceLister is the interface the data source needs to implement to allow the ListWorkspaceHandler to fetch data from it
type WorkspaceLister interface {
	ListUserWorkspaces(ctx context.Context, user string, objs *workspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error
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
	// TODO: disable unauthenticated access in the HTTP Server
	// If required, implement here complex logic like multiple-domains filtering, etc
	u, ok := ctx.Value("user").(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// validate query
	// TODO: sanitize input, block reserved labels, etc

	// data access
	ww := workspacesv1alpha1.WorkspaceList{}
	if err := h.lister.ListUserWorkspaces(ctx, u, &ww, &client.ListOptions{}); err != nil {
		return nil, err
	}

	// reply
	return &ListWorkspaceResponse{Workspaces: ww}, nil
}
