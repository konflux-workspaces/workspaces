package cache

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

type Cache struct{}

func NewCache(cli *client.Client) *Cache {
	return &Cache{}
}

func (c *Cache) ListUserWorkspaces(ctx context.Context, user string, objs *workspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error {
	panic("not implemented") // TODO: Implement
}
