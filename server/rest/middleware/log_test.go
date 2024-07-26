package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/konflux-workspaces/workspaces/server/log"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware"
	"github.com/konflux-workspaces/workspaces/server/rest/middleware/mocks"
)

var _ = Describe("LoggerInjectorMiddleware", Label("middleware"), Label("log"), func() {
	var ctrl *gomock.Controller
	var nextHandler *mocks.MockFakeHTTPHandler
	var logHandler *mocks.MockFakeSlogHandler

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		nextHandler = mocks.NewMockFakeHTTPHandler(ctrl)
		logHandler = mocks.NewMockFakeSlogHandler(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("injects logger into request", func() {
		// given
		request := httptest.NewRequest(http.MethodGet, "/whatever", nil)
		writer := httptest.NewRecorder()

		injectedLogger := slog.New(logHandler)
		loggerInjectorMiddleware := middleware.NewLoggerInjectorMiddleware(injectedLogger, nextHandler)

		// set expectation
		nextHandler.EXPECT().
			ServeHTTP(gomock.Any(), gomock.Any()).
			Times(1).
			DoAndReturn(func(_ http.ResponseWriter, r *http.Request) {
				// then
				logger := log.FromContext(r.Context())
				Expect(logger).To(Equal(injectedLogger))
			})

		// when
		loggerInjectorMiddleware.ServeHTTP(writer, request)

		// then
		// checked in nextHandler's expectations
	})

	It("injects logger with correlation-id into request", func() {
		// given
		request := httptest.NewRequest(http.MethodGet, "/whatever", nil)
		writer := httptest.NewRecorder()

		injectedLogger := slog.New(logHandler)
		loggerInjectorMiddleware := middleware.NewLoggerInjectorMiddlewareWithTracing(injectedLogger, nextHandler, func() string {
			return "my-correlation-id"
		})

		// set expectation
		nextHandler.EXPECT().
			ServeHTTP(gomock.Any(), gomock.Any()).
			Times(1).
			DoAndReturn(func(_ http.ResponseWriter, r *http.Request) {
				// then
				logger := log.FromContext(r.Context())
				Expect(logger).ToNot(BeNil())
				Expect(logger).ToNot(Equal(injectedLogger))
			})
		logHandler.EXPECT().WithAttrs(gomock.Any()).Times(1)

		// when
		loggerInjectorMiddleware.ServeHTTP(writer, request)

		// then
		// checked in nextHandler's expectations
	})

})

var _ = Describe("RequestLoggerMiddleware", Label("middleware"), Label("log"), func() {
	var ctrl *gomock.Controller
	var nextHandler *mocks.MockFakeHTTPHandler
	var logHandler *mocks.MockFakeSlogHandler

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		nextHandler = mocks.NewMockFakeHTTPHandler(ctrl)
		logHandler = mocks.NewMockFakeSlogHandler(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("logs at Info Level by default", func() {
		// given
		injectedLogger := slog.New(logHandler)
		ctx := log.IntoContext(context.TODO(), injectedLogger)
		request := httptest.
			NewRequest(http.MethodGet, "/whatever", nil).
			WithContext(ctx)
		writer := httptest.NewRecorder()

		m := middleware.NewRequestLoggerMiddleware(nextHandler)

		// set expectations
		nextHandler.EXPECT().
			ServeHTTP(writer, request).
			Times(1)
		logHandler.EXPECT().
			Handle(gomock.Any(), gomock.Any()).
			Times(1)
		logHandler.EXPECT().
			Enabled(gomock.Any(), gomock.Any()).
			Times(1).
			DoAndReturn(func(ctx context.Context, level slog.Level) bool {
				// then
				Expect(level).To(Equal(slog.LevelInfo))
				return true
			})

		// when
		m.ServeHTTP(writer, request)

		// then
		// checked in logHandler's expectations
	})

	It("logs at Info Level by default", func() {
		// given
		logLevel := slog.LevelDebug
		injectedLogger := slog.New(logHandler)
		ctx := log.IntoContext(context.TODO(), injectedLogger)
		request := httptest.
			NewRequest(http.MethodGet, "https://mydomain/myendpoint", nil).
			WithContext(ctx)
		writer := httptest.NewRecorder()
		m := middleware.NewRequestLoggerMiddlewareWithLogLevel(nextHandler, logLevel)

		expectedAttrs := map[string]string{
			middleware.LogKeyMethod: request.Method,
			middleware.LogKeyURL:    request.URL.String(),
		}

		// set expectations
		nextHandler.EXPECT().
			ServeHTTP(writer, request).
			Times(1)
		logHandler.EXPECT().
			Enabled(gomock.Any(), gomock.Any()).
			Times(1).
			Return(true)
		logHandler.EXPECT().
			Handle(gomock.Any(), gomock.Any()).
			Times(1).
			DoAndReturn(func(ctx context.Context, record slog.Record) error {
				// then
				Expect(record.Level).To(Equal(logLevel))
				Expect(record.Message).To(Equal("request"))
				Expect(record.NumAttrs()).To(Equal(2))

				// extract attrs
				attrs := map[string]string{}
				record.Attrs(func(a slog.Attr) bool {
					Expect(a.Value.String()).NotTo(BeEmpty())
					attrs[a.Key] = a.Value.String()
					return true
				})

				Expect(attrs).To(Equal(expectedAttrs))
				return nil
			})

		// when
		m.ServeHTTP(writer, request)

		// then
		// checked in logHandler's expectations
	})
})
