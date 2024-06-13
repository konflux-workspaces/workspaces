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

	iw := &workspacesv1alpha1.InternalWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Labels:     ll,
			Generation: workspace.Generation,
		},
		Spec: workspacesv1alpha1.InternalWorkspaceSpec{
			DisplayName: workspace.Name,
			Visibility:  workspacesv1alpha1.InternalWorkspaceVisibility(workspace.Spec.Visibility),
			Owner: workspacesv1alpha1.UserInfo{
				JwtInfo: workspacesv1alpha1.JwtInfo{},
			},
		},
		Status: workspacesv1alpha1.InternalWorkspaceStatus{
			Space: workspacesv1alpha1.SpaceInfo{
				IsHome: workspace.Name == "default",
			},
			Owner: workspacesv1alpha1.UserInfoStatus{
				Username: workspace.Namespace,
			},
		},
	}

	if o := workspace.Status.Owner; o != nil {
		iw.Spec.Owner.JwtInfo.Email = o.Email
	}
	if s := workspace.Status.Space; s != nil {
		iw.Status.Space.Name = s.Name
	}

	return iw, nil
}
