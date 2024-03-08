package workspace

import (
	"context"
	"errors"
	"net/http"

	"github.com/konflux-workspaces/workspaces/server/core"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
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
	l := log.FromContext(r.Context())
	l.Debug("executing read")

	// build marshaler for the given request
	l.Debug("building marshaler for request")
	m, err := h.MarshalerProvider(r)
	if err != nil {
		l.Error("error building marshaler for request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map
	l.Debug("mapping request to read query")
	q, err := h.MapperFunc(r)
	if err != nil {
		l.Error("error mapping request to read query", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// execute
	l.Debug("executing create query", "query", q)
	qr, err := h.QueryHandler(r.Context(), *q)
	if err != nil {
		l.Error("error executing read query", "error", err)
		switch {
		case errors.Is(err, core.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// marshal response
	l.Debug("marshaling response", "query", qr)
	d, err := m.Marshal(qr.Workspace)
	if err != nil {
		l.Error("error handling command", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// reply
	l.Debug("writing response", "response", d)
	w.Header().Add(header.ContentType, m.ContentType())
	if _, err := w.Write(d); err != nil {
		l.Error("error writing response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func MapReadWorkspaceHttp(r *http.Request) (*workspace.ReadWorkspaceQuery, error) {
	c := r.PathValue("name")
	ns := r.PathValue("namespace")
	return &workspace.ReadWorkspaceQuery{Name: c, Owner: ns}, nil
}
