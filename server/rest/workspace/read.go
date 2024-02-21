package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/filariow/workspaces/server/core"
	"github.com/filariow/workspaces/server/core/workspace"
	"github.com/filariow/workspaces/server/rest/marshal"
)

var (
	_ http.Handler = &ReadWorkspaceHandler{}

	_ ReadWorkspaceMapperFunc = MapReadWorkspaceHttp
)

// handler dependencies
type ReadWorkspaceMapperFunc func(prefix string, r *http.Request) (*workspace.ReadWorkspaceQuery, error)
type ReadWorkspaceQueryHandlerFunc func(context.Context, workspace.ReadWorkspaceQuery) (*workspace.ReadWorkspaceResponse, error)

// ReadWorkspaceHandler the http.Request handler for Read Workspaces endpoint
type ReadWorkspaceHandler struct {
	Prefix string

	MapperFunc   ReadWorkspaceMapperFunc
	QueryHandler ReadWorkspaceQueryHandlerFunc

	Marshal   marshal.MarshalFunc
}

// NewReadWorkspaceHandler creates a ReadWorkspaceHandler
func NewDefaultReadWorkspaceHandler(
	prefix string,
	handler ReadWorkspaceQueryHandlerFunc,
) *ReadWorkspaceHandler {
	return NewReadWorkspaceHandler(
		prefix,
		MapReadWorkspaceHttp,
		handler,
		marshal.DefaultMarshal,
	)
}

// NewReadWorkspaceHandler creates a ReadWorkspaceHandler
func NewReadWorkspaceHandler(
	prefix string,
	mapperFunc ReadWorkspaceMapperFunc,
	queryHandler ReadWorkspaceQueryHandlerFunc,
	marshalFunc marshal.MarshalFunc,
) *ReadWorkspaceHandler {
	return &ReadWorkspaceHandler{
		Prefix:       prefix,
		MapperFunc:   mapperFunc,
		QueryHandler: queryHandler,
		Marshal:      marshalFunc,
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

func MapReadWorkspaceHttp(prefix string, r *http.Request) (*workspace.ReadWorkspaceQuery, error) {
	c, ok := strings.CutPrefix(r.URL.Path, prefix)
	if !ok {
		return nil, fmt.Errorf("")
	}
	c = strings.TrimLeft(c, "/")

	if strings.ContainsRune(c, '/') {
		// TODO: not found
		return nil, fmt.Errorf("")
	}

	return &workspace.ReadWorkspaceQuery{Name: c}, nil
}
