package workspace

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/konflux-workspaces/workspaces/server/core"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/rest/header"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
)

var (
	_ http.Handler = &ReadWorkspaceHandler{}

	_ ReadWorkspaceMapperFunc = MapReadWorkspaceHttp
)

// handler dependencies
type ReadWorkspaceMapperFunc func(r *http.Request) (*workspace.ReadWorkspaceQuery, error)
type ReadWorkspaceQueryHandlerFunc func(context.Context, workspace.ReadWorkspaceQuery) (*workspace.ReadWorkspaceResponse, error)

// ReadWorkspaceHandler the http.Request handler for Read Workspaces endpoint
type ReadWorkspaceHandler struct {
	MapperFunc   ReadWorkspaceMapperFunc
	QueryHandler ReadWorkspaceQueryHandlerFunc

	MarshalerProvider marshal.MarshalerProvider
}

// NewReadWorkspaceHandler creates a ReadWorkspaceHandler
func NewDefaultReadWorkspaceHandler(
	handler ReadWorkspaceQueryHandlerFunc,
) *ReadWorkspaceHandler {
	return NewReadWorkspaceHandler(
		MapReadWorkspaceHttp,
		handler,
		marshal.DefaultMarshalerProvider,
	)
}

// NewReadWorkspaceHandler creates a ReadWorkspaceHandler
func NewReadWorkspaceHandler(
	mapperFunc ReadWorkspaceMapperFunc,
	queryHandler ReadWorkspaceQueryHandlerFunc,
	marshalerProvider marshal.MarshalerProvider,
) *ReadWorkspaceHandler {
	return &ReadWorkspaceHandler{
		MapperFunc:        mapperFunc,
		QueryHandler:      queryHandler,
		MarshalerProvider: marshalerProvider,
	}
}

func (h *ReadWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// build marshaler for the given request
	m, err := h.MarshalerProvider(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map
	q, err := h.MapperFunc(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("executing read query %v", q)

	// execute
	qr, err := h.QueryHandler(r.Context(), *q)
	if err != nil {
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

func MapReadWorkspaceHttp(r *http.Request) (*workspace.ReadWorkspaceQuery, error) {
	c := r.PathValue("name")
	ns := r.PathValue("namespace")
	return &workspace.ReadWorkspaceQuery{Name: c, Owner: ns}, nil
}
