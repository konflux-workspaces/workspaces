package readclient

import (
	"context"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/konflux-workspaces/workspaces/server/persistence/clientinterface"
	icache "github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
)

// ReadClient implements the WorkspaceLister and WorkspaceReader interfaces
// using a client.Reader as backend
type ReadClient struct {
	internalClient clientinterface.InternalWorkspacesReadClient
	mapper         clientinterface.InternalWorkspacesMapper
}

// NewDefaultWithCache creates a controller-runtime cache and use it as KubeReadClient's backend.
// It also uses the default InternalWorkspaces/Workspaces mapper.
func NewDefaultWithCache(ctx context.Context, cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (*ReadClient, cache.Cache, error) {
	c, err := icache.NewCache(ctx, cfg, workspacesNamespace, kubesawNamespace)
	if err != nil {
		return nil, nil, err
	}

	internalClient := iwclient.New(c, workspacesNamespace, kubesawNamespace)
	return NewDefaultWithInternalClient(internalClient), c, nil
}

// NewDefaultWithInternalClient creates a new KubeReadClient with the provided backend and default InternalWorkspaces/Workspaces mapper
func NewDefaultWithInternalClient(internalClient clientinterface.InternalWorkspacesReadClient) *ReadClient {
	return New(internalClient, mapper.Default)
}

// New creates a new KubeReadClient with the provided backend and a custom InternalWorkspaces/Workspaces mapper
func New(internalClient clientinterface.InternalWorkspacesReadClient, mapper clientinterface.InternalWorkspacesMapper) *ReadClient {
	return &ReadClient{
		internalClient: internalClient,
		mapper:         mapper,
	}
}
