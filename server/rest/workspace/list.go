package workspace

import (
	"context"
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

	Marshal   marshal.MarshalFunc
	Unmarshal marshal.UnmarshalFunc
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
	marshalFunc marshal.MarshalFunc,
	unmarshalFunc marshal.UnmarshalFunc,
) *ListWorkspaceHandler {
	return &ListWorkspaceHandler{
		MapperFunc:   mapperFunc,
		QueryHandler: queryHandler,
		Marshal:      marshalFunc,
		Unmarshal:    unmarshalFunc,
	}
}

func (h *ListWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// map
	q, err := h.MapperFunc(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// execute
	qr, err := h.QueryHandler(r.Context(), *q)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// marshal response
	d, err := h.Marshal(qr.Workspaces)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// reply
	if _, err := w.Write(d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func MapListWorkspaceHttp(r *http.Request) (*workspace.ListWorkspaceQuery, error) {
	return &workspace.ListWorkspaceQuery{}, nil
}
