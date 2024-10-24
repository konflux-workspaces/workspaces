package rest

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/rest/marshal"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware"
	"github.com/konflux-workspaces/workspaces/server/rest/workspace"
)

const (
	WorkspacesPrefix           string = `/apis/workspaces.konflux-ci.dev/v1alpha1/workspaces`
	NamespacedWorkspacesPrefix string = `/apis/workspaces.konflux-ci.dev/v1alpha1/namespaces/{namespace}/workspaces`
)

func New(
	logger *slog.Logger,
	addr string,
	cache cache.Cache,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
	createHandle workspace.CreateWorkspaceCommandHandlerFunc,
	updateHandle workspace.UpdateWorkspaceCommandHandlerFunc,
	patchHandle workspace.PatchWorkspaceCommandHandlerFunc,
) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           buildServerHandler(logger, cache, readHandle, listHandle, createHandle, updateHandle, patchHandle),
		ReadHeaderTimeout: 3 * time.Second,
	}
}

func buildServerHandler(
	logger *slog.Logger,
	cache cache.Cache,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
	createHandle workspace.CreateWorkspaceCommandHandlerFunc,
	updateHandle workspace.UpdateWorkspaceCommandHandlerFunc,
	patchHandle workspace.PatchWorkspaceCommandHandlerFunc,
) http.Handler {
	mux := http.NewServeMux()
	addHealthz(mux)
	addWorkspaces(mux, cache, readHandle, listHandle, createHandle, updateHandle, patchHandle)
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
	cache cache.Cache,
	readHandle workspace.ReadWorkspaceQueryHandlerFunc,
	listHandle workspace.ListWorkspaceQueryHandlerFunc,
	_ workspace.CreateWorkspaceCommandHandlerFunc,
	updateHandle workspace.UpdateWorkspaceCommandHandlerFunc,
	patchHandle workspace.PatchWorkspaceCommandHandlerFunc,
) {
	// Read
	mux.Handle(fmt.Sprintf("GET %s/{name}", NamespacedWorkspacesPrefix),
		withAuthHeaderInfo(
			withUserSignupAuth(cache,
				workspace.NewReadWorkspaceHandler(
					workspace.MapReadWorkspaceHttp,
					readHandle,
					marshal.DefaultMarshalerProvider,
				))))

	// List
	lh := withAuthHeaderInfo(
		withUserSignupAuth(cache,
			workspace.NewListWorkspaceHandler(
				workspace.MapListWorkspaceHttp,
				listHandle,
				marshal.DefaultMarshalerProvider,
			),
		))
	mux.Handle(fmt.Sprintf("GET %s", WorkspacesPrefix), lh)
	mux.Handle(fmt.Sprintf("GET %s", NamespacedWorkspacesPrefix), lh)

	// Update
	mux.Handle(fmt.Sprintf("PUT %s/{name}", NamespacedWorkspacesPrefix),
		withAuthHeaderInfo(
			withUserSignupAuth(cache,
				workspace.NewUpdateWorkspaceHandler(
					workspace.MapPutWorkspaceHttp,
					updateHandle,
					marshal.DefaultMarshalerProvider,
					marshal.DefaultUnmarshalerProvider,
				))))

	// Patch
	mux.Handle(fmt.Sprintf("PATCH %s/{name}", NamespacedWorkspacesPrefix),
		withAuthHeaderInfo(
			withUserSignupAuth(cache,
				workspace.NewPatchWorkspaceHandler(
					workspace.MapPatchWorkspaceHttp,
					patchHandle,
					marshal.DefaultMarshalerProvider,
				))))

	// Create
	// mux.Handle(fmt.Sprintf("POST %s", NamespacedWorkspacesPrefix),
	// 	withAuthHeaderInfo(
	// 		withUserSignupAuth(cache, true,
	// 			workspace.NewPostWorkspaceHandler(
	// 				workspace.MapPostWorkspaceHttp,
	// 				postHandle,
	// 				marshal.DefaultMarshalerProvider,
	// 				marshal.DefaultUnmarshalerProvider,
	// 			))))
}

func withAuthHeaderInfo(next http.Handler) http.Handler {
	return middleware.NewHeaderInfoMiddleware(next, map[string]interface{}{
		"X-Subject": ccontext.UserSubKey,
	})
}

func withUserSignupAuth(cache cache.Cache, next http.Handler) http.Handler {
	return middleware.NewUserSignupMiddleware(next, cache)
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
