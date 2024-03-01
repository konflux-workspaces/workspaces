package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware/auth"
	"github.com/konflux-workspaces/workspaces/server/rest/workspace"
)

const (
	WorkspacesPrefix           string = `/apis/workspaces.io/v1alpha1/workspaces`
	NamespacedWorkspacesPrefix string = `/apis/workspaces.io/v1alpha1/namespaces/{namespace}/workspaces`
)

func New(
	addr string,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
	createHandle workspace.CreateWorkspaceCreateHandlerFunc,
	updateHandle workspace.UpdateWorkspaceCommandHandlerFunc,
) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           buildServerHandler(readHandle, listHandle, createHandle, updateHandle),
		ReadHeaderTimeout: 3 * time.Second,
	}
}

func buildServerHandler(
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
	createHandle workspace.CreateWorkspaceCreateHandlerFunc,
	updateHandle workspace.UpdateWorkspaceCommandHandlerFunc,
) http.Handler {
	mux := http.NewServeMux()
	addHealthz(mux)
	addWorkspaces(mux, readHandle, listHandle, createHandle, updateHandle)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return middleware.NewLogRequestMiddleware(mux)
}

func addWorkspaces(
	mux *http.ServeMux,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
	postHandle workspace.CreateWorkspaceCreateHandlerFunc,
	updateHandle workspace.UpdateWorkspaceCommandHandlerFunc,
) {
	// Read
	mux.Handle(fmt.Sprintf("GET %s/{name}", NamespacedWorkspacesPrefix),
		auth.NewJwtBearerMiddleware(
			workspace.NewReadWorkspaceHandler(
				workspace.MapReadWorkspaceHttp,
				readHandle,
				marshal.DefaultMarshalerProvider,
			)))

	// List
	lh := auth.NewJwtBearerMiddleware(
		workspace.NewListWorkspaceHandler(
			workspace.MapListWorkspaceHttp,
			listHandle,
			marshal.DefaultMarshalerProvider,
			marshal.DefaultUnmarshalerProvider,
		),
	)
	mux.Handle(fmt.Sprintf("GET %s", WorkspacesPrefix), lh)
	mux.Handle(fmt.Sprintf("GET %s", NamespacedWorkspacesPrefix), lh)

	// Update
	mux.Handle(fmt.Sprintf("PUT %s/{name}", NamespacedWorkspacesPrefix),
		auth.NewJwtBearerMiddleware(
			workspace.NewUpdateWorkspaceHandler(
				workspace.MapUpdateWorkspaceHttp,
				updateHandle,
				marshal.DefaultMarshalerProvider,
				marshal.DefaultUnmarshalerProvider,
			)))

	// Create
	mux.Handle(fmt.Sprintf("POST %s", NamespacedWorkspacesPrefix),
		auth.NewJwtBearerMiddleware(
			workspace.NewPostWorkspaceHandler(
				workspace.MapPostWorkspaceHttp,
				postHandle,
				marshal.DefaultMarshalerProvider,
				marshal.DefaultUnmarshalerProvider,
			)))
}

func addHealthz(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("alive")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
