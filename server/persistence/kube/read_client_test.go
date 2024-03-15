package kube_test

import (
	"context"
	"fmt"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"

	"github.com/konflux-workspaces/workspaces/server/persistence/kube"
)

var _ = Describe("ReadClient", func() {
	var ctx context.Context
	var c *kube.ReadClient

	ksns := "kubesaw-namespace"
	wsns := "workspaces-namespace"

	BeforeEach(func() {
		ctx = context.Background()
	})

	When("no SpaceBinding exists for a workspace", func() {
		// given
		c = buildCache(ksns, wsns,
			&workspacesv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-space-binding",
					Namespace: wsns,
					Labels: map[string]string{
						workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
					},
				},
			})

		It("should not return the workspace in list", func() {
			// when
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "owner", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(HaveLen(0))
		})

		It("should not return the workspace in read", func() {
			// when
			var ww workspacesv1alpha1.Workspace
			err := c.ReadUserWorkspace(ctx, "owner", "owner", "no-space-binding", &ww)

			// then
			Expect(err).To(HaveOccurred())
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})
	})

	When("owner label is not set on workspace", func() {
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

		It("should not return the workspace in list", func() {
			// when
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "owner", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(HaveLen(0))
		})

		It("should not return the workspace in read", func() {
			// when
			var ww workspacesv1alpha1.Workspace
			err := c.ReadUserWorkspace(ctx, "owner", "owner", "no-label", &ww)

			// then
			Expect(err).To(HaveOccurred())
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})
	})

	When("one valid workspace exists", func() {
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

		It("should be returned in list of owner's workspaces", func() {
			// when the list of workspaces owned by 'owner-user' is requested
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			Expect(ww.Items).Should(HaveLen(1))
			Expect(ww.Items[0].Name).Should(Equal(w.Name))
			Expect(ww.Items[0].Namespace).Should(Equal("owner-user"))
		})

		It("should be returned in read", func() {
			// when
			var rw workspacesv1alpha1.Workspace
			err := c.ReadUserWorkspace(ctx, "owner-user", "owner-user", w.Name, &rw)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(rw.Name).Should(Equal(w.Name))
			Expect(rw.Namespace).Should(Equal("owner-user"))
		})

		It("should NOT be returned in list of not-owner's workspaces", func() {
			// when
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "not-owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(HaveLen(0))
		})

		It("should NOT be returned in read of not-owner-user workspace", func() {
			// when
			var rw *workspacesv1alpha1.Workspace
			err := c.ReadUserWorkspace(ctx, "not-owner-user", "owner-user", w.Name, rw)

			// then
			Expect(err).To(HaveOccurred())
			Expect(kerrors.IsNotFound(err)).Should(BeTrue())
			Expect(rw).Should(BeNil())
		})
	})

	When("more than one valid workspace exist", func() {
		ww := make([]*workspacesv1alpha1.Workspace, 10)
		sbs := make([]*toolchainv1alpha1.SpaceBinding, len(ww))
		for i := 0; i < len(ww); i++ {
			wsName := fmt.Sprintf("owner-ws-%d", i)
			ww[i] = &workspacesv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wsName,
					Namespace: wsns,
					Labels: map[string]string{
						workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
					},
				},
			}

			sbs[i] = &toolchainv1alpha1.SpaceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("owner-sb-%d", i),
					Namespace: ksns,
				},
				Spec: toolchainv1alpha1.SpaceBindingSpec{
					MasterUserRecord: "owner-user",
					SpaceRole:        "admin",
					Space:            wsName,
				},
			}
		}

		ees := len(ww) + len(sbs)
		ee := make([]client.Object, ees, ees)
		for i, w := range ww {
			ee[i] = w
		}
		for i, sb := range sbs {
			ee[10+i] = sb
		}

		BeforeEach(func() {
			c = buildCache(ksns, wsns, ee...)
		})

		It("should be returned in list of owner's workspaces", func() {
			// when the list of workspaces owned by 'owner-user' is requested
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			wwi := ww.Items
			Expect(wwi).Should(HaveLen(len(wwi)))

			for _, w := range wwi {
				sw := slices.ContainsFunc(wwi, func(z workspacesv1alpha1.Workspace) bool {
					return w.Name == z.Name && w.Namespace == "owner-user"
				})
				Expect(sw).Should(BeTrue())
			}
		})

		It("should be returned in read", func() {
			for _, w := range ww {
				// when
				var rw workspacesv1alpha1.Workspace
				err := c.ReadUserWorkspace(ctx, "owner-user", "owner-user", w.Name, &rw)
				Expect(err).NotTo(HaveOccurred())

				// then
				Expect(rw.Name).Should(Equal(w.Name))
				Expect(rw.Namespace).Should(Equal("owner-user"))
			}
		})

		It("should NOT be returned in list of not-owner's workspaces", func() {
			// when
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "not-owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(HaveLen(0))
		})

		It("should NOT be returned in read of not-owner-user workspace", func() {
			for _, w := range ww {
				// when
				var rw *workspacesv1alpha1.Workspace
				err := c.ReadUserWorkspace(ctx, "not-owner-user", "owner-user", w.Name, rw)

				// then
				Expect(err).To(HaveOccurred())
				Expect(kerrors.IsNotFound(err)).Should(BeTrue())
				Expect(rw).Should(BeNil())
			}
		})
	})

	When("workspace is created outside monitored namespaced", func() {
		BeforeEach(func() {
			c = buildCache(ksns, wsns,
				&workspacesv1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-ws",
						Namespace: "not-monitored",
						Labels: map[string]string{
							workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
						},
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
						Space:            "owner-ws",
					},
				},
			)
		})

		It("is not returned in list", func() {
			// when
			var ww workspacesv1alpha1.WorkspaceList
			err := c.ListUserWorkspaces(ctx, "owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(HaveLen(0))
		})

		It("is not returned in read", func() {
			// when
			var w workspacesv1alpha1.Workspace
			err := c.ReadUserWorkspace(ctx, "owner-user", "owner-user", "owner-ws", &w)
			// then
			Expect(err).To(HaveOccurred())
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})
	})
})

func buildCache(ksns, wsns string, objs ...client.Object) *kube.ReadClient {
	var err error
	scheme := runtime.NewScheme()
	err = workspacesv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = toolchainv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	fc := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	return kube.NewReadClientWithReader(fc, wsns, ksns)
}
