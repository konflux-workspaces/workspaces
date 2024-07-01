package workspace

import (
	"context"
	"fmt"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -destination=mocks/mocks_create.go -package=mocks . WorkspaceCreator

// CreateWorkspaceCommand contains the information needed to create a new workspace
type CreateWorkspaceCommand struct {
	Workspace restworkspacesv1alpha1.Workspace
}

// CreateWorkspaceResponse contains the newly-created workspace
type CreateWorkspaceResponse struct {
	Workspace *restworkspacesv1alpha1.Workspace
}

type WorkspaceCreator interface {
	CreateUserWorkspace(ctx context.Context, user string, workspace *restworkspacesv1alpha1.Workspace, opts ...client.CreateOption) error
}

type CreateWorkspaceHandler struct {
	creator WorkspaceCreator
}

func NewCreateWorkspaceHandler(creator WorkspaceCreator) *CreateWorkspaceHandler {
	return &CreateWorkspaceHandler{creator: creator}
}

func (h *CreateWorkspaceHandler) Handle(ctx context.Context, request CreateWorkspaceCommand) (*CreateWorkspaceResponse, error) {
	u, ok := ctx.Value(ccontext.UserSignupComplaintNameKey).(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// TODO: validate the workspace; maybe punt to a webhook down the line?

	// write the workspace
	workspace := request.Workspace.DeepCopy()
	opts := &client.CreateOptions{}
	if err := h.creator.CreateUserWorkspace(ctx, u, workspace, opts); err != nil {
		return nil, err
	}

	response := &CreateWorkspaceResponse{
		Workspace: workspace,
	}
	return response, nil
}
