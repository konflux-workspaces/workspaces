package workspace

import (
	"context"
)

type ReadWorkspaceQuery struct {
	Name string
}

type ReadWorkspaceResponse struct {
	Workspace struct{}
}

func ReadWorkspaceHandler(ctx context.Context, query ReadWorkspaceQuery) (*ReadWorkspaceResponse, error) {
	// TODO:
	panic("to be implemented")
}
