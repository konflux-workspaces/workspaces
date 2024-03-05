package kube

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
)

var (
	_ workspace.WorkspaceUpdater = &Client{}
)

type Client struct {
	workspacesNamespace string
	config              *rest.Config
}

func New(cfg *rest.Config, workspacesNamespace string) *Client {
	return &Client{
		workspacesNamespace: workspacesNamespace,
		config:              rest.CopyConfig(cfg),
	}
}

func (c *Client) UpdateUserWorkspace(ctx context.Context, user string, workspace *workspacesv1alpha1.Workspace, opts ...client.UpdateOption) error {
	cli, err := c.buildImpersonatingClient(user)
	if err != nil {
		return err
	}

	ow := workspace.GetNamespace()
	defer func() {
		workspace.SetNamespace(ow)
	}()

	workspace.SetNamespace(c.workspacesNamespace)
	log.Printf("cli.Update %v ", workspace)
	return cli.Update(ctx, workspace, opts...)
}

func (c *Client) buildImpersonatingClient(user string) (client.Client, error) {
	config := rest.CopyConfig(c.config)
	config.Impersonate.UserName = user

	s := runtime.NewScheme()
	if err := workspacesv1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}

	return client.New(config, client.Options{
		Scheme: s,
	})
}
