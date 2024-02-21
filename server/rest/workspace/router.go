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
  mux.HandleFunc(prefix, func(w http.ResponseWriter, rq *http.Request) {
    switch rq.Method {
    case http.MethodGet:
      l.ServeHTTP(w, rq)

    default:
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte(fmt.Sprintf("Method %s not supported", rq.Method)))
    }
  })

  mux.HandleFunc(fmt.Sprintf("%s/", prefix), func(w http.ResponseWriter, rq *http.Request) {
    switch rq.Method {
    case http.MethodGet:
      r.ServeHTTP(w, rq)
    case http.MethodPost:
      c.ServeHTTP(w, rq)
    case http.MethodPut:
      u.ServeHTTP(w, rq)
    case http.MethodDelete:
      d.ServeHTTP(w, rq)

    default:
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte(fmt.Sprintf("Method %s not supported", rq.Method)))
    }
  })
}
