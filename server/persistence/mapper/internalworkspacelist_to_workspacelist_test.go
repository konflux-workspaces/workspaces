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
	emptyList := restworkspacesv1alpha1.WorkspaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "WorkspaceList",
			APIVersion: restworkspacesv1alpha1.GroupVersion.String(),
		},
		Items: make([]restworkspacesv1alpha1.Workspace, 0),
	}

	DescribeTable("returns empty list",
		func(workspaces *workspacesv1alpha1.InternalWorkspaceList) {
			// when
			w, err := mapper.Default.InternalWorkspaceListToWorkspaceList(workspaces)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(w).ToNot(BeNil())
			Expect(*w).To(Equal(emptyList))
		},
		Entry("for nil input", nil),
		Entry("for nil input's items", &workspacesv1alpha1.InternalWorkspaceList{
			Items: nil,
		}),
		Entry("for empty input's items", &workspacesv1alpha1.InternalWorkspaceList{
			Items: []workspacesv1alpha1.InternalWorkspace{},
		}),
		Entry("for nil input", nil),
	)

	It("maps all input items", func() {
		// given
		workspaces := make([]workspacesv1alpha1.InternalWorkspace, 5)
		for i := 0; i < len(workspaces); i++ {
			workspaces[i] = buildExampleValidInternalWorkspace(
				fmt.Sprintf("workspace-%d", i),
				"workspaces-system",
				fmt.Sprintf("owner-%d", i),
			)
		}

		// when
		ww, err := mapper.Default.InternalWorkspaceListToWorkspaceList(&workspacesv1alpha1.InternalWorkspaceList{Items: workspaces})

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(ww.Items).To(HaveLen(len(workspaces)))
		for _, w := range ww.Items {
			i := slices.IndexFunc(workspaces,
				func(iw workspacesv1alpha1.InternalWorkspace) bool {
					return iw.Spec.DisplayName == w.Name && iw.Status.Owner.Username == w.Namespace
				})
			Expect(i).NotTo(Equal(-1), "could not find internal workspaces for mapped workspace %s/%s", w.Namespace, w.Name)
			validateMappedWorkspace(&w, workspaces[i])
		}
	})
})
