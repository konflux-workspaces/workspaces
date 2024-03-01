package workspace

import (
	"context"
	"fmt"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReadWorkspaceQuery contains the information needed to retrieve a Workspace the user has access to from the data source
type ReadWorkspaceQuery struct {
	Name  string
	Owner string
}

// ReadWorkspaceResponse contains the workspace the user requested
type ReadWorkspaceResponse struct {
	Workspace *workspacesv1alpha1.Workspace
}

// WorkspaceReader is the interface the data source needs to implement to allow the ReadWorkspaceHandler to fetch data from it
type WorkspaceReader interface {
	ReadUserWorkspace(ctx context.Context, user, owner, space string, obj *workspacesv1alpha1.Workspace, opts ...client.GetOption) error
}

// ReadWorkspaceHandler processes ReadWorkspaceQuery and returns ReadWorkspaceResponse fetching data from a WorkspaceReader
type ReadWorkspaceHandler struct {
	reader WorkspaceReader
}

// NewReadWorkspaceHandler creates a new ReadWorkspaceHandler that uses a specified WorkspaceReader
func NewReadWorkspaceHandler(reader WorkspaceReader) *ReadWorkspaceHandler {
	return &ReadWorkspaceHandler{reader: reader}
}

// Handle handles a ReadWorkspaceQuery and returns a ReadWorkspaceResponse or an error
func (h *ReadWorkspaceHandler) Handle(ctx context.Context, query ReadWorkspaceQuery) (*ReadWorkspaceResponse, error) {
	// authorization
	// If required, implement here complex logic like multiple-domains filtering, etc
	u, ok := ctx.Value("user").(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// validate query
	// TODO: sanitize input, block reserved labels, etc

	// data access
	var w workspacesv1alpha1.Workspace
	if err := h.reader.ReadUserWorkspace(ctx, u, query.Owner, query.Name, &w); err != nil {
		return nil, err
	}

	// reply
	return &ReadWorkspaceResponse{
		Workspace: &w,
	}, nil
}
