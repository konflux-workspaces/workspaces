package writeclient_test

import (
	"context"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/writeclient"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

var _ = Describe("WriteclientCreate", func() {
	var ctx context.Context
	var fakeClient client.WithWatch
	var cli *writeclient.WriteClient

	user := "foo"
	namespace := "bar"
	workspace := restworkspacesv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "owner",
			Name:      "workspace-foo",
		},
		Spec: restworkspacesv1alpha1.WorkspaceSpec{},
		Status: restworkspacesv1alpha1.WorkspaceStatus{
			Space: &restworkspacesv1alpha1.SpaceInfo{
				Name: "space",
			},
		},
	}

	validateCreatedInternalWorkspace := func(w *restworkspacesv1alpha1.Workspace, expectedVisibility workspacesv1alpha1.InternalWorkspaceVisibility) {
		ww := workspacesv1alpha1.InternalWorkspaceList{}
		err := fakeClient.List(
			ctx,
			&ww,
			client.InNamespace(namespace),
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(ww.Items).ToNot(BeEmpty())
		Expect(ww.Items).To(Satisfy(func(ww []workspacesv1alpha1.InternalWorkspace) bool {
			return slices.ContainsFunc(ww, func(lw workspacesv1alpha1.InternalWorkspace) bool {
				return lw.Spec.DisplayName == w.Name &&
					lw.Status.Owner.Username == "owner" &&
					lw.Spec.Visibility == expectedVisibility
			})
		}))
	}

	BeforeEach(func() {
		ctx = context.Background()
		fakeClient = fake.NewFakeClient()
		err := workspacesv1alpha1.AddToScheme(fakeClient.Scheme())
		Expect(err).NotTo(HaveOccurred())

		clientFunc := func(string) (client.Client, error) {
			return fakeClient, nil
		}

		iwcli := iwclient.New(fakeClient, namespace, namespace)
		cli = writeclient.New(clientFunc, namespace, iwcli)
	})

	When("creating a community workspace", func() {
		It("should create workspaces using the given client", func() {
			// given
			workspace.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibilityCommunity

			// when
			err := cli.CreateUserWorkspace(ctx, user, &workspace)

			// then
			Expect(err).NotTo(HaveOccurred())
			validateCreatedInternalWorkspace(&workspace, workspacesv1alpha1.InternalWorkspaceVisibilityCommunity)
		})
	})

	When("creating a private workspace", func() {
		It("should create workspaces using the given client", func() {
			// given
			workspace.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibilityPrivate

			// when
			err := cli.CreateUserWorkspace(ctx, user, &workspace)

			// then
			Expect(err).NotTo(HaveOccurred())
			validateCreatedInternalWorkspace(&workspace, workspacesv1alpha1.InternalWorkspaceVisibilityPrivate)
		})
	})

	When("creating an owned workspace", func() {
		It("should have the is-owner label", func() {
			// given
			workspace.Spec.Visibility = restworkspacesv1alpha1.WorkspaceVisibilityPrivate

			// when
			err := cli.CreateUserWorkspace(ctx, "owner", &workspace)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(workspace.Labels).To(HaveKeyWithValue(restworkspacesv1alpha1.LabelIsOwner, "true"))
			validateCreatedInternalWorkspace(&workspace, workspacesv1alpha1.InternalWorkspaceVisibilityPrivate)
		})
	})
})
