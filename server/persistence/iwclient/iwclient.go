package iwclient

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
)

type SpaceKey struct {
	Owner string
	Name  string
}

type Client struct {
	backend client.Reader

	kubesawNamespace    string
	workspacesNamespace string
}

func New(backend client.Reader, workspacesNamespace, kubesawNamespace string) *Client {
	return &Client{
		backend:             backend,
		kubesawNamespace:    kubesawNamespace,
		workspacesNamespace: workspacesNamespace,
	}
}

func (c *Client) existsSpaceBindingForUserAndSpace(ctx context.Context, user, space string) (bool, error) {
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

func (c *Client) workspaceNamespacedName(space string) client.ObjectKey {
	return types.NamespacedName{Namespace: c.workspacesNamespace, Name: space}
}
