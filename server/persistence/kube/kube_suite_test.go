package kube_test

import (
	"context"
	"log/slog"
	"testing"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/kube"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestKube(t *testing.T) {
	slog.SetDefault(slog.New(&log.NoOpHandler{}))

	RegisterFailHandler(Fail)
	RunSpecs(t, "Kube Suite")
}

var _ = Describe("Kube Client", func() {
	var ctx context.Context
	user := "foo"
	namespace := "bar"
	workspace := workspacesv1alpha1.Workspace{
		ObjectMeta: v1.ObjectMeta{
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

		cli := kube.New(clientFunc, namespace)

		// when
		err = cli.CreateUserWorkspace(ctx, user, &workspace)

		// then
		Expect(err).NotTo(HaveOccurred())

		w := workspacesv1alpha1.Workspace{}
		err = fakeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: workspace.Name}, &w)
		Expect(err).NotTo(HaveOccurred())
		Expect(w).To(Equal(workspace))
	})
})
