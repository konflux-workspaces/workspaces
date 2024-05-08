package kube_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"

	"github.com/konflux-workspaces/workspaces/server/persistence/kube"
)

var _ = Describe("WriteClient", func() {
	var ctx context.Context
	user := "foo"
	namespace := "bar"
	workspace := workspacesv1alpha1.InternalWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "workspace-foo",
		},
	}

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("should create workspaces using the given client", func() {
		fakeClient := fake.NewFakeClient()
		err := workspacesv1alpha1.AddToScheme(fakeClient.Scheme())
		Expect(err).NotTo(HaveOccurred())

		clientFunc := func(string) (client.Client, error) {
			return fakeClient, nil
		}

		cli := kube.NewWriteClient(clientFunc, namespace)

		// when
		err = cli.CreateUserWorkspace(ctx, user, &workspace)

		// then
		Expect(err).NotTo(HaveOccurred())

		w := workspacesv1alpha1.InternalWorkspace{}
		err = fakeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: workspace.Name}, &w)
		Expect(err).NotTo(HaveOccurred())
		Expect(w).To(Equal(workspace))
	})
})
