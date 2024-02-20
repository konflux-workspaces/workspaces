package router

import (
	"fmt"
	"net/http"
)

type Router struct {
	Workspaces WorkspacesRouter
}

type WorkspacesRouter struct {
	Get http.Handler
}

func NewWorkspacesRouter(get http.Handler) *WorkspacesRouter {
	return &WorkspacesRouter{
		Get: get,
	}
}

func (wr *WorkspacesRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wr.Get.ServeHTTP(w, r)

	case http.MethodPut:
		fallthrough
	case http.MethodPost:
		fallthrough
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Method %s not supported", r.Method)))
	}
}
