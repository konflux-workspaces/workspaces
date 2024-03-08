package middleware

import (
	"log/slog"
	"net/http"

	"github.com/konflux-workspaces/workspaces/server/log"
)

var (
	_ http.Handler = &RequestLoggerMiddleware{}
	_ http.Handler = &LoggerInjectorMiddleware{}
)

// LoggerInjectorMiddleware injects the logger in the request then calls the next handler
type LoggerInjectorMiddleware struct {
	logger *slog.Logger
	next   http.Handler
}

// NewLoggerInjectorMiddleware builds a new LoggerInjectorMiddleware
func NewLoggerInjectorMiddleware(logger *slog.Logger, next http.Handler) *LoggerInjectorMiddleware {
	return &LoggerInjectorMiddleware{
		logger: logger,
		next:   next,
	}
}

// ServeHTTP injects the logger in the request then calls the next handler
func (m *LoggerInjectorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := log.IntoContext(r.Context(), m.logger)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

// RequestLoggerMiddleware logs the request method and path then calls the next handler
type RequestLoggerMiddleware struct {
	next http.Handler
}

// NewRequestLoggerMiddleware builds a new LogRequestMiddleware
func NewRequestLoggerMiddleware(next http.Handler) *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{next: next}
}

// ServeHTTP logs the request method and path then calls the next handler
func (m *RequestLoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.FromContext(r.Context()).Info("request", "method", r.Method, "url", r.URL.String())

	m.next.ServeHTTP(w, r)
}
