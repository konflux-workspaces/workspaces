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

		validateWorkspace := func(w *restworkspacesv1alpha1.Workspace) {
			Expect(w).ToNot(BeNil())
			Expect(w.GetName()).To(Equal(displayName))
			Expect(w.GetNamespace()).To(Equal(ownerName))
			Expect(w.GetLabels()).To(HaveKey("expected-label"))
			Expect(w.GetLabels()["expected-label"]).To(Equal("not-empty"))
			Expect(w.GetLabels()).NotTo(HaveKey(workspacesv1alpha1.LabelInternalDomain + "not-expected-label"))
			Expect(w.Generation).To(Equal(int64(1)))
			Expect(w.Spec).ToNot(BeNil())
		}

		BeforeEach(func() {
			// given
			internalWorkspace = workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: displayName,
					Namespace:    workspacesNamespace,
					Labels: map[string]string{
						workspacesv1alpha1.LabelDisplayName:                           displayName,
						workspacesv1alpha1.LabelWorkspaceOwner:                        ownerName,
						"expected-label":                                              "not-empty",
						workspacesv1alpha1.LabelInternalDomain + "not-expected-label": "not-empty",
					},
					Generation: 1,
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{},
			}
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
				validateWorkspace(w)
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
				validateWorkspace(w)
				Expect(w.Spec.Visibility).To(Equal(restworkspacesv1alpha1.WorkspaceVisibilityPrivate))
			})
		})
	})

	When("a InternalWorkspace is missing internal owner name label", func() {
		var internalWorkspace workspacesv1alpha1.InternalWorkspace
		workspacesNamespace := "foo"
		displayName := "bar"

		BeforeEach(func() {
			// given
			internalWorkspace = workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: displayName,
					Namespace:    workspacesNamespace,
					Labels: map[string]string{
						workspacesv1alpha1.LabelDisplayName: displayName,
					},
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Visibility: workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
				},
			}
		})

		It("returns an error", func() {
			// when
			w, err := mapper.Default.InternalWorkspaceToWorkspace(&internalWorkspace)

			// then
			Expect(err).To(HaveOccurred())
			Expect(w).To(BeNil())
		})
	})

	When("a InternalWorkspace is missing internal display name label", func() {
		var internalWorkspace workspacesv1alpha1.InternalWorkspace
		workspacesNamespace := "foo"
		ownerName := "baz"

		BeforeEach(func() {
			// given
			internalWorkspace = workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    workspacesNamespace,
					Labels: map[string]string{
						workspacesv1alpha1.LabelWorkspaceOwner: ownerName,
					},
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Visibility: workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
				},
			}
		})

		It("returns an error", func() {
			// when
			w, err := mapper.Default.InternalWorkspaceToWorkspace(&internalWorkspace)

			// then
			Expect(err).To(HaveOccurred())
			Expect(w).To(BeNil())
		})
	})
})
