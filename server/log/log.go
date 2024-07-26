package log

import (
	"context"
	"log/slog"
)

type contextKey string

const contextKeyLogger contextKey = "logger"

var DefaultLogger *slog.Logger = slog.New(&NoOpHandler{})

// FromContext extracts logger from the given context.
// If the given context does not contain any, then a NoOpHandler is returned
func FromContext(ctx context.Context) *slog.Logger {
	// use the logger from context, if it has any
	if l, ok := ctx.Value(contextKeyLogger).(*slog.Logger); ok && l != nil {
		return l
	}

	// to prevent panics return a NoOpHandler
	return DefaultLogger
}

// IntoContext build a child context with the logger
func IntoContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}
