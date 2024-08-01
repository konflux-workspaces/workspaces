package usersignup_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/konflux-workspaces/workspaces/operator/internal/controller/usersignup"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

const (
	workspacesNamespace = "workspaces-system"
	kubesawNamespace    = "toolchain-host-operator"
)

var _ = Describe("InternalWorkspaceController", func() {
	var clientBuilder *fake.ClientBuilder
	var ctx context.Context
	var scheme *runtime.Scheme
	var aliceUserSignup toolchainv1alpha1.UserSignup
	var aliceInternalWorkspace workspacesv1alpha1.InternalWorkspace

	BeforeEach(func() {
		aliceUserSignup = toolchainv1alpha1.UserSignup{
			ObjectMeta: corev1.ObjectMeta{
				Name:      "alice",
				Namespace: kubesawNamespace,
			},
			Spec: toolchainv1alpha1.UserSignupSpec{
				IdentityClaims: toolchainv1alpha1.IdentityClaimsEmbedded{
					PropagatedClaims: toolchainv1alpha1.PropagatedClaims{
						Sub:    "alice-sub",
						Email:  "alice@email.com",
						UserID: "alice-userid",
					},
				},
			},
			Status: toolchainv1alpha1.UserSignupStatus{
				HomeSpace: "alice-home-space",
			},
		}
		aliceInternalWorkspace = workspacesv1alpha1.InternalWorkspace{
			ObjectMeta: corev1.ObjectMeta{
				Namespace: workspacesNamespace,
				Name:      aliceUserSignup.Status.HomeSpace,
			},
			Spec: workspacesv1alpha1.InternalWorkspaceSpec{
				Visibility: workspacesv1alpha1.InternalWorkspaceVisibilityCommunity,
				Owner: workspacesv1alpha1.UserInfo{
					JwtInfo: workspacesv1alpha1.JwtInfo{
						Sub:    aliceUserSignup.Spec.IdentityClaims.Sub,
						Email:  aliceUserSignup.Spec.IdentityClaims.Email,
						UserId: aliceUserSignup.Spec.IdentityClaims.UserID,
					},
				},
			},
			Status: workspacesv1alpha1.InternalWorkspaceStatus{
				Space: workspacesv1alpha1.SpaceInfo{
					IsHome: true,
				},
				Owner: workspacesv1alpha1.UserInfoStatus{
					Username: aliceUserSignup.Name,
				},
			},
		}
	})

	reconcile := func(obj client.Object) (*usersignup.UserSignupReconciler, ctrl.Result, error) {
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}
		r := &usersignup.UserSignupReconciler{
			Client:              clientBuilder.Build(),
			Scheme:              scheme,
			WorkspacesNamespace: workspacesNamespace,
		}
		res, err := r.Reconcile(ctx, req)
		return r, res, err
	}

	BeforeEach(func() {
		ctx = context.TODO()

		scheme = runtime.NewScheme()
		Expect(workspacesv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(toolchainv1alpha1.AddToScheme(scheme)).To(Succeed())
		clientBuilder = fake.NewClientBuilder().WithScheme(scheme)
	})

	Context("UserSignup is not found", func() {
		When("InternalWorkspace exists for user's home space", func() {
			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&aliceInternalWorkspace)
			})

			It("deletes the InternalWorkspace", func() {
				// when
				r, res, err := reconcile(&aliceUserSignup)

				// then
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(BeZero())
				Expect(r.Get(ctx, client.ObjectKeyFromObject(&aliceInternalWorkspace), &workspacesv1alpha1.InternalWorkspace{})).
					To(MatchError(errors.IsNotFound, "expected NotFound error"))
			})

			It("forwards the error if deletion was not successful", func() {
				// given
				deleteErr := fmt.Errorf("test delete error")
				clientBuilder = clientBuilder.WithInterceptorFuncs(
					interceptor.Funcs{
						Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
							return deleteErr
						},
					},
				)

				// when
				_, res, err := reconcile(&aliceUserSignup)

				// then
				Expect(err).To(MatchError(deleteErr))
				Expect(res).To(BeZero())
			})
		})

		When("InternalWorkspace doesn't exist for user's home space", func() {
			It("returns without error", func() {
				// when
				_, res, err := reconcile(&aliceUserSignup)

				// then
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(BeZero())
			})
		})
	})

	Context("UserSignup is found", func() {
		BeforeEach(func() {
			clientBuilder = clientBuilder.WithObjects(&aliceUserSignup)
		})

		When("an InternalWorkspace already exists", func() {
			BeforeEach(func() {
				clientBuilder = clientBuilder.WithObjects(&aliceInternalWorkspace)
			})

			It("updates user's JWTInfo", func() {
				// given
				aliceUserSignup.Spec.IdentityClaims.Sub = "new-alice-sub"
				aliceUserSignup.Spec.IdentityClaims.UserID = "new-alice-userid"
				aliceUserSignup.Spec.IdentityClaims.Email = "new-alice@email.com"

				// when
				r, res, err := reconcile(&aliceUserSignup)

				// then
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(BeZero())

				iw := workspacesv1alpha1.InternalWorkspace{}
				Expect(r.Get(ctx, client.ObjectKeyFromObject(&aliceInternalWorkspace), &iw)).To(Succeed())
				Expect(iw.Spec.Owner.JwtInfo.Sub).To(Equal(aliceUserSignup.Spec.IdentityClaims.Sub))
				Expect(iw.Spec.Owner.JwtInfo.Email).To(Equal(aliceUserSignup.Spec.IdentityClaims.Email))
				Expect(iw.Spec.Owner.JwtInfo.UserId).To(Equal(aliceUserSignup.Spec.IdentityClaims.UserID))
			})

			It("forwards the error if update was not successful", func() {
				// given
				updateErr := fmt.Errorf("test update error")
				clientBuilder = clientBuilder.WithInterceptorFuncs(interceptor.Funcs{
					Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						return updateErr
					},
				})

				// when
				_, res, err := reconcile(&aliceUserSignup)

				// then
				Expect(err).To(MatchError(updateErr))
				Expect(res).To(BeZero())
			})
		})

		When("an InternalWorkspace does not exists yet", func() {
			It("creates an InternalWorkspace for the user's home space", func() {
				// when
				r, res, err := reconcile(&aliceUserSignup)

				// then
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(BeZero())

				iw := workspacesv1alpha1.InternalWorkspace{}
				Expect(r.Get(ctx, client.ObjectKeyFromObject(&aliceInternalWorkspace), &iw)).To(Succeed())
				Expect(iw.Spec.Owner.JwtInfo.Sub).To(Equal(aliceUserSignup.Spec.IdentityClaims.Sub))
				Expect(iw.Spec.Owner.JwtInfo.Email).To(Equal(aliceUserSignup.Spec.IdentityClaims.Email))
				Expect(iw.Spec.Owner.JwtInfo.UserId).To(Equal(aliceUserSignup.Spec.IdentityClaims.UserID))
				Expect(iw.Spec.DisplayName).To(Equal("default"))
				Expect(iw.Spec.Visibility).To(Equal(workspacesv1alpha1.InternalWorkspaceVisibilityPrivate))
			})

			It("forwards the error if creation was not successful", func() {
				// given
				createErr := fmt.Errorf("test create error")
				clientBuilder = clientBuilder.WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						return createErr
					},
				})

				// when
				_, res, err := reconcile(&aliceUserSignup)

				// then
				Expect(res).To(BeZero())
				Expect(err).To(MatchError(createErr))
			})
		})
	})
})
