package workspace

import "context"

type ListWorkspaceQuery struct{}

type ListWorkspaceResponse struct {
	Workspaces struct{}
}

func ListWorkspaceHandler(ctx context.Context, query ListWorkspaceQuery) (*ListWorkspaceResponse, error) {
	// TODO:
	panic("to be implemented")
}
