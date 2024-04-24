package rest

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware"
	"github.com/konflux-workspaces/workspaces/server/rest/workspace"
)

const (
	WorkspacesPrefix           string = `/apis/workspaces.konflux.io/v1alpha1/internalworkspaces`
	NamespacedWorkspacesPrefix string = `/apis/workspaces.konflux.io/v1alpha1/namespaces/{namespace}/internalworkspaces`
)

func New(
	logger *slog.Logger,
	addr string,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
	createHandle workspace.CreateWorkspaceCreateHandlerFunc,
	updateHandle workspace.UpdateWorkspaceCommandHandlerFunc,
) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           buildServerHandler(logger, readHandle, listHandle, createHandle, updateHandle),
		ReadHeaderTimeout: 3 * time.Second,
	}
}

func buildServerHandler(
	logger *slog.Logger,
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

	if logger == nil {
		return mux
	}

	return middleware.NewLoggerInjectorMiddleware(logger,
		middleware.NewRequestLoggerMiddleware(mux),
	)
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
		withAuthHeaderInfo(
			workspace.NewReadWorkspaceHandler(
				workspace.MapReadWorkspaceHttp,
				readHandle,
				marshal.DefaultMarshalerProvider,
			)))

	// List
	lh := withAuthHeaderInfo(
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
		withAuthHeaderInfo(
			workspace.NewUpdateWorkspaceHandler(
				workspace.MapUpdateWorkspaceHttp,
				updateHandle,
				marshal.DefaultMarshalerProvider,
				marshal.DefaultUnmarshalerProvider,
			)))

	// Create
	mux.Handle(fmt.Sprintf("POST %s", NamespacedWorkspacesPrefix),
		withAuthHeaderInfo(
			workspace.NewPostWorkspaceHandler(
				workspace.MapPostWorkspaceHttp,
				postHandle,
				marshal.DefaultMarshalerProvider,
				marshal.DefaultUnmarshalerProvider,
			)))
}

func withAuthHeaderInfo(next http.Handler) http.Handler {
	return middleware.NewHeaderInfoMiddleware(next, map[string]interface{}{
		"X-Subject": ccontext.UserKey,
	})
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
