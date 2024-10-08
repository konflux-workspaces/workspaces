package mapper

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

func (m *Mapper) InternalWorkspaceToWorkspace(workspace *workspacesv1alpha1.InternalWorkspace) (*restworkspacesv1alpha1.Workspace, error) {
	// retrieve external labels
	wll := map[string]string{}
	for k, v := range workspace.GetLabels() {
		if !strings.HasPrefix(k, workspacesv1alpha1.LabelInternalDomain) {
			wll[k] = v
		}
	}

	return &restworkspacesv1alpha1.Workspace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workspace",
			APIVersion: restworkspacesv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              workspace.Spec.DisplayName,
			Namespace:         workspace.Status.Owner.Username,
			CreationTimestamp: workspace.CreationTimestamp,
			Labels:            wll,
			Generation:        workspace.Generation,
		},
		Spec: restworkspacesv1alpha1.WorkspaceSpec{
			Visibility: restworkspacesv1alpha1.WorkspaceVisibility(workspace.Spec.Visibility),
		},
		Status: restworkspacesv1alpha1.WorkspaceStatus{
			Space: &restworkspacesv1alpha1.SpaceInfo{
				Name:          workspace.Status.Space.Name,
				TargetCluster: workspace.Status.Space.TargetCluster,
			},
			Owner: &restworkspacesv1alpha1.UserInfoStatus{
				Email: workspace.Spec.Owner.JwtInfo.Email,
			},
			Conditions: workspace.Status.Conditions,
		},
	}, nil
}
