package internalworkspace_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

	ownerSub := "owner-sub"
	workspaceName := "workspace"
	workspacesNamespace := "workspaces-system"
	kubesawNamespace := "toolchain-host-operator"
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

	})

	When("Workspace is not found", func() {
		BeforeEach(func() {
			clientBuilder = fake.NewClientBuilder().WithScheme(scheme)
		})

		It("assumes that was deleted and does nothing", func() {
			// given
			r = buildReconciler()
			key := client.ObjectKeyFromObject(&workspace)

			// when
			res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})

			// then
			Expect(res).To(Equal(ctrl.Result{}))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("the Owner's UserSignup doesn't exist", func() {
		It("sets Ready condition to False", func() {
			// given
			userSignup := toolchainv1alpha1.UserSignup{
				ObjectMeta: corev1.ObjectMeta{Name: "not-matching-user", Namespace: kubesawNamespace},
				Spec: toolchainv1alpha1.UserSignupSpec{
					IdentityClaims: toolchainv1alpha1.IdentityClaimsEmbedded{
						PropagatedClaims: toolchainv1alpha1.PropagatedClaims{
							Sub: "not-me",
						},
					},
				},
			}
			clientBuilder = clientBuilder.
				WithObjects(&workspace, &userSignup).
				WithStatusSubresource(&workspace)

			r = buildReconciler()
			key := client.ObjectKeyFromObject(&workspace)

			// when
			res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})

			// then
			Expect(res).To(Equal(ctrl.Result{}))
			Expect(err).NotTo(HaveOccurred())

			w := workspacesv1alpha1.InternalWorkspace{}
			Expect(r.Get(ctx, key, &w)).To(Succeed())
			Expect(w.Status.Conditions).NotTo(BeEmpty())
			Expect(w.Status.Conditions).To(Satisfy(func(cc []metav1.Condition) bool {
				return meta.IsStatusConditionFalse(cc, workspacesv1alpha1.ConditionTypeReady)
			}))
		})
	})

	When("the Owner's UserSignup exists", func() {
		owner := toolchainv1alpha1.UserSignup{
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

		BeforeEach(func() {
			clientBuilder = clientBuilder.WithObjects(&owner)
		})

		It("sets the Ready condition to True", func() {
			// given
			key := client.ObjectKeyFromObject(&workspace)
			r = buildReconciler()

			// when
			res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})

			// then
			Expect(res).To(Equal(ctrl.Result{}))
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
				Name:      fmt.Sprintf("%s-community", workspace.Name),
				Namespace: kubesawNamespace,
			},
		}

		When("the Visibility is set to private", func() {
			BeforeEach(func() {
				workspace.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityPrivate
				clientBuilder = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&workspace).
					WithStatusSubresource(&workspace)
			})

			It("deletes the existing community SpaceBinding", func() {
				// given
				clientBuilder = clientBuilder.WithObjects(&communitySpaceBinding)
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(res).To(Equal(ctrl.Result{}))

				err = r.Get(ctx, client.ObjectKeyFromObject(&communitySpaceBinding), &toolchainv1alpha1.SpaceBinding{})
				Expect(err).To(MatchError(errors.IsNotFound, "IsNotFound error expected"))
			})

			It("is fine if the community SpaceBinding does not exists", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(res).To(Equal(ctrl.Result{}))

				err = r.Get(ctx, client.ObjectKeyFromObject(&communitySpaceBinding), &toolchainv1alpha1.SpaceBinding{})
				Expect(err).To(MatchError(errors.IsNotFound, "IsNotFound error expected"))
			})
		})

		When("the Visibility is set to community", func() {
			BeforeEach(func() {
				workspace.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityCommunity
				clientBuilder = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&workspace).
					WithStatusSubresource(&workspace)
			})

			It("creates the community SpaceBinding", func() {
				// given
				r = buildReconciler()
				key := client.ObjectKeyFromObject(&workspace)

				// when
				res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(res).To(Equal(ctrl.Result{}))

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
				res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(res).To(Equal(ctrl.Result{}))

				sb := toolchainv1alpha1.SpaceBinding{}
				Expect(r.Get(ctx, client.ObjectKeyFromObject(&communitySpaceBinding), &sb)).To(Succeed())
				Expect(sb.Spec.MasterUserRecord).To(Equal(toolchainv1alpha1.KubesawAuthenticatedUsername))
				Expect(sb.Spec.Space).To(Equal(workspace.Name))
				Expect(sb.Spec.SpaceRole).To(Equal("viewer"))
			})
		})

	})
})
