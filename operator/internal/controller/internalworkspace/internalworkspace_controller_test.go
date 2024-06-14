package internalworkspace_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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

var _ = Describe("InternalworkspaceController", func() {
	var clientBuilder *fake.ClientBuilder
	var r internalworkspace.WorkspaceReconciler
	var ctx context.Context
	var scheme *runtime.Scheme
	var workspace workspacesv1alpha1.InternalWorkspace

	ownerSub := "owner-sub"
	workspaceName := "workspace"
	workspacesNamespace := "workspaces-system"
	kubesawNamespace := "toolchain-host-operator"

	BeforeEach(func() {
		ctx = context.TODO()

		scheme = runtime.NewScheme()
		Expect(workspacesv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(toolchainv1alpha1.AddToScheme(scheme)).To(Succeed())

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
		clientBuilder = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(&workspace).
			WithStatusSubresource(&workspace)
	})

	When("the Owner's UserSignup doesn't exist", func() {
		BeforeEach(func() {
			r = internalworkspace.WorkspaceReconciler{Client: clientBuilder.Build(), Scheme: scheme}
		})

		It("sets Ready condition to False", func() {
			// given
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
			r = internalworkspace.WorkspaceReconciler{Client: clientBuilder.Build(), Scheme: scheme}
		})

		It("sets the Ready condition to True", func() {
			// given
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
				return meta.IsStatusConditionTrue(cc, workspacesv1alpha1.ConditionTypeReady)
			}))
		})
	})
})
