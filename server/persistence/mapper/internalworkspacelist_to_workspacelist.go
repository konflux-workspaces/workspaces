package mapper

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

func (m *Mapper) InternalWorkspaceListToWorkspaceList(workspaces *workspacesv1alpha1.InternalWorkspaceList) (*restworkspacesv1alpha1.WorkspaceList, error) {
	ww := restworkspacesv1alpha1.WorkspaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "WorkspaceList",
			APIVersion: restworkspacesv1alpha1.GroupVersion.String(),
		},
	}
	if workspaces == nil {
		ww.Items = make([]restworkspacesv1alpha1.Workspace, 0)
		return &ww, nil
	}

	ww.Items = make([]restworkspacesv1alpha1.Workspace, len(workspaces.Items))
	for i, w := range workspaces.Items {
		rw, err := m.InternalWorkspaceToWorkspace(&w)
		if err != nil {
			return nil, err
		}
		ww.Items[i] = *rw
	}
	return &ww, nil
}
