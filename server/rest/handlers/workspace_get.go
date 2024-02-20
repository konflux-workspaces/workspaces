package handlers

import (
	"net/http"
	"strings"
)

var _ http.Handler = &GetWorkspaceHttpHandler{}

func NewGetWorkspaceHttpHandler(
	prefix string,
	readWorkspaceHandler *ReadWorkspaceHandler,
	listWorkspaceHandler *ListWorkspaceHandler,
) *GetWorkspaceHttpHandler {
	return &GetWorkspaceHttpHandler{
		Prefix: prefix,

		ReadWorkspaceHandler: readWorkspaceHandler,
		ListWorkspaceHandler: listWorkspaceHandler,
	}
}

type GetWorkspaceHttpHandler struct {
	Prefix               string
	ReadWorkspaceHandler http.Handler
	ListWorkspaceHandler http.Handler
}

func (h *GetWorkspaceHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.Prefix) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch {
	case r.URL.Path == h.Prefix:
		h.ListWorkspaceHandler.ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, h.Prefix):
		h.ReadWorkspaceHandler.ServeHTTP(w, r)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
