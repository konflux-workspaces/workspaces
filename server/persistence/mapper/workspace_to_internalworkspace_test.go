package mapper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
)

var _ = Describe("WorkspaceToInternalworkspace", func() {
	When("a valid InternalWorkspace is converted", func() {
		var workspace restworkspacesv1alpha1.Workspace
		// workspacesNamespace := "foo"
		displayName := "bar"
		ownerName := "baz"

		BeforeEach(func() {
			// given
			workspace = buildExampleValidWorkspace(displayName, ownerName)
		})

		When("visibility is community", func() {
			BeforeEach(func() {
				workspace.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibilityCommunity
			})

			It("converts successfully", func() {
				// when
				iw, err := mapper.Default.WorkspaceToInternalWorkspace(&workspace)

				// then
				Expect(err).NotTo(HaveOccurred())
				validateMappedInternalWorkspace(iw, &workspace)
				Expect(iw.Spec.Visibility).To(Equal(workspacesv1alpha1.InternalWorkspaceVisibilityCommunity))
			})
		})

		When("visibility is private", func() {
			BeforeEach(func() {
				workspace.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibilityPrivate
			})

			It("converts successfully", func() {
				// when
				iw, err := mapper.Default.WorkspaceToInternalWorkspace(&workspace)

				// then
				Expect(err).NotTo(HaveOccurred())
				validateMappedInternalWorkspace(iw, &workspace)
				Expect(iw.Spec.Visibility).To(Equal(workspacesv1alpha1.InternalWorkspaceVisibilityPrivate))
			})
		})
	})
})

func buildExampleValidWorkspace(displayName, ownerName string) restworkspacesv1alpha1.Workspace {
	return restworkspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      displayName,
			Namespace: ownerName,
			Labels: map[string]string{
				"expected-label": "not-empty",
				workspacesv1alpha1.LabelInternalDomain + "not-expected-label": "not-empty",
			},
			Generation: 1,
		},
		Spec: restworkspacesv1alpha1.WorkspaceSpec{
			Visibility: restworkspacesv1alpha1.WorkspaceVisibilityCommunity,
		},
		Status: restworkspacesv1alpha1.WorkspaceStatus{
			Owner: &restworkspacesv1alpha1.UserInfoStatus{
				Email: "user@email.com",
			},
			Space: &restworkspacesv1alpha1.SpaceInfo{
				Name: "space-name",
			},
		},
	}
}

func validateMappedInternalWorkspace(w *workspacesv1alpha1.InternalWorkspace, from *restworkspacesv1alpha1.Workspace) {
	Expect(w).ToNot(BeNil())
	Expect(w.Generation).To(Equal(int64(1)))
	Expect(w.GetName()).To(BeZero())
	Expect(w.GetNamespace()).To(BeZero())
	Expect(w.GetLabels()).To(HaveKey("expected-label"))
	Expect(w.GetLabels()["expected-label"]).To(Equal("not-empty"))
	Expect(w.GetLabels()).NotTo(HaveKey(workspacesv1alpha1.LabelInternalDomain + "not-expected-label"))
	Expect(w.Spec).ToNot(BeNil())
	Expect(w.Spec.DisplayName).To(Equal(from.Name))
	Expect(w.Status.Owner.Username).To(Equal(from.Namespace))
}
