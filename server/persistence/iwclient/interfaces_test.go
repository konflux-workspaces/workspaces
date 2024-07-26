package iwclient_test

import "sigs.k8s.io/controller-runtime/pkg/client"

//go:generate mockgen -source=interfaces_test.go -destination=mocks/client_reader.go -package=mocks
type FakeCRReader interface {
	client.Reader
}
