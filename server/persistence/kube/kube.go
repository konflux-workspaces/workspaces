package kube

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/server/core/workspace"
	"github.com/konflux-workspaces/workspaces/server/log"
)

var (
	_ workspace.WorkspaceUpdater = &Client{}
)

type BuildClientFunc func(string) (client.Client, error)

type Client struct {
	buildClient         BuildClientFunc
	workspacesNamespace string
}

func New(buildClient BuildClientFunc, workspacesNamespace string) *Client {
	return &Client{
		buildClient:         buildClient,
		workspacesNamespace: workspacesNamespace,
	}
}

func (c *Client) CreateUserWorkspace(ctx context.Context, user string, workspace *workspacesv1alpha1.Workspace, opts ...client.CreateOption) error {
	cli, err := c.buildClient(user)
	if err != nil {
		return err
	}

	ow := workspace.GetNamespace()
	defer func() {
		workspace.SetNamespace(ow)
	}()

	workspace.SetNamespace(c.workspacesNamespace)
	log.FromContext(ctx).Debug("creating user workspace", "workspace", workspace, "user", user)
	return cli.Create(ctx, workspace, opts...)
}

func (c *Client) UpdateUserWorkspace(ctx context.Context, user string, workspace *workspacesv1alpha1.Workspace, opts ...client.UpdateOption) error {
	cli, err := c.buildClient(user)
	if err != nil {
		return err
	}

	ow := workspace.GetNamespace()
	defer func() {
		workspace.SetNamespace(ow)
	}()

	workspace.SetNamespace(c.workspacesNamespace)
	log.FromContext(ctx).Debug("updating user workspace", "workspace", workspace, "user", user)
	return cli.Update(ctx, workspace, opts...)
}

func BuildClient(config *rest.Config) BuildClientFunc {
	newConfig := rest.CopyConfig(config)
	return func(user string) (client.Client, error) {
		newConfig.Impersonate.UserName = user

		s := runtime.NewScheme()
		if err := workspacesv1alpha1.AddToScheme(s); err != nil {
			return nil, err
		}

		return client.New(newConfig, client.Options{
			Scheme: s,
		})
	}
}
