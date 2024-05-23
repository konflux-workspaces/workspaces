package workspace

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/log"
)

// UpdateWorkspaceCommand contains the information needed to retrieve a Workspace the user has access to from the data source
type UpdateWorkspaceCommand struct {
	Owner     string
	Workspace restworkspacesv1alpha1.Workspace
}

// UpdateWorkspaceResponse contains the workspace the user requested
type UpdateWorkspaceResponse struct {
	Workspace *restworkspacesv1alpha1.Workspace
}

// WorkspaceUpdater is the interface the data source needs to implement to allow the UpdateWorkspaceHandler to fetch data from it
type WorkspaceUpdater interface {
	UpdateUserWorkspace(ctx context.Context, user string, obj *restworkspacesv1alpha1.Workspace, opts ...client.UpdateOption) error
}

// UpdateWorkspaceHandler processes UpdateWorkspaceCommand and returns UpdateWorkspaceResponse fetching data from a WorkspaceUpdater
type UpdateWorkspaceHandler struct {
	updater WorkspaceUpdater
}

// NewUpdateWorkspaceHandler creates a new UpdateWorkspaceHandler that uses a specified WorkspaceUpdater
func NewUpdateWorkspaceHandler(updater WorkspaceUpdater) *UpdateWorkspaceHandler {
	return &UpdateWorkspaceHandler{updater: updater}
}

// Handle handles a UpdateWorkspaceCommand and returns a UpdateWorkspaceResponse or an error
func (h *UpdateWorkspaceHandler) Handle(ctx context.Context, query UpdateWorkspaceCommand) (*UpdateWorkspaceResponse, error) {
	// authorization
	// If required, implement here complex logic like multiple-domains filtering, etc
	u, ok := ctx.Value(ccontext.UserKey).(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// validate query
	// TODO: sanitize input, block reserved labels, etc

	// data access
	w := query.Workspace.DeepCopy()
	log.FromContext(ctx).Debug("updating workspace", "workspace", w)
	opts := &client.UpdateOptions{}
	if err := h.updater.UpdateUserWorkspace(ctx, u, w, opts); err != nil {
		return nil, err
	}

	// reply
	return &UpdateWorkspaceResponse{
		Workspace: w,
	}, nil
}
