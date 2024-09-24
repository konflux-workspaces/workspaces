package iwclient

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/persistence/clientinterface"
	"github.com/konflux-workspaces/workspaces/server/persistence/mapper"
)

var (
	_ clientinterface.InternalWorkspacesReadClient = &Client{}
	_ clientinterface.InternalWorkspacesMapper     = mapper.Default
)

type Client struct {
	backend client.Reader

	kubesawNamespace    string
	workspacesNamespace string
}

// New creates a client that uses the provided backend as source
func New(backend client.Reader, workspacesNamespace, kubesawNamespace string) *Client {
	return &Client{
		backend:             backend,
		kubesawNamespace:    kubesawNamespace,
		workspacesNamespace: workspacesNamespace,
	}
}

func (c *Client) UserHasDirectAccess(ctx context.Context, user, space string) (bool, error) {
	ml := client.MatchingLabels{
		toolchainv1alpha1.SpaceBindingMasterUserRecordLabelKey: user,
		toolchainv1alpha1.SpaceBindingSpaceLabelKey:            space,
	}
	sbb := toolchainv1alpha1.SpaceBindingList{}
	if err := c.backend.List(ctx, &sbb, ml); err != nil {
		return false, err
	}

	return len(sbb.Items) > 0, nil
}
