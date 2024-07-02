package workspace

import (
	"context"
	"net/http"

	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/rest/header"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
)

var (
	_ http.Handler = &ListWorkspaceHandler{}

	_ ListWorkspaceMapperFunc = MapListWorkspaceHttp
)

// handler dependencies
type ListWorkspaceMapperFunc func(*http.Request) (*workspace.ListWorkspaceQuery, error)
type ListWorkspaceQueryHandlerFunc func(context.Context, workspace.ListWorkspaceQuery) (*workspace.ListWorkspaceResponse, error)

// ListWorkspaceHandler the http.Request handler for List Workspaces endpoint
type ListWorkspaceHandler struct {
	MapperFunc   ListWorkspaceMapperFunc
	QueryHandler ListWorkspaceQueryHandlerFunc

	MarshalerProvider marshal.MarshalerProvider
}

func NewDefaultListWorkspaceHandler(
	handler ListWorkspaceQueryHandlerFunc,
) *ListWorkspaceHandler {
	return NewListWorkspaceHandler(
		MapListWorkspaceHttp,
		handler,
		marshal.DefaultMarshalerProvider,
	)
}

// NewListWorkspaceHandler creates a ListWorkspaceHandler
func NewListWorkspaceHandler(
	mapperFunc ListWorkspaceMapperFunc,
	queryHandler ListWorkspaceQueryHandlerFunc,
	marshalerProvider marshal.MarshalerProvider,
) *ListWorkspaceHandler {
	return &ListWorkspaceHandler{
		MapperFunc:        mapperFunc,
		QueryHandler:      queryHandler,
		MarshalerProvider: marshalerProvider,
	}
}

func (h *ListWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := log.FromContext(r.Context())
	l.Debug("executing list")

	// build marshaler for the given request
	l.Debug("building marshaler for request")
	m, err := h.MarshalerProvider(r)
	if err != nil {
		l.Error("error building marshaler for request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map to query
	l.Debug("mapping request to list query")
	q, err := h.MapperFunc(r)
	if err != nil {
		l.Error("error mapping request to query", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// execute
	l.Debug("executing create query", "query", q)
	qr, err := h.QueryHandler(r.Context(), *q)
	if err != nil {
		l.Error("error executing list query", "query", q, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// marshal response
	l.Debug("marshaling response", "query", qr)
	d, err := m.Marshal(qr.Workspaces)
	if err != nil {
		l.Error("error marshaling response", "error", err)
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

func MapListWorkspaceHttp(r *http.Request) (*workspace.ListWorkspaceQuery, error) {
	q := workspace.ListWorkspaceQuery{}
	ns := r.PathValue("namespace")
	if ns != "" {
		q.Namespace = ns
	}
	return &q, nil
}
