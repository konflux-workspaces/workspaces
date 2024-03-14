package middleware

import (
	"context"
	"net/http"
)

var (
	_ http.Handler = &HeaderInfoMiddleware{}
	_ http.Handler = &HeaderInfoMiddleware{}
)

// HeaderInfoMiddleware reads headers from request and add their value in request context
type HeaderInfoMiddleware struct {
	headers map[string]interface{}
	next    http.Handler
}

// HeaderInfoMiddleware reads headers from request and add their value in request context
func (m *HeaderInfoMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	for header, contextKey := range m.headers {
		value := r.Header.Get(header)
		ctx = context.WithValue(ctx, contextKey, value)
	}

	m.next.ServeHTTP(w, r.WithContext(ctx))
}

// NewHeaderInfoMiddleware builds a new HeaderInfoMiddleware
func NewHeaderInfoMiddleware(next http.Handler, headers map[string]interface{}) *HeaderInfoMiddleware {
	return &HeaderInfoMiddleware{next: next, headers: headers}
}
