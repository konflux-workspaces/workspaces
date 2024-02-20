package mapper

import (
	"net/http"

	"github.com/filariow/workspaces/server/query"
	"github.com/filariow/workspaces/server/rest/handlers"
)

var _ handlers.ListWorkspaceMapperFunc = MapListWorkspaceHttp

func MapListWorkspaceHttp(r *http.Request) (*query.ListWorkspaceQuery, error) {
	return &query.ListWorkspaceQuery{}, nil
}
