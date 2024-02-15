package cli

import (
	"context"
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cli struct {
	client.Client

	ScenarioId string
}

func New(c client.Client, id string) Cli {
	return Cli{ScenarioId: id, Client: c}
}

func (c *Cli) HasScenarioPrefix(name string) bool {
	return strings.HasPrefix(name, fmt.Sprintf("test-%s-", c.ScenarioId))
}
func (c *Cli) EnsurePrefix(name string) string {
	if c.HasScenarioPrefix(name) {
		return name
	}

	return fmt.Sprintf("test-%s-%s", c.ScenarioId, name)
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (c *Cli) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	key.Name = c.EnsurePrefix(key.Name)
	obj.SetName(key.Name)
	return c.Client.Get(ctx, key, obj, opts...)
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
// Deprecated: Use c.Client.List directly. This function is needed for using this struct in controllerutil.CreateOrUpdate.
func (c *Cli) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	panic("not implemented: use c.Client.List and do it wisely as it is not enforcing test naming convention")
}

// Create saves the object obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (c *Cli) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	obj.SetName(c.EnsurePrefix(obj.GetName()))
	return c.Client.Create(ctx, obj, opts...)
}

// Delete deletes the given obj from Kubernetes cluster.
func (c *Cli) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	obj.SetName(c.EnsurePrefix(obj.GetName()))
	return c.Client.Delete(ctx, obj, opts...)
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (c *Cli) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	obj.SetName(c.EnsurePrefix(obj.GetName()))
	return c.Client.Update(ctx, obj, opts...)
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (c *Cli) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	obj.SetName(c.EnsurePrefix(obj.GetName()))
	return c.Client.Patch(ctx, obj, patch, opts...)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
// Deprecated: Use c.Client.DeleteAllOf directly. This function is needed for using this struct in controllerutil.CreateOrUpdate.
func (c *Cli) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	panic("not implemented: use c.Client.DeleteAllOf and do it wisely as it is not enforcing test naming convention")
}
