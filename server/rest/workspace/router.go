package workspace

import (
	"fmt"
	"net/http"
)

func AddWorkspaces(
	mux *http.ServeMux,
	prefix string,
	c http.Handler,
	r *ReadWorkspaceHandler,
	l *ListWorkspaceHandler,
	u, d http.Handler,
) {

	mux.Handle(fmt.Sprintf("GET %s", prefix), l)

	mux.Handle(fmt.Sprintf("GET %s/{name}", prefix), r)
}
