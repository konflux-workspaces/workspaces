package cache_test

import (
	"context"
	"log/slog"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/persistence/cache"
)

func TestCache(t *testing.T) {
	slog.SetDefault(slog.New(&log.NoOpHandler{}))

	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}

var _ = Describe("Cache", func() {
	var ctx context.Context
	var c *cache.Cache

	ksns := "kubesaw-namespace"
	wsns := "workspaces-namespace"

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("should not return any workspace if no SpaceBinding exists", func() {
		// given
		c = buildCache(ksns, wsns,
			&workspacesv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-label",
					Namespace: wsns,
				},
			})

		// when
		var ww workspacesv1alpha1.WorkspaceList
		err := c.ListUserWorkspaces(ctx, "owner", &ww)
		Expect(err).NotTo(HaveOccurred())

		// then
		Expect(ww.Items).Should(HaveLen(0))
	})

	It("should not return any workspace that has no owner label", func() {
		// given
		c = buildCache(ksns, wsns,
			&workspacesv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-label",
					Namespace: wsns,
				},
			},
			&toolchainv1alpha1.SpaceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "owner-sb",
					Namespace: ksns,
				},
				Spec: toolchainv1alpha1.SpaceBindingSpec{
					MasterUserRecord: "owner-user",
					SpaceRole:        "admin",
					Space:            "no-label",
				},
			},
		)

		// when
		var ww workspacesv1alpha1.WorkspaceList
		err := c.ListUserWorkspaces(ctx, "owner", &ww)
		Expect(err).NotTo(HaveOccurred())

		// then
		Expect(ww.Items).Should(HaveLen(0))
	})

	When("an user has one workspace", func() {
		w := &workspacesv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "owner-ws",
				Namespace: wsns,
				Labels: map[string]string{
					workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
				},
			},
		}
		BeforeEach(func() {
			// given that just the 'owner-ws' workspace owned by the user 'owner-user' exists
			c = buildCache(ksns, wsns,
				w,
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "owner-user",
						SpaceRole:        "admin",
						Space:            "owner-ws",
					},
				},
			)
		})

		It("list should contains the one owned workspace", func() {
			// when the list of workspaces owned by 'owner-user' is requested
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			Expect(ww.Items).Should(HaveLen(1))
			Expect(ww.Items[0].Name).Should(Equal(w.Name))
			Expect(ww.Items[0].Namespace).Should(Equal("owner-user"))
		})

		It("read returns the owned workspace", func() {
			// when the list of workspaces owned by 'owner-user' is requested
			var rw workspacesv1alpha1.Workspace
			err := c.ReadUserWorkspace(ctx, "owner-user", "owner-user", w.Name, &rw)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			Expect(rw.Name).Should(Equal(w.Name))
			Expect(rw.Namespace).Should(Equal("owner-user"))
		})
	})
})

func buildCache(ksns, wsns string, objs ...client.Object) *cache.Cache {
	var err error
	scheme := runtime.NewScheme()
	err = workspacesv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = toolchainv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	fc := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	return cache.NewWithReader(fc, wsns, ksns)
}
