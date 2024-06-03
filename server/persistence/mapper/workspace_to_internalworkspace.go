package mapper

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

// WorkspaceToInternalWorkspace builds an InternalWorkspace starting from a Workspace.
// IMPORTANT: The Name and Namespace fields are left empty.
func (m *Mapper) WorkspaceToInternalWorkspace(workspace *restworkspacesv1alpha1.Workspace) (*workspacesv1alpha1.InternalWorkspace, error) {
	ll := map[string]string{}
	for k, v := range workspace.GetLabels() {
		if !strings.HasPrefix(k, workspacesv1alpha1.LabelInternalDomain) {
			ll[k] = v
		}
	}
	ll[workspacesv1alpha1.LabelDisplayName] = workspace.GetName()

	return &workspacesv1alpha1.InternalWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Labels:     ll,
			Generation: workspace.Generation,
		},
		Spec: workspacesv1alpha1.InternalWorkspaceSpec{
			Visibility: workspacesv1alpha1.InternalWorkspaceVisibility(workspace.Spec.Visibility),
			Owner: workspacesv1alpha1.UserInfo{
				JWTInfo: workspacesv1alpha1.JwtInfo{
					Username: workspace.Namespace,
				},
			},
		},
	}, nil
}
