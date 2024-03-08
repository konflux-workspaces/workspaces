package workspace

import (
	"context"
	"io"
	"net/http"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/rest/header"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
)

type PostWorkspaceMapperFunc func(*http.Request, marshal.UnmarshalerProvider) (*workspace.CreateWorkspaceCommand, error)
type CreateWorkspaceCreateHandlerFunc func(context.Context, workspace.CreateWorkspaceCommand) (*workspace.CreateWorkspaceResponse, error)

type PostWorkspaceHandler struct {
	MapperFunc          PostWorkspaceMapperFunc
	CreateHandler       CreateWorkspaceCreateHandlerFunc
	MarshalerProvider   marshal.MarshalerProvider
	UnmarshalerProvider marshal.UnmarshalerProvider
}

func NewPostWorkspaceHandler(
	mapperFunc PostWorkspaceMapperFunc,
	createHandler CreateWorkspaceCreateHandlerFunc,
	marshalProvider marshal.MarshalerProvider,
	unmarshalProvider marshal.UnmarshalerProvider,
) *PostWorkspaceHandler {
	return &PostWorkspaceHandler{
		MapperFunc:          mapperFunc,
		CreateHandler:       createHandler,
		MarshalerProvider:   marshalProvider,
		UnmarshalerProvider: unmarshalProvider,
	}
}

var _ http.Handler = &PostWorkspaceHandler{}

// ServeHTTP implements http.Handler.
func (p *PostWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  l := log.FromContext(r.Context())

  // build marshaler for the given request
	m, err := p.MarshalerProvider(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// map
	q, err := p.MapperFunc(r, p.UnmarshalerProvider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l.Debug("executing create command", "command", q)

	// execute
	cwr, err := p.CreateHandler(r.Context(), *q)
	if err != nil {
		l.Error("error handling query", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// marshal response
	d, err := m.Marshal(cwr.Workspace)
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

	l.Debug("written data", "data", string(d))
}

func MapPostWorkspaceHttp(r *http.Request, unmarshaler marshal.UnmarshalerProvider) (*workspace.CreateWorkspaceCommand, error) {
	// build unmarshaler for the given request
	u, err := unmarshaler(r)
	if err != nil {
		return nil, err
	}

	// parse request body
	// According to net/http documentation, reading the body of the request
	// should always succeed in server-side request handling.  We include the
	// check here for completeness's sake.
	d, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal body to Workspace
	w := workspacesv1alpha1.Workspace{}
	if err := u.Unmarshal(d, &w); err != nil {
		return nil, err
	}

	// retrieve namespace from path
	ns := r.PathValue("namespace")
	w.SetNamespace(ns)

	// build command
	return &workspace.CreateWorkspaceCommand{
		Workspace: w,
	}, nil
}
