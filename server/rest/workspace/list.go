package workspace

import (
	"context"
	"log"
	"net/http"

	"github.com/filariow/workspaces/server/core/workspace"
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

	Marshaler marshal.Marshaler
	Unmarshal marshal.Unmarshaler
}

func NewDefaultListWorkspaceHandler(
	handler ListWorkspaceQueryHandlerFunc,
) *ListWorkspaceHandler {
	return NewListWorkspaceHandler(
		MapListWorkspaceHttp,
		handler,
		marshal.DefaultMarshal,
		marshal.DefaultUnmarshal,
	)
}

// NewListWorkspaceHandler creates a ListWorkspaceHandler
func NewListWorkspaceHandler(
	mapperFunc ListWorkspaceMapperFunc,
	queryHandler ListWorkspaceQueryHandlerFunc,
	marshaler marshal.Marshaler,
	unmarshaler marshal.Unmarshaler,
) *ListWorkspaceHandler {
	return &ListWorkspaceHandler{
		MapperFunc:   mapperFunc,
		QueryHandler: queryHandler,
		Marshaler:    marshaler,
		Unmarshal:    unmarshaler,
	}
}

func (h *ListWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	d, err := h.Marshaler.Marshal(qr.Workspaces)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// reply
	w.Header().Add("Content-Type", h.Marshaler.ContentType())
	if _, err := w.Write(d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("written: %s", string(d))
}

func MapListWorkspaceHttp(r *http.Request) (*workspace.ListWorkspaceQuery, error) {
	return &workspace.ListWorkspaceQuery{}, nil
}
