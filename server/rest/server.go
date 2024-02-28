package rest

import (
	"fmt"
	"log"
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
		Handler: buildServerHandler(readHandle, listHandle),
	}
}

func buildServerHandler(
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
) http.HandlerFunc {
	mux := http.NewServeMux()
	addHealthz(mux)
	addWorkspaces(mux, readHandle, listHandle)

	h := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		mux.ServeHTTP(w, r)
	}
	return h
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

func addHealthz(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alive"))
	})
}
