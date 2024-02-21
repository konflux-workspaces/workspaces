package rest

import (
	"net/http"

	cw "github.com/filariow/workspaces/server/core/workspace"
	"github.com/filariow/workspaces/server/rest/marshal"
	"github.com/filariow/workspaces/server/rest/workspace"
)

const WorkspacesPrefix string = "/workspaces"

func New(addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: buildServerMux(),
	}
}

func buildServerMux() *http.ServeMux {
	mux := http.NewServeMux()
	addWorkspaces(mux, WorkspacesPrefix)
	return mux
}

func addWorkspaces(mux *http.ServeMux, prefix string) {
	workspace.AddWorkspaces(
    mux,
    prefix,
		nil,
		workspace.NewReadWorkspaceHandler(
			prefix,
			workspace.MapReadWorkspaceHttp,
			cw.ReadWorkspaceHandler,
			marshal.DefaultMarshal,
		),
		workspace.NewListWorkspaceHandler(
			workspace.MapListWorkspaceHttp,
			cw.ListWorkspaceHandler,
			marshal.DefaultMarshal,
			marshal.DefaultUnmarshal,
		),
		nil,
		nil,
	)
}
