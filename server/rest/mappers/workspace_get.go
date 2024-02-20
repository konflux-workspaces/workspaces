package mapper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/filariow/workspaces/server/query"
	"github.com/filariow/workspaces/server/rest/handlers"
)

var _ handlers.ReadWorkspaceMapperFunc = MapReadWorkspaceHttp

func MapReadWorkspaceHttp(prefix string, r *http.Request) (*query.ReadWorkspaceQuery, error) {
	c, ok := strings.CutPrefix(r.URL.Path, prefix)
	if !ok {
		return nil, fmt.Errorf("")
	}

	n := strings.Trim(c, "/")
	return &query.ReadWorkspaceQuery{Name: n}, nil
}
