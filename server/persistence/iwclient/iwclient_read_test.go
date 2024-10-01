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

	"github.com/konflux-workspaces/workspaces/server/persistence/clientinterface"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
)

var _ = Describe("Read", func() {
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
					Name:      generateName("no-space-binding"),
					Namespace: wsns,
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					DisplayName: "no-space-binding",
				},
				Status: workspacesv1alpha1.InternalWorkspaceStatus{
					Owner: workspacesv1alpha1.UserInfoStatus{
						Username: "owner-user",
					},
				},
			})

		It("should not return the workspace in read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := clientinterface.SpaceKey{Owner: "owner", Name: "no-space-binding"}
			err := c.GetAsUser(ctx, "owner", key, &w)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(iwclient.ErrWorkspaceNotFound))
		})
	})

	When("owner is not set on workspace", func() {
		// given
		c = buildCache(wsns, ksns,
			&workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName("no-owner-label"),
					Namespace: wsns,
					Labels:    map[string]string{},
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					DisplayName: "no-owner-label",
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

		It("should not return the workspace in read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := clientinterface.SpaceKey{Owner: "owner", Name: "no-label"}
			err := c.GetAsUser(ctx, "owner", key, &w)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(iwclient.ErrWorkspaceNotFound))
		})
	})

	When("one valid workspace exists", func() {
		w := &workspacesv1alpha1.InternalWorkspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateName("owner-ws"),
				Namespace: wsns,
			},
			Spec: workspacesv1alpha1.InternalWorkspaceSpec{
				DisplayName: "owner-ws",
				Owner: workspacesv1alpha1.UserInfo{
					JwtInfo: workspacesv1alpha1.JwtInfo{},
				},
			},
			Status: workspacesv1alpha1.InternalWorkspaceStatus{
				Owner: workspacesv1alpha1.UserInfoStatus{
					Username: "owner-user",
				},
				Space: workspacesv1alpha1.SpaceInfo{
					IsHome: true,
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
						Space:            w.Name,
					},
				},
				&toolchainv1alpha1.UserSignup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      w.Status.Owner.Username,
						Namespace: ksns,
					},
					Status: toolchainv1alpha1.UserSignupStatus{
						CompliantUsername: w.Status.Owner.Username,
					},
				},
			)
		})

		It("should be returned in read", func() {
			// when
			var rw workspacesv1alpha1.InternalWorkspace
			key := clientinterface.SpaceKey{Owner: "owner-user", Name: "owner-ws"}
			err := c.GetAsUser(ctx, "owner-user", key, &rw)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(*w).To(Equal(rw))
		})

		It("should NOT be returned in read of not-owner-user workspace", func() {
			// when
			rw := workspacesv1alpha1.InternalWorkspace{}
			key := clientinterface.SpaceKey{Owner: "owner-user", Name: "owner-ws"}
			err := c.GetAsUser(ctx, "not-owner-user", key, &rw)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(iwclient.ErrUnauthorized))
			Expect(rw).To(BeZero())
		})
	})

	When("more than one valid workspace exist", func() {
		var ww []*workspacesv1alpha1.InternalWorkspace

		BeforeEach(func() {
			ee := make([]client.Object, 30)
			ww = make([]*workspacesv1alpha1.InternalWorkspace, 10)
			sbs := make([]*toolchainv1alpha1.SpaceBinding, len(ww))
			uu := make([]*toolchainv1alpha1.UserSignup, len(ww))

			for i := 0; i < len(ww); i++ {
				wName := fmt.Sprintf("owner-ws-%d", i)
				ww[i] = &workspacesv1alpha1.InternalWorkspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      generateName(wName),
						Namespace: wsns,
					},
					Spec: workspacesv1alpha1.InternalWorkspaceSpec{
						DisplayName: wName,
						Owner: workspacesv1alpha1.UserInfo{
							JwtInfo: workspacesv1alpha1.JwtInfo{},
						},
					},
					Status: workspacesv1alpha1.InternalWorkspaceStatus{
						Owner: workspacesv1alpha1.UserInfoStatus{
							Username: fmt.Sprintf("owner-user-%d", i),
						},
					},
				}
				sbs[i] = &toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("owner-sb-%d", i),
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: fmt.Sprintf("owner-user-%d", i),
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            ww[i].GetName(),
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: fmt.Sprintf("owner-user-%d", i),
						SpaceRole:        "admin",
						Space:            ww[i].GetName(),
					},
				}
				uu[i] = &toolchainv1alpha1.UserSignup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ww[i].Status.Owner.Username,
						Namespace: ksns,
					},
					Status: toolchainv1alpha1.UserSignupStatus{
						CompliantUsername: ww[i].Status.Owner.Username,
					},
				}
			}

			for i, w := range ww {
				ee[i] = w
			}
			for i, sb := range sbs {
				ee[10+i] = sb
			}
			for i, u := range uu {
				ee[20+i] = u
			}

			c = buildCache(wsns, ksns, ee...)
		})

		It("should be returned in read", func() {
			for _, w := range ww {
				// when
				var rw workspacesv1alpha1.InternalWorkspace
				key := clientinterface.SpaceKey{Owner: w.Status.Owner.Username, Name: w.Spec.DisplayName}
				err := c.GetAsUser(ctx, w.Status.Owner.Username, key, &rw)
				Expect(err).NotTo(HaveOccurred())

				// then
				Expect(*w).To(Equal(rw))
			}
		})

		It("should NOT be returned in read of not-owner-user workspace", func() {
			for _, w := range ww {
				// when
				rw := workspacesv1alpha1.InternalWorkspace{}
				key := clientinterface.SpaceKey{Owner: w.Status.Owner.Username, Name: w.Spec.DisplayName}
				err := c.GetAsUser(ctx, "not-owner-user", key, &rw)

				// then
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(iwclient.ErrUnauthorized))
				Expect(rw).To(BeZero())
			}
		})
	})

	// workspace shared with user
	When("workspace is shared with other users", func() {
		var expectedWorkspace workspacesv1alpha1.InternalWorkspace
		wName := "owner-ws"

		BeforeEach(func() {
			expectedWorkspace = workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName(wName),
					Namespace: wsns,
				},
				Spec: workspacesv1alpha1.InternalWorkspaceSpec{
					DisplayName: wName,
					Owner: workspacesv1alpha1.UserInfo{
						JwtInfo: workspacesv1alpha1.JwtInfo{},
					},
				},
				Status: workspacesv1alpha1.InternalWorkspaceStatus{
					Space: workspacesv1alpha1.SpaceInfo{
						IsHome: true,
						Name:   "space",
					},
					Owner: workspacesv1alpha1.UserInfoStatus{
						Username: "owner-user",
					},
				},
			}
			c = buildCache(wsns, ksns,
				&expectedWorkspace,
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: expectedWorkspace.Status.Owner.Username,
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            expectedWorkspace.GetName(),
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: expectedWorkspace.Status.Owner.Username,
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
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            expectedWorkspace.GetName(),
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "other-user",
						SpaceRole:        "viewer",
						Space:            wName,
					},
				},
				&toolchainv1alpha1.UserSignup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      expectedWorkspace.Status.Owner.Username,
						Namespace: ksns,
					},
					Status: toolchainv1alpha1.UserSignupStatus{
						CompliantUsername: expectedWorkspace.Status.Owner.Username,
					},
				},
			)
		})

		It("is returned in read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := clientinterface.SpaceKey{Owner: "owner-user", Name: wName}
			err := c.GetAsUser(ctx, "owner-user", key, &w)

			// then
			Expect(err).NotTo(HaveOccurred())
			Expect(w).To(Equal(expectedWorkspace))
		})
	})

	// community workspace
	When("workspace is flagged as community", func() {
		wName := "owner-ws"
		expectedWorkspace := workspacesv1alpha1.InternalWorkspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateName(wName),
				Namespace: wsns,
			},
			Spec: workspacesv1alpha1.InternalWorkspaceSpec{
				Visibility:  workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
				DisplayName: wName,
				Owner: workspacesv1alpha1.UserInfo{
					JwtInfo: workspacesv1alpha1.JwtInfo{},
				},
			},
			Status: workspacesv1alpha1.InternalWorkspaceStatus{
				Owner: workspacesv1alpha1.UserInfoStatus{
					Username: "owner-user",
				},
				Space: workspacesv1alpha1.SpaceInfo{
					IsHome: true,
					Name:   "space",
				},
			},
		}

		BeforeEach(func() {
			c = buildCache(wsns, ksns,
				&expectedWorkspace,
				&toolchainv1alpha1.SpaceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owner-sb",
						Namespace: ksns,
						Labels: map[string]string{
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            expectedWorkspace.GetName(),
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "owner-user",
						SpaceRole:        "admin",
						Space:            expectedWorkspace.GetName(),
					},
				},
				&toolchainv1alpha1.UserSignup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      expectedWorkspace.Status.Owner.Username,
						Namespace: ksns,
					},
					Status: toolchainv1alpha1.UserSignupStatus{
						CompliantUsername: expectedWorkspace.Status.Owner.Username,
					},
				},
			)
		})

		It("is returned in other-user's read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := clientinterface.SpaceKey{Owner: "owner-user", Name: wName}
			err := c.GetAsUser(ctx, "other-user", key, &w)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(w).To(Equal(expectedWorkspace))
		})
	})
})
