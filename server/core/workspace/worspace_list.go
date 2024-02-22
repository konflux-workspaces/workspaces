package workspace

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

type ListWorkspaceQuery struct{}

type ListWorkspaceResponse struct {
	Workspaces workspacesv1alpha1.WorkspaceList
}

type WorkspaceLister interface {
	ListUserWorkspaces(ctx context.Context, user string, objs *workspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error
}

type ListWorkspaceHandler struct {
	WorkspaceLister
}

func (h *ListWorkspaceHandler) Handle(ctx context.Context, query ListWorkspaceQuery) (*ListWorkspaceResponse, error) {
	// Authorization
	// TODO: disable unauthenticated access in the HTTP Server
	// Eventually, implement here complex logic like multiple-domains filtering, etc
	u, ok := ctx.Value("user").(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// validate query
	// TODO: sanitize input, block reserved labels, etc

	// retrieves workspaces from data source
	ww := workspacesv1alpha1.WorkspaceList{}
	if err := h.ListUserWorkspaces(ctx, u, &ww, &client.ListOptions{}); err != nil {
		return nil, err
	}

	// reply
	return &ListWorkspaceResponse{Workspaces: ww}, nil
}
