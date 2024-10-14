package workspace

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	jsonpatch "github.com/evanphx/json-patch/v5"
	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/log"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

// PatchWorkspaceCommand contains the information needed to retrieve a Workspace the user has access to from the data source
type PatchWorkspaceCommand struct {
	Owner     string
	Workspace string
	Patch     []byte
	PatchType types.PatchType
}

// PatchWorkspaceResponse contains the workspace the user requested
type PatchWorkspaceResponse struct {
	Workspace *restworkspacesv1alpha1.Workspace
}

// PatchWorkspaceHandler processes PatchWorkspaceCommand and returns PatchWorkspaceResponse fetching data from a WorkspacePatcher
type PatchWorkspaceHandler struct {
	reader  WorkspaceReader
	updater WorkspaceUpdater
}

// NewPatchWorkspaceHandler creates a new PatchWorkspaceHandler that uses a specified WorkspacePatcher
func NewPatchWorkspaceHandler(reader WorkspaceReader, updater WorkspaceUpdater) *PatchWorkspaceHandler {
	return &PatchWorkspaceHandler{
		reader:  reader,
		updater: updater,
	}
}

// Handle handles a PatchWorkspaceCommand and returns a PatchWorkspaceResponse or an error
func (h *PatchWorkspaceHandler) Handle(ctx context.Context, command PatchWorkspaceCommand) (*PatchWorkspaceResponse, error) {
	// authorization
	// If required, implement here complex logic like multiple-domains filtering, etc
	u, ok := ctx.Value(ccontext.UserSignupComplaintNameKey).(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// validate query
	// TODO: sanitize input, block reserved labels, etc

	// retrieve workspace
	w := workspacesv1alpha1.Workspace{}
	if err := h.reader.ReadUserWorkspace(ctx, u, command.Owner, command.Workspace, &w); err != nil {
		return nil, err
	}

	// apply patch
	pw, err := h.applyPatch(&w, command)
	if err != nil {
		return nil, fmt.Errorf("error patching Workspace %s/%s: %w", command.Owner, command.Workspace, err)
	}

	log.FromContext(ctx).Debug("updating workspace", "workspace", pw)
	opts := &client.UpdateOptions{}
	if err := h.updater.UpdateUserWorkspace(ctx, u, pw, opts); err != nil {
		return nil, err
	}

	// reply
	return &PatchWorkspaceResponse{
		Workspace: pw,
	}, nil
}

func (h *PatchWorkspaceHandler) applyPatch(w *workspacesv1alpha1.Workspace, command PatchWorkspaceCommand) (*workspacesv1alpha1.Workspace, error) {
	switch command.PatchType {
	case types.MergePatchType:
		return h.applyMergePatch(w, command.Patch)
	case types.StrategicMergePatchType:
		return h.applyStrategicMergePatch(w, command.Patch)
	default:
		return nil, fmt.Errorf("unsupported patch type: %s", command.PatchType)
	}
}

func (h *PatchWorkspaceHandler) applyMergePatch(w *workspacesv1alpha1.Workspace, patch []byte) (*workspacesv1alpha1.Workspace, error) {
	// marshal workspace as json
	wj, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}

	// apply jsonpatch
	pwj, err := jsonpatch.MergePatch(wj, patch)
	if err != nil {
		return nil, err
	}

	// unmarshal json to struct
	pw := workspacesv1alpha1.Workspace{}
	if err := json.Unmarshal(pwj, &pw); err != nil {
		return nil, err
	}

	return &pw, nil
}

func (h *PatchWorkspaceHandler) applyStrategicMergePatch(w *workspacesv1alpha1.Workspace, patch []byte) (*workspacesv1alpha1.Workspace, error) {
	// marshal workspace as json
	wj, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}

	// apply jsonpatch
	pwj, err := strategicpatch.StrategicMergePatch(wj, patch, *w)
	if err != nil {
		return nil, err
	}

	// unmarshal json to struct
	pw := workspacesv1alpha1.Workspace{}
	if err := json.Unmarshal(pwj, &pw); err != nil {
		return nil, err
	}

	return &pw, nil
}
