package rest

import (
	"encoding/json"
	"net/http"

	"github.com/filariow/workspaces/server/query"
	"github.com/filariow/workspaces/server/rest/handlers"
	mapper "github.com/filariow/workspaces/server/rest/mappers"
	"github.com/filariow/workspaces/server/rest/marshal"
	"github.com/filariow/workspaces/server/rest/router"
)

const WorkspacesPrefix string = "/workspaces/"

type Server struct {
	*http.Server
}

type Config struct {
	Marshaler   marshal.MarshalFunc
	Unmarshaler marshal.UnmarshalFunc
}

func New(addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: buildServerMux(),
	}
}

func buildServerMux() *http.ServeMux {
	mux := http.NewServeMux()

	addWorkspaces(mux, WorkspacesPrefix)
	return mux
}

func addWorkspaces(mux *http.ServeMux, prefix string) {
	g := handlers.NewReadWorkspaceHandler(prefix, mapper.MapReadWorkspaceHttp, query.ReadWorkspaceHandler, json.Marshal, json.Unmarshal)
	r := router.NewWorkspacesRouter(g)

	mux.Handle(prefix, r)
}
