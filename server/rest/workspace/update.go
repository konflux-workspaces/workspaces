package workspace

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/konflux-workspaces/workspaces/server/core"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/rest/header"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

var (
	_ http.Handler = &UpdateWorkspaceHandler{}

	_ UpdateWorkspaceMapperFunc = MapUpdateWorkspaceHttp
)

// handler dependencies
type UpdateWorkspaceMapperFunc func(*http.Request, marshal.UnmarshalerProvider) (*workspace.UpdateWorkspaceCommand, error)
type UpdateWorkspaceCommandHandlerFunc func(context.Context, workspace.UpdateWorkspaceCommand) (*workspace.UpdateWorkspaceResponse, error)

// UpdateWorkspaceHandler the http.Request handler for Update Workspaces endpoint
type UpdateWorkspaceHandler struct {
	MapperFunc     UpdateWorkspaceMapperFunc
	CommandHandler UpdateWorkspaceCommandHandlerFunc

	MarshalerProvider   marshal.MarshalerProvider
	UnmarshalerProvider marshal.UnmarshalerProvider
}

// NewUpdateWorkspaceHandler creates a UpdateWorkspaceHandler
func NewDefaultUpdateWorkspaceHandler(
	handler UpdateWorkspaceCommandHandlerFunc,
) *UpdateWorkspaceHandler {
	return NewUpdateWorkspaceHandler(
		MapUpdateWorkspaceHttp,
		handler,
		marshal.DefaultMarshalerProvider,
		marshal.DefaultUnmarshalerProvider,
	)
}

// NewUpdateWorkspaceHandler creates a UpdateWorkspaceHandler
func NewUpdateWorkspaceHandler(
	mapperFunc UpdateWorkspaceMapperFunc,
	queryHandler UpdateWorkspaceCommandHandlerFunc,
	marshalerProvider marshal.MarshalerProvider,
	unmarshalerProvider marshal.UnmarshalerProvider,
) *UpdateWorkspaceHandler {
	return &UpdateWorkspaceHandler{
		MapperFunc:          mapperFunc,
		CommandHandler:      queryHandler,
		MarshalerProvider:   marshalerProvider,
		UnmarshalerProvider: unmarshalerProvider,
	}
}

func (h *UpdateWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// build marshaler for the given request
	m, err := h.MarshalerProvider(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map
	q, err := h.MapperFunc(r, h.UnmarshalerProvider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("executing update query %+v", q)

	// execute
	qr, err := h.CommandHandler(r.Context(), *q)
	if err != nil {
		log.Printf("error executing update command: %v", err)
		switch {
		case errors.Is(err, core.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// marshal response
	d, err := m.Marshal(qr.Workspace)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// reply
	w.Header().Add(header.ContentType, m.ContentType())
	if _, err := w.Write(d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func MapUpdateWorkspaceHttp(r *http.Request, provider marshal.UnmarshalerProvider) (*workspace.UpdateWorkspaceCommand, error) {
	// build unmarshaler for the given request
	u, err := provider(r)
	if err != nil {
		return nil, err
	}

	// parse request body
	d, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal body to Workspace
	w := workspacesv1alpha1.Workspace{}
	if err := u.Unmarshal(d, &w); err != nil {
		return nil, err
	}

	// retrieve namespace from path
	n := r.PathValue("name")
	ns := r.PathValue("namespace")

	w.SetName(n)
	w.SetNamespace(ns)

	// build command
	return &workspace.UpdateWorkspaceCommand{
		Workspace: w,
		Owner:     ns,
	}, nil
}
