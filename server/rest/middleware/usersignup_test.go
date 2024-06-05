package middleware_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"

	ccontext "github.com/konflux-workspaces/workspaces/server/core/context"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware/mocks"
)

const (
	endpointWhatever = "/whatever"

	methodGet = "GET"

	testUserSub = "test-user-sub"
)

var _ = Describe("Usersignup", func() {

	var (
		ctx context.Context

		// mocks
		c *mocks.MockFakeCRCache
		h *mocks.MockFakeHTTPHandler

		// http
		w *httptest.ResponseRecorder
		r *http.Request

		// middleware
		m *middleware.UserSignupMiddleware
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// mocks
		ctrl := gomock.NewController(GinkgoT())
		c = mocks.NewMockFakeCRCache(ctrl)
		h = mocks.NewMockFakeHTTPHandler(ctrl)

		// http
		w = httptest.NewRecorder()
		b := bytes.NewBuffer([]byte{})
		lr, err := http.NewRequestWithContext(ctx, methodGet, endpointWhatever, b)
		Expect(err).NotTo(HaveOccurred())
		r = lr

		// middleware
		m = middleware.NewUserSignupMiddleware(h, c, true)
	})

	When("sub HTTP header is not present", func() {
		It("invokes next handler", func() {
			// set expectations
			c.EXPECT().List(gomock.Any(), gomock.Any()).Times(0)
			h.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(
					func(_ http.ResponseWriter, r *http.Request) {
						v := r.Context().Value(ccontext.UserSignupComplaintNameKey)
						Expect(v).To(BeNil(), "expecting context key '%s' not to be set", ccontext.UserSignupComplaintNameKey)

						w.WriteHeader(999)
						_, err := w.Write(nil)
						Expect(err).NotTo(HaveOccurred(), "error writing empty payload")
					},
				)

			// when
			m.ServeHTTP(w, r)

			// then
			Expect(w.Code).To(Equal(999))
			Expect(w.Body.String()).To(BeZero())
		})
	})

	When("sub HTTP header is present", func() {
		BeforeEach(func() {
			ctx = context.WithValue(context.TODO(), ccontext.UserSubKey, testUserSub)
		})

		It("requires an usersignup", func() {
			// set expectations
			c.EXPECT().List(gomock.Any(), gomock.Any()).Times(1)
			h.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(0)

			// when
			m.ServeHTTP(w, r.WithContext(ctx))

			// then
			Expect(w.Code).To(Equal(http.StatusForbidden))
			Expect(w.Body.String()).To(Equal("user needs to sign in"))
		})

		It("requires the usersignup fetch to complete successfully", func() {
			// set expectations
			c.EXPECT().
				List(gomock.Any(), gomock.Any()).
				Times(1).
				Return(fmt.Errorf("error"))
			h.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(0)

			// when
			m.ServeHTTP(w, r.WithContext(ctx))

			// then
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("requires the usersignup to be approved", func() {
			// set expectations
			c.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(
					func(_ context.Context, list *toolchainv1alpha1.UserSignupList, _ ...client.ListOption) error {
						*list = toolchainv1alpha1.UserSignupList{
							Items: []toolchainv1alpha1.UserSignup{
								{
									ObjectMeta: metav1.ObjectMeta{
										Name:      "test-user",
										Namespace: "toolchain-host-operator",
									},
									Spec: toolchainv1alpha1.UserSignupSpec{
										IdentityClaims: toolchainv1alpha1.IdentityClaimsEmbedded{
											PropagatedClaims: toolchainv1alpha1.PropagatedClaims{
												Sub: testUserSub,
											},
										},
									},
									Status: toolchainv1alpha1.UserSignupStatus{
										// CompliantUsername: "test-user",
									},
								},
							},
						}
						return nil
					},
				)
			h.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(0)

			// when
			m.ServeHTTP(w, r.WithContext(ctx))

			// then
			Expect(w.Code).To(Equal(http.StatusForbidden))
			Expect(w.Body.String()).To(Equal("user is waiting for approval"))
		})

		It("succeeds when ComplaintName is set", func() {
			// set expectations
			c.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(
					func(_ context.Context, list *toolchainv1alpha1.UserSignupList, _ ...client.ListOption) error {
						*list = toolchainv1alpha1.UserSignupList{
							Items: []toolchainv1alpha1.UserSignup{
								{
									ObjectMeta: metav1.ObjectMeta{
										Name:      "test-user",
										Namespace: "toolchain-host-operator",
									},
									Spec: toolchainv1alpha1.UserSignupSpec{
										IdentityClaims: toolchainv1alpha1.IdentityClaimsEmbedded{
											PropagatedClaims: toolchainv1alpha1.PropagatedClaims{
												Sub: testUserSub,
											},
										},
									},
									Status: toolchainv1alpha1.UserSignupStatus{
										CompliantUsername: "test-user",
									},
								},
							},
						}
						return nil
					},
				)
			h.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(
					func(_ http.ResponseWriter, r *http.Request) {
						u, ok := r.Context().Value(ccontext.UserSignupComplaintNameKey).(string)
						Expect(ok).To(BeTrue(), "expecting UserSignup.ComplaintName to be forwarded in the context")
						Expect(u).To(Equal("test-user"))

						w.WriteHeader(999)
						_, err := w.Write(nil)
						Expect(err).NotTo(HaveOccurred(), "error writing empty payload")
					},
				)

			// when
			m.ServeHTTP(w, r.WithContext(ctx))

			// then
			Expect(w.Code).To(Equal(999))
			Expect(w.Body.String()).To(BeZero())
		})
	})
})
