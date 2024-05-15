package readclient

import (
	"context"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
	icache "github.com/konflux-workspaces/workspaces/server/persistence/internal/cache"
	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
)

var (
	_ InternalWorkspacesReadClient = &iwclient.Client{}
	_ InternalWorkspacesMapper     = mapper.Default
)

// InternalWorkspacesReadClient is the definition for a InternalWorkspaces Read Client
type InternalWorkspacesReadClient interface {
	GetAsUser(context.Context, string, iwclient.SpaceKey, *workspacesv1alpha1.InternalWorkspace, ...client.GetOption) error
	ListAsUser(context.Context, string, *workspacesv1alpha1.InternalWorkspaceList) error
}

// InternalWorkspacesMapper is the definition for a InternalWorkspaces/Workspaces Mapper
type InternalWorkspacesMapper interface {
	InternalWorkspaceListToWorkspaceList(*workspacesv1alpha1.InternalWorkspaceList) (*restworkspacesv1alpha1.WorkspaceList, error)
	InternalWorkspaceToWorkspace(*workspacesv1alpha1.InternalWorkspace) (*restworkspacesv1alpha1.Workspace, error)
	WorkspaceToInternalWorkspace(*restworkspacesv1alpha1.Workspace) (*workspacesv1alpha1.InternalWorkspace, error)
}

// ReadClient implements the WorkspaceLister and WorkspaceReader interfaces
// using a client.Reader as backend
type ReadClient struct {
	internalClient InternalWorkspacesReadClient
	mapper         InternalWorkspacesMapper
}

// NewDefaultWithCache creates a controller-runtime cache and use it as KubeReadClient's backend.
func NewDefaultWithCache(ctx context.Context, cfg *rest.Config, workspacesNamespace, kubesawNamespace string) (*ReadClient, cache.Cache, error) {
	c, err := icache.NewCache(ctx, cfg, workspacesNamespace, kubesawNamespace)
	if err != nil {
		return nil, nil, err
	}

	internalClient := iwclient.New(c, workspacesNamespace, kubesawNamespace)
	return NewDefaultWithInternalClient(internalClient), c, nil
}

// NewDefaultWithInternalClient creates a new KubeReadClient with the provided backend
func NewDefaultWithInternalClient(internalClient InternalWorkspacesReadClient) *ReadClient {
	return New(internalClient, mapper.Default)
}

func New(internalClient InternalWorkspacesReadClient, mapper InternalWorkspacesMapper) *ReadClient {
	return &ReadClient{
		internalClient: internalClient,
		mapper:         mapper,
	}
}
