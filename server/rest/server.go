package rest

import (
	"fmt"
	"net/http"

	"github.com/filariow/workspaces/server/rest/marshal"
	"github.com/filariow/workspaces/server/rest/workspace"
)

const WorkspacesPrefix string = "/workspaces"

func New(
	addr string,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: buildServerMux(readHandle, listHandle),
	}
}

func buildServerMux(
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
) *http.ServeMux {
	mux := http.NewServeMux()
	addWorkspaces(mux, readHandle, listHandle)
	return mux
}

func addWorkspaces(
	mux *http.ServeMux,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
) {
	// Read
	mux.Handle(fmt.Sprintf("GET %s/{name}", WorkspacesPrefix),
		workspace.NewReadWorkspaceHandler(
			workspace.MapReadWorkspaceHttp,
			readHandle,
			marshal.DefaultMarshal,
		))

	// List
	mux.Handle(fmt.Sprintf("GET %s", WorkspacesPrefix),
		workspace.NewListWorkspaceHandler(
			workspace.MapListWorkspaceHttp,
			listHandle,
			marshal.DefaultMarshal,
			marshal.DefaultUnmarshal,
		))
}
