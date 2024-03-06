package mapper

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

func (m *Mapper) InternalWorkspaceToWorkspace(workspace *workspacesv1alpha1.InternalWorkspace) (*restworkspacesv1alpha1.Workspace, error) {
	ll := workspace.GetLabels()

	// retrieve WorkspaceOwner
	ownerUsername := workspace.Spec.Owner.JWTInfo.Username

	// retrieve DisplayName
	lname, ok := ll[workspacesv1alpha1.LabelDisplayName]
	if !ok {
		return nil, ErrLabelDisplayNameNotFound
	}

	// retrieve external labels
	wll := map[string]string{}
	for k, v := range ll {
		if !strings.HasPrefix(k, workspacesv1alpha1.LabelInternalDomain) {
			wll[k] = v
		}
	}

	return &restworkspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       lname,
			Namespace:  ownerUsername,
			Labels:     wll,
			Generation: workspace.Generation,
		},
		Spec: restworkspacesv1alpha1.WorkspaceSpec{
			Visibility: restworkspacesv1alpha1.WorkspaceVisibility(workspace.Spec.Visibility),
		},
	}, nil
}
