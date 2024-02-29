package rest

import (
	"fmt"
	"net/http"

	"github.com/filariow/workspaces/server/rest/marshal"
	"github.com/filariow/workspaces/server/rest/middleware"
	"github.com/filariow/workspaces/server/rest/middleware/auth"
	"github.com/filariow/workspaces/server/rest/workspace"
)

const WorkspacesPrefix string = "/apis/workspaces.io/v1alpha1/workspaces"

func New(
	addr string,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: buildServerHandler(readHandle, listHandle),
	}
}

func buildServerHandler(
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
) http.Handler {
	mux := http.NewServeMux()
	addHealthz(mux)
	addWorkspaces(mux, readHandle, listHandle)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return middleware.NewLogRequestMiddleware(mux)
}

func addWorkspaces(
	mux *http.ServeMux,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
) {
	// Read
	mux.Handle(fmt.Sprintf("GET %s/{name}", WorkspacesPrefix),
		auth.NewJwtBearerMiddleware(
			workspace.NewReadWorkspaceHandler(
				workspace.MapReadWorkspaceHttp,
				readHandle,
				marshal.DefaultMarshal,
			)))

	// List
	mux.Handle(fmt.Sprintf("GET %s", WorkspacesPrefix),
		auth.NewJwtBearerMiddleware(
			workspace.NewListWorkspaceHandler(
				workspace.MapListWorkspaceHttp,
				listHandle,
				marshal.DefaultMarshal,
				marshal.DefaultUnmarshal,
			)))
}

func addHealthz(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alive"))
	})
}
