package workspace

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/konflux-workspaces/workspaces/server/core"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/rest/header"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
	"k8s.io/apimachinery/pkg/types"
)

var (
	_ http.Handler = &PatchWorkspaceHandler{}

	_ PatchWorkspaceMapperFunc = MapPatchWorkspaceHttp
)

// handler dependencies
type PatchWorkspaceMapperFunc func(*http.Request) (*workspace.PatchWorkspaceCommand, error)
type PatchWorkspaceCommandHandlerFunc func(context.Context, workspace.PatchWorkspaceCommand) (*workspace.PatchWorkspaceResponse, error)

// PatchWorkspaceHandler the http.Request handler for Patch Workspaces endpoint
type PatchWorkspaceHandler struct {
	MapperFunc     PatchWorkspaceMapperFunc
	CommandHandler PatchWorkspaceCommandHandlerFunc

	MarshalerProvider marshal.MarshalerProvider
}

// NewPatchWorkspaceHandler creates a PatchWorkspaceHandler
func NewDefaultPatchWorkspaceHandler(
	handler PatchWorkspaceCommandHandlerFunc,
) *PatchWorkspaceHandler {
	return NewPatchWorkspaceHandler(
		MapPatchWorkspaceHttp,
		handler,
		marshal.DefaultMarshalerProvider,
	)
}

// NewPatchWorkspaceHandler creates a PatchWorkspaceHandler
func NewPatchWorkspaceHandler(
	mapperFunc PatchWorkspaceMapperFunc,
	commandHandler PatchWorkspaceCommandHandlerFunc,
	marshalerProvider marshal.MarshalerProvider,
) *PatchWorkspaceHandler {
	return &PatchWorkspaceHandler{
		MapperFunc:        mapperFunc,
		CommandHandler:    commandHandler,
		MarshalerProvider: marshalerProvider,
	}
}

func (h *PatchWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := log.FromContext(r.Context())
	l.Debug("executing patch")

	// build marshaler for the given request
	l.Debug("building marshaler for request")
	m, err := h.MarshalerProvider(r)
	if err != nil {
		l.Debug("error building marshaler for request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map
	l.Debug("mapping request to patch command")
	c, err := h.MapperFunc(r)
	if err != nil {
		l.Debug("error mapping request to command", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// execute
	l.Debug("executing patch command", "command", c)
	cr, err := h.CommandHandler(r.Context(), *c)
	if err != nil {
		l = l.With("error", err)
		switch {
		case errors.Is(err, core.ErrNotFound):
			l.Debug("error executing patch command: resource not found")
			w.WriteHeader(http.StatusNotFound)
		default:
			l.Error("error executing patch command")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// marshal response
	l.Debug("marshaling response", "response", &cr)
	d, err := m.Marshal(cr.Workspace)
	if err != nil {
		l.Error("unexpected error marshaling response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// reply
	l.Debug("writing response", "response", d)
	w.Header().Add(header.ContentType, m.ContentType())
	if _, err := w.Write(d); err != nil {
		l.Error("unexpected error writing response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func MapPatchWorkspaceHttp(r *http.Request) (*workspace.PatchWorkspaceCommand, error) {
	// parse request body
	d, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}

	ct, ok := r.Header["Content-Type"]
	if !ok || len(ct) != 1 {
		return nil, fmt.Errorf("Content-Type header is required")
	}

	// retrieve namespace from path
	n := r.PathValue("name")
	ns := r.PathValue("namespace")

	// build command
	return &workspace.PatchWorkspaceCommand{
		Workspace: n,
		Owner:     ns,
		PatchType: types.PatchType(ct[0]),
		Patch:     d,
	}, nil
}
