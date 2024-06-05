package middleware_test

import (
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/cache"
)

//go:generate mockgen -source=interfaces_test.go -destination=mocks/cache.go -package=mocks -exclude_interfaces=FakeHTTPHandler
type FakeCRCache interface {
	cache.Cache
}

//go:generate mockgen -source=interfaces_test.go -destination=mocks/http_handler.go -package=mocks -exclude_interfaces=FakeCRCache
type FakeHTTPHandler interface {
	http.Handler
}
