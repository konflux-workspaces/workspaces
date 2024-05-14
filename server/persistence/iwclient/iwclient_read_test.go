package iwclient_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"

	icache "github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
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

	// TODO: add workspace with invalid labels

	When("no SpaceBinding exists for a workspace", func() {
		// given
		c = buildCache(wsns, ksns,
			&workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName("no-space-binding"),
					Namespace: wsns,
					Labels: map[string]string{
						workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
						workspacesv1alpha1.LabelDisplayName:    "no-space-binding",
					},
				},
			})

		It("should not return the workspace in read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := iwclient.SpaceKey{Owner: "owner", Name: "no-space-binding"}
			err := c.GetAsUser(ctx, "owner", key, &w)

			// then
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, iwclient.ErrWorkspaceNotFound)).To(BeTrue())
		})
	})

	When("owner label is not set on workspace", func() {
		// given
		c = buildCache(wsns, ksns,
			&workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName("no-owner-label"),
					Namespace: wsns,
					Labels: map[string]string{
						workspacesv1alpha1.LabelDisplayName: "no-owner-label",
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
					Space:            "no-label",
				},
			},
		)

		It("should not return the workspace in read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := iwclient.SpaceKey{Owner: "owner", Name: "no-label"}
			err := c.GetAsUser(ctx, "owner", key, &w)

			// then
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, iwclient.ErrWorkspaceNotFound)).To(BeTrue())
		})
	})

	When("one valid workspace exists", func() {
		w := &workspacesv1alpha1.InternalWorkspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      generateName("owner-ws"),
				Namespace: wsns,
				Labels: map[string]string{
					workspacesv1alpha1.LabelDisplayName:    "owner-ws",
					workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
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
			)
		})

		It("should be returned in read", func() {
			// when
			var rw workspacesv1alpha1.InternalWorkspace
			key := iwclient.SpaceKey{Owner: "owner-user", Name: "owner-ws"}
			err := c.GetAsUser(ctx, "owner-user", key, &rw)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(*w).To(Equal(rw))
		})

		It("should NOT be returned in read of not-owner-user workspace", func() {
			// when
			rw := workspacesv1alpha1.InternalWorkspace{}
			key := iwclient.SpaceKey{Owner: "owner-user", Name: "owner-ws"}
			err := c.GetAsUser(ctx, "not-owner-user", key, &rw)

			// then
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, iwclient.ErrUnauthorized)).To(BeTrue())
			Expect(rw).To(BeZero())
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
						workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
						workspacesv1alpha1.LabelDisplayName:    wName,
					},
				},
			}
			sbs[i] = &toolchainv1alpha1.SpaceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("owner-sb-%d", i),
					Namespace: ksns,
					Labels: map[string]string{
						toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
						toolchainv1alpha1.SpaceBindingSpaceLabelKey:            ww[i].GetName(),
					},
				},
				Spec: toolchainv1alpha1.SpaceBindingSpec{
					MasterUserRecord: "owner-user",
					SpaceRole:        "admin",
					Space:            ww[i].GetName(),
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

		It("should be returned in read", func() {
			for _, w := range ww {
				wName := w.GetLabels()[workspacesv1alpha1.LabelDisplayName]

				// when
				var rw workspacesv1alpha1.InternalWorkspace
				key := iwclient.SpaceKey{Owner: "owner-user", Name: wName}
				err := c.GetAsUser(ctx, "owner-user", key, &rw)
				Expect(err).NotTo(HaveOccurred())

				// then
				Expect(*w).To(Equal(rw))
			}
		})

		It("should NOT be returned in read of not-owner-user workspace", func() {
			for _, w := range ww {
				wName := w.GetLabels()[workspacesv1alpha1.LabelDisplayName]

				// when
				rw := workspacesv1alpha1.InternalWorkspace{}
				key := iwclient.SpaceKey{Owner: "owner-user", Name: wName}
				err := c.GetAsUser(ctx, "not-owner-user", key, &rw)

				// then
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, iwclient.ErrUnauthorized)).To(BeTrue())
				Expect(rw).To(BeZero())
			}
		})
	})

	// TODO: disabled because we are not caching namespaces in internal/cache
	//
	// When("workspace is created outside monitored namespaces", func() {
	// 	BeforeEach(func() {
	// 		wName := "owner-ws"
	// 		fakeGenWName := wName + "-jjdjk"
	// 		c = buildCache(wsns, ksns,
	// 			&workspacesv1alpha1.InternalWorkspace{
	// 				ObjectMeta: metav1.ObjectMeta{
	// 					Name:      fakeGenWName,
	// 					Namespace: "not-monitored",
	// 					Labels: map[string]string{
	// 						workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
	// 						workspacesv1alpha1.LabelDisplayName:    wName,
	// 					},
	// 				},
	// 			},
	// 			&toolchainv1alpha1.SpaceBinding{
	// 				ObjectMeta: metav1.ObjectMeta{
	// 					Name:      "owner-sb",
	// 					Namespace: ksns,
	// 					Labels: map[string]string{
	// 						toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
	// 						toolchainv1alpha1.SpaceBindingSpaceLabelKey:            fakeGenWName,
	// 					},
	// 				},
	// 				Spec: toolchainv1alpha1.SpaceBindingSpec{
	// 					MasterUserRecord: "owner-user",
	// 					SpaceRole:        "admin",
	// 					Space:            fakeGenWName,
	// 				},
	// 			},
	// 		)
	// 	})

	// 	It("is not returned in read", func() {
	// 		// when
	// 		var w workspacesv1alpha1.InternalWorkspace
	// 		key := iwclient.SpaceKey{Owner: "owner-user", Name: "owner-ws"}
	// 		err := c.GetAsUser(ctx, "owner-user", key, &w)

	// 		// then
	// 		Expect(err).To(HaveOccurred())
	// 		Expect(errors.Is(err, iwclient.ErrUnauthorized)).To(BeTrue())
	// 	})

	// 	// TODO: iwclient.ErrMoreThanOneFound
	// })

	// workspace shared with user
	When("workspace is shared with other users", func() {
		var expectedWorkspace workspacesv1alpha1.InternalWorkspace
		wName := "owner-ws"

		BeforeEach(func() {
			expectedWorkspace = workspacesv1alpha1.InternalWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      generateName(wName),
					Namespace: wsns,
					Labels: map[string]string{
						workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
						workspacesv1alpha1.LabelDisplayName:    wName,
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
							toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: "owner-user",
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            expectedWorkspace.GetName(),
						},
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
							toolchainv1alpha1.SpaceBindingSpaceLabelKey:            expectedWorkspace.GetName(),
						},
					},
					Spec: toolchainv1alpha1.SpaceBindingSpec{
						MasterUserRecord: "other-user",
						SpaceRole:        "viewer",
						Space:            wName,
					},
				},
			)
		})

		It("is returned in read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := iwclient.SpaceKey{Owner: "owner-user", Name: wName}
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
				Labels: map[string]string{
					workspacesv1alpha1.LabelWorkspaceOwner: "owner-user",
					icache.LabelWorkspaceVisibility:        string(workspacesv1alpha1.InternalWorkspaceVisibilityCommunity),
					workspacesv1alpha1.LabelDisplayName:    wName,
				},
			},
			Spec: workspacesv1alpha1.InternalWorkspaceSpec{
				Visibility: workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
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
			)
		})

		It("is returned in other-user's read", func() {
			// when
			var w workspacesv1alpha1.InternalWorkspace
			key := iwclient.SpaceKey{Owner: "owner-user", Name: wName}
			err := c.GetAsUser(ctx, "other-user", key, &w)
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(w).To(Equal(expectedWorkspace))
		})
	})
})
