package workspace_test

import "net/http"

//go:generate mockgen -source=interface_test.go -destination=mocks/writer.go -package=mocks FakeResponseWriter

type FakeResponseWriter interface {
	http.ResponseWriter
}
