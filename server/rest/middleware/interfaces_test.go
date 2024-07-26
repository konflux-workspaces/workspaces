package middleware_test

import (
	"log/slog"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/cache"
)

//go:generate mockgen -source=interfaces_test.go -destination=mocks/cache.go -package=mocks -exclude_interfaces=FakeHTTPHandler,FakeSlogHandler
type FakeCRCache interface {
	cache.Cache
}

//go:generate mockgen -source=interfaces_test.go -destination=mocks/http_handler.go -package=mocks -exclude_interfaces=FakeCRCache,FakeSlogHandler
type FakeHTTPHandler interface {
	http.Handler
}

//go:generate mockgen -source=interfaces_test.go -destination=mocks/slog_handler.go -package=mocks -exclude_interfaces=FakeCRCache,FakeHTTPHandler
type FakeSlogHandler interface {
	slog.Handler
}
