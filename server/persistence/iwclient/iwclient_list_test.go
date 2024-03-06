package iwclient_test

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"

	icache "github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
)

var _ = Describe("List", func() {
	var ctx context.Context
	var c *iwclient.Client

	ksns := "kubesaw-namespace"
	wsns := "workspaces-namespace"

	ownerName := "owner-user"
	uuidSub := uuid.New()
	ownerUserInfo := workspacesv1alpha1.UserInfo{
		JWTInfo: workspacesv1alpha1.JwtInfo{
			Username: ownerName,
			Sub:      fmt.Sprintf("f:%s:%s", uuidSub, ownerName),
			Email:    fmt.Sprintf("%s@domain.com", ownerName),
		},
	}

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
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Owner: ownerUserInfo,
				},
			})

		It("should not return the workspace in list", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "owner", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(BeEmpty())
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
					MasterUserRecord: ownerName,
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
			Expect(ww.Items).Should(BeEmpty())
		})

	})

	When("one valid workspace exists", func() {
		wName := "owner-ws"
		w := &workspacesv1alpha1.InternalWorkspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateName(wName),
				Namespace: wsns,
				Labels: map[string]string{
					workspacesv1alpha1.LabelDisplayName: wName,
				},
			},
			Spec: workspacesv1alpha1.InternalWorkspaceSpec{
				Owner: ownerUserInfo,
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
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: ownerName,
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            w.Name,
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: ownerName,
						SpaceRole:        "admin",
						Space:            w.Name,
					},
				},
			)
		})

		It("should be returned in list of owner's workspaces", func() {
			// when the list of workspaces owned by 'owner-user' is requested
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, ownerName, &ww)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			Expect(ww.Items).Should(HaveLen(1))
			Expect(ww.Items[0].GetLabels()).ShouldNot(
				And(
					BeEmpty(),
					HaveKeyWithValue(workspacesv1alpha1.LabelDisplayName, "owner-ws"),
				),
			)
			Expect(ww.Items[0].Name).ShouldNot(Equal("owner-ws"))
			Expect(ww.Items[0].Namespace).ShouldNot(Equal(ownerName))
		})

		It("should NOT be returned in list of not-owner's workspaces", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "not-owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(BeEmpty())
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
					Labels: map[string]string{
						workspacesv1alpha1.LabelDisplayName: wName,
					},
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Owner: ownerUserInfo,
				},
			}

			sbs[i] = &toolchainv1alpha1.SpaceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wName,
					Namespace: ksns,
					Labels: map[string]string{
						toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: ownerName,
						toolchainv1alpha1.SpaceBindingSpaceLabelKey:            wName,
					},
				},
				Spec: toolchainv1alpha1.SpaceBindingSpec{
					MasterUserRecord: ownerName,
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
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, ownerName, &ww)
			Expect(err).NotTo(HaveOccurred())

			// then the list contains just the 'owner-ws' workspace
			wwi := ww.Items
			Expect(wwi).Should(HaveLen(len(wwi)))

			for _, w := range wwi {
				sw := slices.ContainsFunc(wwi, func(z workspacesv1alpha1.InternalWorkspace) bool {
					zll, wll := z.GetLabels(), w.GetLabels()
					return z.Spec.Owner.JWTInfo.Username != w.Spec.Owner.JWTInfo.Username &&
						zll[workspacesv1alpha1.LabelDisplayName] == wll[workspacesv1alpha1.LabelDisplayName]
				})
				Expect(sw).Should(BeTrue())
			}
		})

		It("should NOT be returned in list of not-owner's workspaces", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, "not-owner-user", &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(BeEmpty())
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
					Spec: workspacesv1alpha1.InternalWorkspaceSpec{
						Owner: ownerUserInfo,
					},
				},
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: ownerName,
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            "owner-ws",
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: ownerName,
						SpaceRole:        "admin",
						Space:            "owner-ws",
					},
				},
			)
		})

		It("is not returned in list", func() {
			// when
			var ww workspacesv1alpha1.InternalWorkspaceList
			err := c.ListAsUser(ctx, ownerName, &ww)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(ww.Items).Should(BeEmpty())
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
					Labels: map[string]string{
						workspacesv1alpha1.LabelDisplayName: wName,
					},
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Owner: ownerUserInfo,
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
						MasterUserRecord: ownerName,
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
						Space:            w.Name,
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
			Expect(ww.Items).Should(HaveLen(1))
			Expect(ww.Items[0].GetLabels()).ShouldNot(
				And(
					BeEmpty(),
					HaveKeyWithValue(workspacesv1alpha1.LabelDisplayName, "owner-ws"),
				),
			)
			Expect(ww.Items[0].Name).ShouldNot(Equal("owner-ws"))
			Expect(ww.Items[0].Namespace).ShouldNot(Equal(ownerName))
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
						workspacesv1alpha1.LabelDisplayName: wName,
						icache.LabelWorkspaceVisibility:     string(workspacesv1alpha1.InternalWorkspaceVisibilityCommunity),
					},
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					Owner:      ownerUserInfo,
					Visibility: workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
				},
			}
			c = buildCache(wsns, ksns,
				w,
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: ownerName,
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            w.Name,
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: ownerName,
						SpaceRole:        "admin",
						Space:            w.Name,
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
			Expect(ww.Items).Should(HaveLen(1))
			Expect(ww.Items[0].GetLabels()).ShouldNot(
				And(
					BeEmpty(),
					HaveKeyWithValue(workspacesv1alpha1.LabelDisplayName, "owner-ws"),
				),
			)
			Expect(ww.Items[0].Name).ShouldNot(Equal("owner-ws"))
			Expect(ww.Items[0].Namespace).ShouldNot(Equal(ownerName))
		})
	})
})
