package readclient_test

import (
	"github.com/konflux-workspaces/workspaces/server/persistence/clientinterface"
)

//go:generate mockgen -source=interface_test.go -destination=mocks/readclient.go -package=mocks -exclude_interfaces=FakeIWMapper
type FakeIWReadClient interface {
	clientinterface.InternalWorkspacesReadClient
}

//go:generate mockgen -source=interface_test.go -destination=mocks/mapper.go -package=mocks -exclude_interfaces=FakeIWReadClient
type FakeIWMapper interface {
	clientinterface.InternalWorkspacesMapper
}
