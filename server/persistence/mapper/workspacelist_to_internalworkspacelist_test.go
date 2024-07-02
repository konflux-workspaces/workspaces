package mapper_test

import (
	"fmt"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
)

var _ = Describe("InternalworkspacelistToWorkspacelist", func() {
	emptyList := workspacesv1alpha1.InternalWorkspaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InternalWorkspaceList",
			APIVersion: workspacesv1alpha1.GroupVersion.String(),
		},
		Items: make([]workspacesv1alpha1.InternalWorkspace, 0),
	}

	DescribeTable("returns empty list",
		func(workspaces *restworkspacesv1alpha1.WorkspaceList) {
			// when
			w, err := mapper.Default.WorkspaceListToInternalWorkspaceList(workspaces)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(w).ToNot(BeNil())
			Expect(*w).To(Equal(emptyList))
		},
		Entry("for nil input", nil),
		Entry("for nil input's items", &restworkspacesv1alpha1.WorkspaceList{
			Items: nil,
		}),
		Entry("for empty input's items", &restworkspacesv1alpha1.WorkspaceList{
			Items: []restworkspacesv1alpha1.Workspace{},
		}),
		Entry("for nil input", nil),
	)

	It("maps all input items", func() {
		// given
		workspaces := make([]restworkspacesv1alpha1.Workspace, 5)
		for i := 0; i < len(workspaces); i++ {
			workspaces[i] = buildExampleValidWorkspace(
				fmt.Sprintf("workspace-%d", i),
				fmt.Sprintf("owner-%d", i),
			)
		}

		// when
		ww, err := mapper.Default.WorkspaceListToInternalWorkspaceList(&restworkspacesv1alpha1.WorkspaceList{Items: workspaces})

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(ww.Items).To(HaveLen(len(workspaces)))
		for _, w := range ww.Items {
			i := slices.IndexFunc(workspaces,
				func(iw restworkspacesv1alpha1.Workspace) bool {
					return w.Spec.DisplayName == iw.Name && w.Status.Owner.Username == iw.Namespace
				})
			Expect(i).NotTo(Equal(-1), "could not find internal workspaces for mapped workspace %s/%s", w.Namespace, w.Name)
			validateMappedInternalWorkspace(&w, &workspaces[i])
		}
	})
})
