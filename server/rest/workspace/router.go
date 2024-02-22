package workspace

import (
	"net/http"
)

func AddWorkspaces(
	c http.Handler,
	r *ReadWorkspaceHandler,
	l *ListWorkspaceHandler,
	u, d http.Handler,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /workspaces", l)
	mux.Handle("GET /workspaces/{name}", r)
	return mux
}
