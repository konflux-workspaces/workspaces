package handlers

import (
	"context"
	"net/http"

	"github.com/filariow/workspaces/server/query"
	"github.com/filariow/workspaces/server/rest/marshal"
)

var _ http.Handler = &ReadWorkspaceHandler{}

// handler dependencies
type ReadWorkspaceMapperFunc func(prefix string, r *http.Request) (*query.ReadWorkspaceQuery, error)
type ReadWorkspaceQueryHandlerFunc func(context.Context, query.ReadWorkspaceQuery) (*query.ReadWorkspaceResponse, error)

// ReadWorkspaceHandler the http.Request handler for Read Workspaces endpoint
type ReadWorkspaceHandler struct {
	Prefix string

	MapperFunc   ReadWorkspaceMapperFunc
	QueryHandler ReadWorkspaceQueryHandlerFunc

	Marshal   marshal.MarshalFunc
	Unmarshal marshal.UnmarshalFunc
}

// NewReadWorkspaceHandler creates a ReadWorkspaceHandler
func NewReadWorkspaceHandler(
	prefix string,
	mapperFunc ReadWorkspaceMapperFunc,
	queryHandler ReadWorkspaceQueryHandlerFunc,
	marshalFunc marshal.MarshalFunc,
	unmarshalFunc marshal.UnmarshalFunc,
) *ReadWorkspaceHandler {
	return &ReadWorkspaceHandler{
		Prefix:       prefix,
		MapperFunc:   mapperFunc,
		QueryHandler: queryHandler,
		Marshal:      marshalFunc,
		Unmarshal:    unmarshalFunc,
	}
}

func (h *ReadWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// map
	q, err := h.MapperFunc(h.Prefix, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// execute
	qr, err := h.QueryHandler(r.Context(), *q)
	if err != nil {
		// TODO: switch on error type
		w.WriteHeader(http.StatusInternalServerError)
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
