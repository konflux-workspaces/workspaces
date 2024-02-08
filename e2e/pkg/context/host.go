package context

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

const (
	keyHostClient    string = "host-client"
	keyTestNamespace string = "test-namespace"
	keyWorkspace     string = "default-workspace"
	keyUser          string = "default-user"

	msgNotFound string = "key not found in context"
)

// Host Client
func InjectHostClient(ctx context.Context, cli client.Client) context.Context {
	return context.WithValue(ctx, keyHostClient, cli)
}

func RetrieveHostClient(ctx context.Context) client.Client {
	return get[client.Client](ctx, keyHostClient)
}

// Test Namespace
func InjectTestNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, keyTestNamespace, namespace)
}

func RetrieveTestNamespace(ctx context.Context) string {
	return get[string](ctx, keyTestNamespace)
}

// Default Workspace
func InjectWorkspace(ctx context.Context, w workspacesv1alpha1.Workspace) context.Context {
	return context.WithValue(ctx, keyWorkspace, w)
}

func RetrieveWorkspace(ctx context.Context) workspacesv1alpha1.Workspace {
	return get[workspacesv1alpha1.Workspace](ctx, keyWorkspace)
}

// Default User
func InjectUser(ctx context.Context, u toolchainv1alpha1.MasterUserRecord) context.Context {
	return context.WithValue(ctx, keyUser, u)
}

func RetrieveUser(ctx context.Context) toolchainv1alpha1.MasterUserRecord {
	return get[toolchainv1alpha1.MasterUserRecord](ctx, keyUser)
}

// auxiliary
func get[T any](ctx context.Context, key string) T {
	v, ok := ctx.Value(key).(T)
	if !ok {
		panic(fmt.Sprintf("%s: %s", msgNotFound, key))
	}

	return v
}
