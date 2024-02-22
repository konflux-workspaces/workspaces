package workspace

import (
	"context"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

type ReadWorkspaceQuery struct {
	Name string
}

type ReadWorkspaceResponse struct {
	Workspace workspacesv1alpha1.Workspace
}

type GetWorkspaceFunc func(ctx context.Context, name string, obj *workspacesv1alpha1.Workspace) error

type ReadWorkspaceHandler struct {
	GetWorkspace GetWorkspaceFunc
}

func (h *ReadWorkspaceHandler) Handle(ctx context.Context, query ReadWorkspaceQuery) (*ReadWorkspaceResponse, error) {
	// Authorization

	// Database access
	var w workspacesv1alpha1.Workspace
	if err := h.GetWorkspace(ctx, query.Name, &w); err != nil {
		return nil, err
	}

	// Reply
	return &ReadWorkspaceResponse{
		Workspace: w,
	}, nil
}
