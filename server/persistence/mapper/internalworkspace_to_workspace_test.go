package mapper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
)

var _ = Describe("InternalworkspaceToWorkspace", func() {

	When("a valid InternalWorkspace is converted", func() {
		var internalWorkspace workspacesv1alpha1.InternalWorkspace
		workspacesNamespace := "foo"
		displayName := "bar"
		ownerName := "baz"

		BeforeEach(func() {
			// given
			internalWorkspace = buildExampleValidInternalWorkspace(displayName, workspacesNamespace, ownerName)
		})

		When("visibility is community", func() {
			BeforeEach(func() {
				internalWorkspace.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityCommunity
			})

			It("converts successfully", func() {
				// when
				w, err := mapper.Default.InternalWorkspaceToWorkspace(&internalWorkspace)

				// then
				Expect(err).NotTo(HaveOccurred())
				validateMappedWorkspace(w, internalWorkspace)
				Expect(w.Spec.Visibility).To(Equal(restworkspacesv1alpha1.WorkspaceVisibilityCommunity))
			})
		})

		When("visibility is private", func() {
			BeforeEach(func() {
				internalWorkspace.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityPrivate
			})

			It("converts successfully", func() {
				// when
				w, err := mapper.Default.InternalWorkspaceToWorkspace(&internalWorkspace)

				// then
				Expect(err).NotTo(HaveOccurred())
				validateMappedWorkspace(w, internalWorkspace)
				Expect(w.Spec.Visibility).To(Equal(restworkspacesv1alpha1.WorkspaceVisibilityPrivate))
			})
		})
	})
})

func buildExampleValidInternalWorkspace(displayName, workspacesNamespace, ownerName string) workspacesv1alpha1.InternalWorkspace {
	return workspacesv1alpha1.InternalWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      displayName,
			Namespace: workspacesNamespace,
			Labels: map[string]string{
				"expected-label": "not-empty",
				workspacesv1alpha1.LabelInternalDomain + "not-expected-label": "not-empty",
			},
			Generation:        1,
			CreationTimestamp: metav1.Now(),
		},
		Spec: workspacesv1alpha1.InternalWorkspaceSpec{
			DisplayName: displayName,
		},
		Status: workspacesv1alpha1.InternalWorkspaceStatus{
			Owner: workspacesv1alpha1.UserInfoStatus{
				Username: ownerName,
			},
			Space: workspacesv1alpha1.SpaceInfo{
				IsHome:        true,
				Name:          displayName,
				TargetCluster: "target-cluster",
			},
			Conditions: []metav1.Condition{
				{Message: "test", Type: "test", Reason: "test", Status: metav1.ConditionTrue},
			},
		},
	}
}

func validateMappedWorkspace(w *restworkspacesv1alpha1.Workspace, from workspacesv1alpha1.InternalWorkspace) {
	Expect(w).ToNot(BeNil())
	Expect(w.GetName()).To(Equal(from.Spec.DisplayName))
	Expect(w.GetNamespace()).To(Equal(from.Status.Owner.Username))
	Expect(w.GetLabels()).To(And(
		HaveKeyWithValue("expected-label", "not-empty"),
		Not(HaveKey(restworkspacesv1alpha1.LabelIsOwner)),
		Not(HaveKey(workspacesv1alpha1.LabelInternalDomain+"not-expected-label")),
	))
	Expect(w.Generation).To(Equal(int64(1)))
	Expect(w.CreationTimestamp).To(Equal(from.CreationTimestamp))
	Expect(w.Spec).ToNot(BeNil())
	Expect(w.Status).ToNot(BeNil())
	Expect(w.Status.Space).ToNot(BeNil())
	Expect(w.Status.Space.Name).To(Equal(from.Status.Space.Name))
	Expect(w.Status.Space.TargetCluster).To(Equal(from.Status.Space.TargetCluster))
	Expect(w.Status.Conditions).To(Equal(from.Status.Conditions))
}
