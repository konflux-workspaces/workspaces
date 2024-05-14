package mapper

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

func (m *Mapper) WorkspaceListToInternalWorkspaceList(workspaces *restworkspacesv1alpha1.WorkspaceList) (*workspacesv1alpha1.InternalWorkspaceList, error) {
	ww := workspacesv1alpha1.InternalWorkspaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InternalWorkspaceList",
			APIVersion: workspacesv1alpha1.GroupVersion.String(),
		},
		Items: make([]workspacesv1alpha1.InternalWorkspace, len(workspaces.Items)),
	}

	for _, w := range workspaces.Items {
		rw, err := m.WorkspaceToInternalWorkspace(&w)
		if err != nil {
			return nil, err
		}
		ww.Items = append(ww.Items, *rw)
	}
	return &ww, nil
}
