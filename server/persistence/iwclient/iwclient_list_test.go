package iwclient_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"

	"github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
)

var _ = Describe("List", func() {
	var ctx context.Context
	var c *iwclient.Client

	ksns := "kubesaw-namespace"
	wsns := "workspaces-namespace"

	BeforeEach(func() {
		ctx = context.Background()
	})

	When("no SpaceBinding exists for a workspace", func() {
		// given
		c = buildCache(wsns, ksns,
			&workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-space-binding",
					Namespace: wsns,
				},
				Status: workspacesv1alpha1.InternalWorkspaceStatus{
					Owner: workspacesv1alpha1.UserInfoStatus{
						Username: "owner-user",
					},
				},
			})

		It("should not return the workspace in list", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "owner", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).To(BeEmpty())
		})
	})

	When("owner label is not set on workspace", func() {
		// given
		c = buildCache(wsns, ksns,
			&workspacesv1alpha1.InternalWorkspace{
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
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "owner", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).To(BeEmpty())
		})

	})

	When("one valid workspace exists", func() {
		wName := "owner-ws"
		w := &workspacesv1alpha1.InternalWorkspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateName(wName),
				Namespace: wsns,
			},
			Spec: workspacesv1alpha1.InternalWorkspaceSpec{
				DisplayName: wName,
			},
			Status: workspacesv1alpha1.InternalWorkspaceStatus{
				Owner: workspacesv1alpha1.UserInfoStatus{
					Username: "owner-user",
				},
				Space: workspacesv1alpha1.SpaceInfo{
					IsHome: false,
					Name:   "space",
				},
			},
		}
		BeforeEach(func() {
			// given that just the 'owner-ws' workspace owned by the user 'owner-user' exists
			c = buildCache(wsns, ksns,
				w,
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            w.Name,
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "owner-user",
						SpaceRole:        "admin",
						Space:            "space",
					},
				},
			)
		})

		It("should be returned in list of owner's workspaces", func() {
			// when the list of workspaces owned by 'owner-user' is requested
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			Expect(ww.Items).To(HaveLen(1))
			Expect(ww.Items[0].Spec.DisplayName).To(Equal("owner-ws"))
			Expect(ww.Items[0].Status.Owner.Username).To(Equal("owner-user"))
			Expect(ww.Items[0].Name).ToNot(Equal("owner-ws"))
			Expect(ww.Items[0].Namespace).ToNot(Equal("owner-user"))
		})

		It("should NOT be returned in list of not-owner's workspaces", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "not-owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).To(BeEmpty())
		})
	})

	When("more than one valid workspace exist", func() {
		ww := make([]*workspacesv1alpha1.InternalWorkspace, 10)
		sbs := make([]*toolchainv1alpha1.SpaceBinding, len(ww))
		for i := 0; i < len(ww); i++ {
			wName := fmt.Sprintf("owner-ws-%d", i)
			ww[i] = &workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName(wName),
					Namespace: wsns,
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					DisplayName: wName,
					Visibility:  workspacesv1alpha1.InternalWorkspaceVisibilityPrivate,
				},
				Status: workspacesv1alpha1.InternalWorkspaceStatus{
					Space: workspacesv1alpha1.SpaceInfo{
						IsHome: false,
						Name:   wName,
					},
					Owner: workspacesv1alpha1.UserInfoStatus{
						Username: "owner-user",
					},
				},
			}

			sbs[i] = &toolchainv1alpha1.SpaceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wName,
					Namespace: ksns,
					Labels: map[string]string{
						toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
						toolchainv1alpha1.SpaceBindingSpaceLabelKey:            wName,
					},
				},
				Spec: toolchainv1alpha1.SpaceBindingSpec{
					MasterUserRecord: "owner-user",
					SpaceRole:        "admin",
					Space:            wName,
				},
			}
		}

		ee := make([]client.Object, len(ww)+len(sbs))
		for i, w := range ww {
			ee[i] = w
		}
		for i, sb := range sbs {
			ee[10+i] = sb
		}

		BeforeEach(func() {
			c = buildCache(wsns, ksns, ee...)
		})

		It("should be returned in list of owner's workspaces", func() {
			// when the list of workspaces owned by 'owner-user' is requested
			var iww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "owner-user", &iww)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			Expect(iww.Items).To(HaveLen(len(ww)))

			for _, w := range iww.Items {
				Expect(ww).To(ContainElement(
					Satisfy(func(z *workspacesv1alpha1.InternalWorkspace) bool {
						return z.Spec.DisplayName == w.Spec.DisplayName && z.Spec.Owner.JwtInfo.Sub == w.Spec.Owner.JwtInfo.Sub
					})))
			}
		})

		It("should NOT be returned in list of not-owner's workspaces", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "not-owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).To(BeEmpty())
		})
	})

	When("workspace is created outside monitored namespaces", func() {
		BeforeEach(func() {
			c = buildCache(wsns, ksns,
				&workspacesv1alpha1.InternalWorkspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-ws",
						Namespace: "not-monitored",
					},
					Status: workspacesv1alpha1.InternalWorkspaceStatus{
						Owner: workspacesv1alpha1.UserInfoStatus{
							Username: "owner-user",
						},
						Space: workspacesv1alpha1.SpaceInfo{
							Name: "space",
						},
					},
				},
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            "owner-ws",
						},
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
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).To(BeEmpty())
		})
	})

	// workspace shared with user
	When("workspace is shared with other users", func() {
		BeforeEach(func() {
			wName := "owner-ws"
			w := &workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName(wName),
					Namespace: wsns,
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					DisplayName: wName,
				},
				Status: workspacesv1alpha1.InternalWorkspaceStatus{
					Owner: workspacesv1alpha1.UserInfoStatus{
						Username: "owner-user",
					},
					Space: workspacesv1alpha1.SpaceInfo{
						IsHome: false,
						Name:   "space",
					},
				},
			}
			c = buildCache(wsns, ksns,
				w,
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "owner-user",
						SpaceRole:        "admin",
						Space:            wName,
					},
				},
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "other-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "other-user",
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            w.Name,
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "other-user",
						SpaceRole:        "viewer",
						Space:            "space",
					},
				},
			)
		})

		It("is returned in other-user's list", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "other-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).To(HaveLen(1))
			Expect(ww.Items[0].Spec.DisplayName).To(Equal("owner-ws"))
			Expect(ww.Items[0].Status.Owner.Username).To(Equal("owner-user"))
			Expect(ww.Items[0].Name).ToNot(Equal("owner-ws"))
			Expect(ww.Items[0].Namespace).ToNot(Equal("owner-user"))
		})
	})

	// community workspace
	When("workspace is flagged as community", func() {
		BeforeEach(func() {
			wName := "owner-ws"

			w := &workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName(wName),
					Namespace: wsns,
					Labels: map[string]string{
						cache.LabelWorkspaceVisibility: string(workspacesv1alpha1.InternalWorkspaceVisibilityCommunity),
					},
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Visibility:  workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
					DisplayName: wName,
				},
				Status: workspacesv1alpha1.InternalWorkspaceStatus{
					Owner: workspacesv1alpha1.UserInfoStatus{
						Username: "owner-user",
					},
					Space: workspacesv1alpha1.SpaceInfo{
						IsHome: false,
						Name:   "space",
					},
				},
			}
			c = buildCache(wsns, ksns,
				w,
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            w.Name,
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "owner-user",
						SpaceRole:        "admin",
						Space:            "space",
					},
				},
			)
		})

		It("is returned in other-user's list", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "other-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).To(HaveLen(1))
			Expect(ww.Items[0].Spec.DisplayName).To(Equal("owner-ws"))
			Expect(ww.Items[0].Status.Owner.Username).To(Equal("owner-user"))
			Expect(ww.Items[0].Name).ToNot(Equal("owner-ws"))
			Expect(ww.Items[0].Namespace).ToNot(Equal("owner-user"))
		})
	})
})
