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

const (
	LogKeyTrace  string = "trace"
	LogKeyMethod string = "method"
	LogKeyURL    string = "url"
)

type GenerateCorrelationIdFunc func() string

// LoggerInjectorMiddleware injects the logger in the request then calls the next handler
type LoggerInjectorMiddleware struct {
	logger                    *slog.Logger
	next                      http.Handler
	generateCorrelationIdFunc GenerateCorrelationIdFunc
}

// NewLoggerInjectorMiddleware builds a new LoggerInjectorMiddleware
func NewLoggerInjectorMiddleware(logger *slog.Logger, next http.Handler) *LoggerInjectorMiddleware {
	return &LoggerInjectorMiddleware{
		logger: logger,
		next:   next,
	}
}

// NewLoggerInjectorMiddleware builds a new LoggerInjectorMiddleware
func NewLoggerInjectorMiddlewareWithTracing(logger *slog.Logger, next http.Handler, generateCorrelationIdFunc GenerateCorrelationIdFunc) *LoggerInjectorMiddleware {
	return &LoggerInjectorMiddleware{
		logger:                    logger,
		next:                      next,
		generateCorrelationIdFunc: generateCorrelationIdFunc,
	}
}

// ServeHTTP injects the logger in the request then calls the next handler
func (m *LoggerInjectorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := func() *slog.Logger {
		if m.generateCorrelationIdFunc != nil {
			return m.logger.With(LogKeyTrace, m.generateCorrelationIdFunc())
		}
		return m.logger
	}()

	ctx := log.IntoContext(r.Context(), l)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

// RequestLoggerMiddleware logs the request method and path then calls the next handler
type RequestLoggerMiddleware struct {
	logLevel slog.Level
	next     http.Handler
}

// NewRequestLoggerMiddleware builds a new LogRequestMiddleware. LogLevel is Info.
func NewRequestLoggerMiddleware(next http.Handler) *RequestLoggerMiddleware {
	return NewRequestLoggerMiddlewareWithLogLevel(next, slog.LevelInfo)
}

// NewRequestLoggerMiddlewareWithLogLevel builds a new LogRequestMiddleware configuring the log level
func NewRequestLoggerMiddlewareWithLogLevel(next http.Handler, logLevel slog.Level) *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{
		next:     next,
		logLevel: logLevel,
	}
}

// ServeHTTP logs the request method and path then calls the next handler
func (m *RequestLoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.FromContext(r.Context()).LogAttrs(
		r.Context(),
		m.logLevel,
		"request",
		slog.String(LogKeyMethod, r.Method),
		slog.String(LogKeyURL, r.URL.String()),
	)

	m.next.ServeHTTP(w, r)
}
