package middleware_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/konflux-workspaces/workspaces/server/rest/middleware"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware/mocks"
)

var _ = Describe("HeaderInfo", func() {

	var (
		ctx context.Context

		// mocks
		h *mocks.MockFakeHTTPHandler

		// http
		w *httptest.ResponseRecorder
		r *http.Request
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// mocks
		ctrl := gomock.NewController(GinkgoT())
		h = mocks.NewMockFakeHTTPHandler(ctrl)

		// http
		w = httptest.NewRecorder()
		b := bytes.NewBuffer([]byte{})
		lr, err := http.NewRequestWithContext(ctx, methodGet, endpointWhatever, b)
		Expect(err).NotTo(HaveOccurred())
		r = lr
	})

	When("headers are nil", func() {
		It("invokes next handler", func() {
			// set expectations
			var hh map[string]interface{} = nil
			h.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(
					func(_ http.ResponseWriter, r *http.Request) {
						w.WriteHeader(999)
						_, err := w.Write(nil)
						Expect(err).NotTo(HaveOccurred(), "error writing empty payload")
					},
				)

			// when
			// middleware
			middleware.NewHeaderInfoMiddleware(h, hh).ServeHTTP(w, r)

			// then
			Expect(w.Code).To(Equal(999))
			Expect(w.Body.String()).To(BeZero())
		})
	})

	When("headers are empty", func() {
		It("invokes next handler", func() {
			// set expectations
			hh := map[string]interface{}{}
			h.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(
					func(_ http.ResponseWriter, r *http.Request) {
						w.WriteHeader(999)
						_, err := w.Write(nil)
						Expect(err).NotTo(HaveOccurred(), "error writing empty payload")
					},
				)

			// when
			// middleware
			middleware.NewHeaderInfoMiddleware(h, hh).ServeHTTP(w, r)

			// then
			Expect(w.Code).To(Equal(999))
			Expect(w.Body.String()).To(BeZero())
		})
	})

	When("headers are set in middleware", func() {
		It("injects request's header values in context", func() {
			myHeader, myHeaderValue, myContextKey := "myHeader", "myHeaderValue", "myContextKey"
			hh := map[string]interface{}{myHeader: myContextKey}
			r.Header.Add(myHeader, myHeaderValue)

			// set expectations
			h.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Times(1).
				DoAndReturn(
					func(_ http.ResponseWriter, r *http.Request) {
						v := r.Context().Value(myContextKey)
						Expect(v).To(Equal(myHeaderValue), "expecting context key '%s' to be set: %+v", myContextKey, r.Context())

						w.WriteHeader(999)
						_, err := w.Write(nil)
						Expect(err).NotTo(HaveOccurred(), "error writing empty payload")
					},
				)

				// when
				// middleware
			middleware.NewHeaderInfoMiddleware(h, hh).ServeHTTP(w, r)

			// then
			Expect(w.Code).To(Equal(999))
			Expect(w.Body.String()).To(BeZero())
		})

		When("request has no headers", func() {
			It("injects no values in context", func() {
				myHeader, myContextKey := "myHeader", "myContextKey"
				hh := map[string]interface{}{myHeader: myContextKey}

				// set expectations
				h.EXPECT().
					ServeHTTP(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(
						func(_ http.ResponseWriter, r *http.Request) {
							v := r.Context().Value(myContextKey)
							Expect(v).To(BeNil(), "expecting context key '%s' to be set: %+v", myContextKey, r.Context())

							w.WriteHeader(999)
							_, err := w.Write(nil)
							Expect(err).NotTo(HaveOccurred(), "error writing empty payload")
						},
					)

					// when
					// middleware
				middleware.NewHeaderInfoMiddleware(h, hh).ServeHTTP(w, r)

				// then
				Expect(w.Code).To(Equal(999))
				Expect(w.Body.String()).To(BeZero())
			})
		})
	})
})
