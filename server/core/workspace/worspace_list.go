package workspace

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
)

// ListWorkspaceQuery contains the information needed to retrieve all the workspaces the user has access to from the data source
type ListWorkspaceQuery struct {
	Namespace string
}

// ListWorkspaceResponse contains all the workspaces the user can access
type ListWorkspaceResponse struct {
	Workspaces workspacesv1alpha1.WorkspaceList
}

// WorkspaceLister is the interface the data source needs to implement to allow the ListWorkspaceHandler to fetch data from it
type WorkspaceLister interface {
	ListUserWorkspaces(ctx context.Context, user string, objs *workspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error
}

// ListWorkspaceHandler process ListWorkspaceQuery and returns a ListWorkspaceResponse fetching data from a WorkspaceLister
type ListWorkspaceHandler struct {
	lister WorkspaceLister
}

// NewListWorkspaceHandler creates a new ListWorkspaceHandler that uses a specified WorkspaceLister
func NewListWorkspaceHandler(lister WorkspaceLister) *ListWorkspaceHandler {
	return &ListWorkspaceHandler{lister: lister}
}

// Handle handles a ListWorkspaceQuery abd returns a ListWorkspaceResponse or an error
func (h *ListWorkspaceHandler) Handle(ctx context.Context, query ListWorkspaceQuery) (*ListWorkspaceResponse, error) {
	// authorization
	// If required, implement here complex logic like multiple-domains filtering, etc
	u, ok := ctx.Value(ccontext.UserKey).(string)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	// validate query
	// TODO: sanitize input, block reserved labels, etc

	// data access
	ww := workspacesv1alpha1.WorkspaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "WorkspaceList",
			APIVersion: "workspaces.io/v1alpha1",
		},
		Items: []workspacesv1alpha1.Workspace{},
	}
	opts := &client.ListOptions{
		Namespace: query.Namespace,
	}
	if err := h.lister.ListUserWorkspaces(ctx, u, &ww, opts); err != nil {
		return nil, err
	}

	for _, w := range ww.Items {
		switch query.Namespace {
		case "":
			w.SetNamespace(query.Namespace)
		default:
			ll := w.GetLabels()
			ow := ll[workspacesv1alpha1.LabelWorkspaceOwner]
			w.SetNamespace(ow)
		}
	}

	// reply
	return &ListWorkspaceResponse{Workspaces: ww}, nil
}
