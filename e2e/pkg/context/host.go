package context

import (
	"context"
	"fmt"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	"k8s.io/client-go/rest"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

type ContextKey string

const (
	keyUnauthenticatedKubeconfig ContextKey = "unauth-kubeconfig"
	keyHostClient                ContextKey = "host-client"
	keyTestNamespace             ContextKey = "test-namespace"
	keyScenarioId                ContextKey = "scenario-id"
	keyKubespaceNamespace        ContextKey = "kubespace-namespace"
	keyWorkspacesNamespace       ContextKey = "workspaces-namespace"
	keyInternalWorkspace         ContextKey = "default-internal-workspace"
	keyUserWorkspace             ContextKey = "user-workspace"
	keyUser                      ContextKey = "default-user"
	keyUserWorkspaces            ContextKey = "workspaces"

	msgNotFound string = "key not found in context"
)

// Kubeconfig
func InjectUnauthKubeconfig(ctx context.Context, cli *rest.Config) context.Context {
	return context.WithValue(ctx, keyUnauthenticatedKubeconfig, cli)
}

func RetrieveUnauthKubeconfig(ctx context.Context) *rest.Config {
	return get[*rest.Config](ctx, keyUnauthenticatedKubeconfig)
}

// Host Client
func InjectHostClient(ctx context.Context, cli cli.Cli) context.Context {
	return context.WithValue(ctx, keyHostClient, cli)
}

func RetrieveHostClient(ctx context.Context) cli.Cli {
	return get[cli.Cli](ctx, keyHostClient)
}

// Kubespace Namespace
func InjectKubespaceNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, keyKubespaceNamespace, namespace)
}

func RetrieveKubespaceNamespace(ctx context.Context) string {
	return get[string](ctx, keyKubespaceNamespace)
}

// Workspaces Namespace
func InjectWorkspacesNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, keyWorkspacesNamespace, namespace)
}

func RetrieveWorkspacesNamespace(ctx context.Context) string {
	return get[string](ctx, keyWorkspacesNamespace)
}

// Test Namespace
func InjectTestNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, keyTestNamespace, namespace)
}

func RetrieveTestNamespace(ctx context.Context) string {
	return get[string](ctx, keyTestNamespace)
}

// Default Workspace
func InjectInternalWorkspace(ctx context.Context, w workspacesv1alpha1.InternalWorkspace) context.Context {
	return context.WithValue(ctx, keyInternalWorkspace, w)
}

func RetrieveInternalWorkspace(ctx context.Context) workspacesv1alpha1.InternalWorkspace {
	return get[workspacesv1alpha1.InternalWorkspace](ctx, keyInternalWorkspace)
}

func LookupInternalWorkspace(ctx context.Context) (workspacesv1alpha1.InternalWorkspace, bool) {
	return lookup[workspacesv1alpha1.InternalWorkspace](ctx, keyInternalWorkspace)
}

// Default User
func InjectUser(ctx context.Context, u toolchainv1alpha1.UserSignup) context.Context {
	return context.WithValue(ctx, keyUser, u)
}

func RetrieveUser(ctx context.Context) toolchainv1alpha1.UserSignup {
	return get[toolchainv1alpha1.UserSignup](ctx, keyUser)
}

// Workspaces
func InjectUserWorkspace(ctx context.Context, w restworkspacesv1alpha1.Workspace) context.Context {
	return context.WithValue(ctx, keyUserWorkspace, w)
}

func RetrieveUserWorkspace(ctx context.Context) restworkspacesv1alpha1.Workspace {
	return get[restworkspacesv1alpha1.Workspace](ctx, keyUserWorkspace)
}

func LookupUserWorkspace(ctx context.Context) (restworkspacesv1alpha1.Workspace, bool) {
	return lookup[restworkspacesv1alpha1.Workspace](ctx, keyUserWorkspace)
}

func InjectUserWorkspaces(ctx context.Context, ww workspacesv1alpha1.InternalWorkspaceList) context.Context {
	return context.WithValue(ctx, keyUserWorkspaces, ww)
}

func RetrieveUserWorkspaces(ctx context.Context) workspacesv1alpha1.InternalWorkspaceList {
	return get[workspacesv1alpha1.InternalWorkspaceList](ctx, keyUserWorkspaces)
}

// Scenario Id
func InjectScenarioId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyTestNamespace, id)
}

func RetrieveScenarioId(ctx context.Context) string {
	return get[string](ctx, keyScenarioId)
}

// auxiliary
func get[T any](ctx context.Context, key ContextKey) T {
	v, ok := lookup[T](ctx, key)
	if !ok {
		panic(fmt.Sprintf("%s: %s", msgNotFound, key))
	}

	return v
}

func lookup[T any](ctx context.Context, key ContextKey) (T, bool) {
	v, ok := ctx.Value(key).(T)
	return v, ok
}
