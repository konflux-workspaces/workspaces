package workspace

import (
	"context"
	"log"
	"net/http"

	"github.com/filariow/workspaces/server/core/workspace"
	"github.com/filariow/workspaces/server/rest/header"
	"github.com/filariow/workspaces/server/rest/marshal"
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
	UnmarshalProvider marshal.UnmarshalerProvider
}

func NewDefaultListWorkspaceHandler(
	handler ListWorkspaceQueryHandlerFunc,
) *ListWorkspaceHandler {
	return NewListWorkspaceHandler(
		MapListWorkspaceHttp,
		handler,
		marshal.DefaultMarshalerProvider,
		marshal.DefaultUnmarshalerProvider,
	)
}

// NewListWorkspaceHandler creates a ListWorkspaceHandler
func NewListWorkspaceHandler(
	mapperFunc ListWorkspaceMapperFunc,
	queryHandler ListWorkspaceQueryHandlerFunc,
	marshalerProvider marshal.MarshalerProvider,
	unmarshalerProvider marshal.UnmarshalerProvider,
) *ListWorkspaceHandler {
	return &ListWorkspaceHandler{
		MapperFunc:        mapperFunc,
		QueryHandler:      queryHandler,
		MarshalerProvider: marshalerProvider,
		UnmarshalProvider: unmarshalerProvider,
	}
}

func (h *ListWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// build marshaler for the given request
	m, err := h.MarshalerProvider(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map to query
	q, err := h.MapperFunc(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("executing list query %v", q)

	// execute
	qr, err := h.QueryHandler(r.Context(), *q)
	if err != nil {
		log.Printf("error handling query: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// marshal response
	d, err := m.Marshal(qr.Workspaces)
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

	log.Printf("written: %s", string(d))
}

func MapListWorkspaceHttp(r *http.Request) (*workspace.ListWorkspaceQuery, error) {
	q := workspace.ListWorkspaceQuery{}
	ns := r.PathValue("namespace")
	if ns != "" {
		q.Namespace = ns
	}
	return &q, nil
}
