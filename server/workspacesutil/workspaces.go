package workspacesutil

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// ErrWorkspaceWithoutOwner is the error returned if the workspace has no owner label
var ErrWorkspaceWithoutOwner = fmt.Errorf("error workspace has no owner")

// GetOwner returns the value of owner label if present, otherwise an error ErrWorkspaceWithoutOwner
func GetOwner(w *workspacesv1alpha1.Workspace) (*string, error) {
	ll := w.GetLabels()
	if len(ll) == 0 {
		return nil, fmt.Errorf("%w: %v", ErrWorkspaceWithoutOwner, client.ObjectKeyFromObject(w))
	}

	ol, ok := ll[workspacesv1alpha1.LabelWorkspaceOwner]
	if !ok || ol == "" {
		return nil, fmt.Errorf("%w: %v", ErrWorkspaceWithoutOwner, client.ObjectKeyFromObject(w))
	}
	return &ol, nil
}
