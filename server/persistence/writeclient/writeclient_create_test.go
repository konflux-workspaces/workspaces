package writeclient_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/writeclient"
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
	}

	validateCreatedInternalWorkspace := func(w *restworkspacesv1alpha1.Workspace, expectedVisibility workspacesv1alpha1.InternalWorkspaceVisibility) {
		ww := workspacesv1alpha1.InternalWorkspaceList{}
		err := fakeClient.List(
			ctx,
			&ww,
			client.InNamespace(namespace),
			client.MatchingLabels{
				workspacesv1alpha1.LabelDisplayName:    workspace.Name,
				workspacesv1alpha1.LabelWorkspaceOwner: "owner",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(ww.Items).To(HaveLen(1))
		Expect(ww.Items[0].Spec).ToNot(BeNil())
		Expect(ww.Items[0].Spec.Visibility).To(Equal(expectedVisibility))
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

	When("creating a community workspace", func() {
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
})
