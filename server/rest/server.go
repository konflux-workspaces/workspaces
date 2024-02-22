package rest

import (
	"fmt"
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
	addWorkspaces(mux)
	return mux
}

func addWorkspaces(mux *http.ServeMux) {
	// Read
	mux.Handle(fmt.Sprintf("GET %s/{name}", WorkspacesPrefix),
		workspace.NewReadWorkspaceHandler(
			workspace.MapReadWorkspaceHttp,
			cw.ReadWorkspaceHandler,
			marshal.DefaultMarshal,
		))

	// List
	mux.Handle(fmt.Sprintf("GET %s", WorkspacesPrefix),
		workspace.NewListWorkspaceHandler(
			workspace.MapListWorkspaceHttp,
			cw.ListWorkspaceHandler,
			marshal.DefaultMarshal,
			marshal.DefaultUnmarshal,
		))
}
