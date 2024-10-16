package workspace

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/konflux-workspaces/workspaces/server/core"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/rest/header"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"

	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var (
	_ http.Handler = &UpdateWorkspaceHandler{}

	_ UpdateWorkspaceMapperFunc = MapPutWorkspaceHttp
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
		MapPutWorkspaceHttp,
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
	l := log.FromContext(r.Context())
	l.Debug("executing update")

	// build marshaler for the given request
	l.Debug("building marshaler for request")
	m, err := h.MarshalerProvider(r)
	if err != nil {
		l.Debug("error building marshaler for request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map
	l.Debug("mapping request to update command")
	c, err := h.MapperFunc(r, h.UnmarshalerProvider)
	if err != nil {
		l.Debug("error mapping request to command", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// execute
	l.Debug("executing update command", "command", c)
	cr, err := h.CommandHandler(r.Context(), *c)
	if err != nil {
		l = l.With("error", err)

		switch {
		case errors.Is(err, core.ErrNotFound):
			l.Debug("error executing update command: resource not found")
			w.WriteHeader(http.StatusNotFound)
		case kerrors.IsForbidden(err):
			serr := new(kerrors.StatusError)
			errors.As(err, &serr)
			w.WriteHeader(int(serr.Status().Code))
			if _, err := w.Write([]byte(serr.Error())); err != nil {
				l.Info("error writing response", "error", err)
			}
		default:
			l.Error("error executing update command")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// marshal response
	l.Debug("marshaling response", "response", &cr)
	d, err := m.Marshal(cr.Workspace)
	if err != nil {
		l.Error("unexpected error marshaling response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// reply
	l.Debug("writing response", "response", d)
	w.Header().Add(header.ContentType, m.ContentType())
	if _, err := w.Write(d); err != nil {
		l.Error("unexpected error writing response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func MapPutWorkspaceHttp(r *http.Request, provider marshal.UnmarshalerProvider) (*workspace.UpdateWorkspaceCommand, error) {
	// build unmarshaler for the given request
	u, err := provider(r)
	if err != nil {
		return nil, fmt.Errorf("error building unmarshaler body: %w", err)
	}

	// parse request body
	d, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}

	// unmarshal body to Workspace
	w := restworkspacesv1alpha1.Workspace{}
	if err := u.Unmarshal(d, &w); err != nil {
		return nil, fmt.Errorf("error unmarshaling request body: %w", err)
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
