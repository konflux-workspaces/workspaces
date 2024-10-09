package internalworkspace_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/operator/internal/controller/internalworkspace"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

var _ = Describe("InternalWorkspaceController", func() {
	var clientBuilder *fake.ClientBuilder
	var r internalworkspace.WorkspaceReconciler
	var ctx context.Context
	var scheme *runtime.Scheme

	var workspace workspacesv1alpha1.InternalWorkspace
	var owner toolchainv1alpha1.UserSignup
	var space toolchainv1alpha1.Space

	ownerSub := "owner-sub"
	workspaceName := "workspace"
	workspacesNamespace := "workspaces-system"
	kubesawNamespace := "toolchain-host-operator"

	buildReconciler := func() internalworkspace.WorkspaceReconciler {
		return internalworkspace.WorkspaceReconciler{
			Client:              clientBuilder.Build(),
			Scheme:              scheme,
			KubesawNamespace:    kubesawNamespace,
			WorkspacesNamespace: workspacesNamespace,
		}
	}

	BeforeEach(func() {
		ctx = context.TODO()

		scheme = runtime.NewScheme()
		Expect(workspacesv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(toolchainv1alpha1.AddToScheme(scheme)).To(Succeed())

		clientBuilder = fake.NewClientBuilder().WithScheme(scheme)

		owner = toolchainv1alpha1.UserSignup{
			ObjectMeta: corev1.ObjectMeta{
				Name:      "owner",
				Namespace: kubesawNamespace,
			},
			Spec: toolchainv1alpha1.UserSignupSpec{
				IdentityClaims: toolchainv1alpha1.IdentityClaimsEmbedded{
					PropagatedClaims: toolchainv1alpha1.PropagatedClaims{
						Sub: ownerSub,
					},
				},
			},
			Status: toolchainv1alpha1.UserSignupStatus{
				CompliantUsername: "owner",
			},
		}
		workspace = workspacesv1alpha1.InternalWorkspace{
			ObjectMeta: corev1.ObjectMeta{
				Namespace: workspacesNamespace,
				Name:      workspaceName,
			},
			Spec: workspacesv1alpha1.InternalWorkspaceSpec{
				Visibility: workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
				Owner: workspacesv1alpha1.UserInfo{
					JwtInfo: workspacesv1alpha1.JwtInfo{
						Sub:    ownerSub,
						Email:  "owner@email.com",
						UserId: "owner-userid",
					},
				},
			},
		}
		space = toolchainv1alpha1.Space{
			ObjectMeta: corev1.ObjectMeta{
				Name:      workspace.Name,
				Namespace: kubesawNamespace,
			},
			Status: toolchainv1alpha1.SpaceStatus{
				TargetCluster: "target-cluster",
			},
		}
	})

	Context("Workspace is not found", func() {
		It("does nothing", func() {
			// given
			r = buildReconciler()
			key := client.ObjectKeyFromObject(&workspace)

			// when
			res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

			// then
			Expect(res).To(BeZero())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Workspace exists", func() {
		BeforeEach(func() {
			clientBuilder = clientBuilder.
				WithObjects(&workspace).
				WithStatusSubresource(&workspace)
		})

		When("neither Owner's UserSignup nor Space exist", func() {
			It("sets Ready condition to False", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).NotTo(HaveOccurred())

				w := workspacesv1alpha1.InternalWorkspace{}
				Expect(r.Get(ctx, key, &w)).To(Succeed())

				c := meta.FindStatusCondition(w.Status.Conditions, workspacesv1alpha1.ConditionTypeReady)
				Expect(c).NotTo(BeNil())
				Expect(c.Status).To(Equal(metav1.ConditionFalse))
				Expect(c.Reason).To(Equal(workspacesv1alpha1.ConditionReasonOwnerNotFound))
			})
		})

		When("Retrieval of Owner's UserSignup and Space fail", Label("none"), func() {
			errListUserSignup := fmt.Errorf("unexpected error retrieving UserSignup list")
			errGetSpace := fmt.Errorf("unexpected error Space")

			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&space).WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if _, ok := obj.(*toolchainv1alpha1.Space); ok {
							return errGetSpace
						}
						return client.Get(ctx, key, obj, opts...)
					},
					List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
						if _, ok := list.(*toolchainv1alpha1.UserSignupList); ok {
							return errListUserSignup
						}
						return client.List(ctx, list, opts...)
					},
				})
			})

			It("joined errors are forwarded with no changes", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).To(MatchError(errors.Join(errListUserSignup, errGetSpace)))

				w := workspacesv1alpha1.InternalWorkspace{}
				err = r.Get(ctx, key, &w)
				Expect(err).ToNot(HaveOccurred())
				Expect(w).To(BeEquivalentTo(workspace))
			})
		})

		When("Status update fails", Label("none"), func() {
			errUpdateStatus := fmt.Errorf("unexpected error updating InternalWorkspace status")

			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&space).WithInterceptorFuncs(interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
						if _, ok := obj.(*workspacesv1alpha1.InternalWorkspace); ok && subResourceName == "status" {
							return errUpdateStatus
						}
						return client.Status().Update(ctx, obj, opts...)
					},
				})
			})

			It("joined errors are forwarded with no changes", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).To(MatchError(errors.Join(errUpdateStatus)))

				w := workspacesv1alpha1.InternalWorkspace{}
				err = r.Get(ctx, key, &w)
				Expect(err).ToNot(HaveOccurred())
				Expect(w).To(BeEquivalentTo(workspace))
			})
		})

		When("Retrieval of Space fails", Label("none"), func() {
			errGetSpace := fmt.Errorf("unexpected error Space")

			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&space).WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if _, ok := obj.(*toolchainv1alpha1.Space); ok {
							return errGetSpace
						}
						return client.Get(ctx, key, obj, opts...)
					},
				})
			})

			It("joined errors are forwarded with owner user's changes", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).To(MatchError(errors.Join(errGetSpace)))

				w := workspacesv1alpha1.InternalWorkspace{}
				err = r.Get(ctx, key, &w)
				Expect(err).ToNot(HaveOccurred())
				Expect(w).NotTo(BeEquivalentTo(workspace))
				Expect(w.Status.Space.Name).To(Equal(space.Name))
			})
		})

		When("Retrieval of Owner's UserSignup fails", Label("none"), func() {
			errListUserSignup := fmt.Errorf("unexpected error retrieving UserSignup list")

			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&space).WithInterceptorFuncs(interceptor.Funcs{
					List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
						if _, ok := list.(*toolchainv1alpha1.UserSignupList); ok {
							return errListUserSignup
						}
						return client.List(ctx, list, opts...)
					},
				})
			})

			It("joined errors are forwarded with space changes", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).To(MatchError(errors.Join(errListUserSignup)))

				w := workspacesv1alpha1.InternalWorkspace{}
				err = r.Get(ctx, key, &w)
				Expect(err).ToNot(HaveOccurred())
				Expect(w).NotTo(BeEquivalentTo(workspace))
				Expect(w.Status.Space.TargetCluster).To(Equal(space.Status.TargetCluster))
			})
		})

		When("Owner's UserSignup does not exist and Space does", func() {
			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&space)
			})

			It("sets Ready condition to False", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).NotTo(HaveOccurred())

				w := workspacesv1alpha1.InternalWorkspace{}
				Expect(r.Get(ctx, key, &w)).To(Succeed())

				c := meta.FindStatusCondition(w.Status.Conditions, workspacesv1alpha1.ConditionTypeReady)
				Expect(c).NotTo(BeNil())
				Expect(c.Status).To(Equal(metav1.ConditionFalse))
				Expect(c.Reason).To(Equal(workspacesv1alpha1.ConditionReasonOwnerNotFound))
			})
		})

		When("the Owner's UserSignup exists and Space does not", func() {
			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&owner)
			})

			It("sets Ready condition to False", func() {
				// given
				key := client.ObjectKeyFromObject(&workspace)
				r = buildReconciler()

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).NotTo(HaveOccurred())

				w := workspacesv1alpha1.InternalWorkspace{}
				Expect(r.Get(ctx, key, &w)).To(Succeed())
				Expect(w.Status.Conditions).NotTo(BeEmpty())

				c := meta.FindStatusCondition(w.Status.Conditions, workspacesv1alpha1.ConditionTypeReady)
				Expect(c).NotTo(BeNil())
				Expect(c.Status).To(Equal(metav1.ConditionFalse))
				Expect(c.Reason).To(Equal(workspacesv1alpha1.ConditionReasonSpaceNotFound))
			})
		})

		When("the Owner's UserSignup and Space exist", func() {
			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&owner, &space)
			})

			It("sets Ready condition to True", func() {
				// given
				key := client.ObjectKeyFromObject(&workspace)
				r = buildReconciler()

				// when
				res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

				// then
				Expect(res).To(BeZero())
				Expect(err).NotTo(HaveOccurred())

				w := workspacesv1alpha1.InternalWorkspace{}
				Expect(r.Get(ctx, key, &w)).To(Succeed())
				Expect(w.Status.Conditions).NotTo(BeEmpty())
				Expect(w.Status.Conditions).To(Satisfy(func(cc []metav1.Condition) bool {
					return meta.IsStatusConditionTrue(cc, workspacesv1alpha1.ConditionTypeReady)
				}))
			})
		})

		Context("community SpaceBinding management", func() {
			communitySpaceBinding := toolchainv1alpha1.SpaceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-community", workspaceName),
					Namespace: kubesawNamespace,
				},
			}

			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&owner, &space)
			})

			When("the Visibility is set to private", func() {
				BeforeEach(func() {
					workspace.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityPrivate
				})

				It("deletes the existing community SpaceBinding", func() {
					// given
					clientBuilder = clientBuilder.WithObjects(&communitySpaceBinding)
					r = buildReconciler()
					key := client.ObjectKeyFromObject(&workspace)

					// when
					res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(res).To(BeZero())

					err = r.Get(ctx, client.ObjectKeyFromObject(&communitySpaceBinding), &toolchainv1alpha1.SpaceBinding{})
					Expect(err).To(MatchError(kerrors.IsNotFound, "IsNotFound error expected"))
				})

				It("is fine if the community SpaceBinding does not exists", func() {
					// given
					r = buildReconciler()
					key := client.ObjectKeyFromObject(&workspace)

					// when
					res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(res).To(BeZero())

					err = r.Get(ctx, client.ObjectKeyFromObject(&communitySpaceBinding), &toolchainv1alpha1.SpaceBinding{})
					Expect(err).To(MatchError(kerrors.IsNotFound, "IsNotFound error expected"))
				})
			})

			When("the Visibility is set to community", func() {
				BeforeEach(func() {
					workspace.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityCommunity
				})

				It("creates the community SpaceBinding", func() {
					// given
					r = buildReconciler()
					key := client.ObjectKeyFromObject(&workspace)

					// when
					res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(res).To(BeZero())

					sb := toolchainv1alpha1.SpaceBinding{}
					Expect(r.Get(ctx, client.ObjectKeyFromObject(&communitySpaceBinding), &sb)).To(Succeed())
					Expect(sb.Spec.MasterUserRecord).To(Equal(toolchainv1alpha1.KubesawAuthenticatedUsername))
					Expect(sb.Spec.Space).To(Equal(workspace.Name))
					Expect(sb.Spec.SpaceRole).To(Equal("viewer"))
				})

				It("updates the existing community SpaceBinding", func() {
					// given
					clientBuilder = clientBuilder.WithObjects(&communitySpaceBinding)
					r = buildReconciler()
					key := client.ObjectKeyFromObject(&workspace)

					// when
					res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})

					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(res).To(BeZero())

					sb := toolchainv1alpha1.SpaceBinding{}
					Expect(r.Get(ctx, client.ObjectKeyFromObject(&communitySpaceBinding), &sb)).To(Succeed())
					Expect(sb.Spec.MasterUserRecord).To(Equal(toolchainv1alpha1.KubesawAuthenticatedUsername))
					Expect(sb.Spec.Space).To(Equal(workspace.Name))
					Expect(sb.Spec.SpaceRole).To(Equal("viewer"))
				})
			})
		})
	})
})
