package workspace

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/filariow/workspaces/server/core"
	"github.com/filariow/workspaces/server/core/workspace"
	"github.com/filariow/workspaces/server/rest/marshal"
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

	Marshal marshal.MarshalFunc
}

// NewReadWorkspaceHandler creates a ReadWorkspaceHandler
func NewDefaultReadWorkspaceHandler(
	handler ReadWorkspaceQueryHandlerFunc,
) *ReadWorkspaceHandler {
	return NewReadWorkspaceHandler(
		MapReadWorkspaceHttp,
		handler,
		marshal.DefaultMarshal,
	)
}

// NewReadWorkspaceHandler creates a ReadWorkspaceHandler
func NewReadWorkspaceHandler(
	mapperFunc ReadWorkspaceMapperFunc,
	queryHandler ReadWorkspaceQueryHandlerFunc,
	marshalFunc marshal.MarshalFunc,
) *ReadWorkspaceHandler {
	return &ReadWorkspaceHandler{
		MapperFunc:   mapperFunc,
		QueryHandler: queryHandler,
		Marshal:      marshalFunc,
	}
}

func (h *ReadWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	d, err := h.Marshal(qr.Workspace)
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

func MapReadWorkspaceHttp(r *http.Request) (*workspace.ReadWorkspaceQuery, error) {
	c := r.PathValue("name")
	return &workspace.ReadWorkspaceQuery{Name: c}, nil
}
