package rest

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/auth"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesiov1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

// NewDefaultClientConfig retrieves the client configuration from the process environment
// using the "k8s.io/client-go/tools/clientcmd" utilities
func NewDefaultClientConfig() (*rest.Config, error) {
	apiConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return nil, fmt.Errorf("error building config: %v", err)
	}

	cfg, err := clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error building config: %v", err)
	}

	mutateConfig(cfg)
	return cfg, nil
}

// BuildDefaultHostClient builds the default host client.
// It uses NewDefaultClientConfig for retrieving the client configuration.
func BuildDefaultHostClient() (client.Client, error) {
	cfg, err := NewDefaultClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error building config: %v", err)
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(workspacesiov1alpha1.AddToScheme(scheme))
	utilruntime.Must(toolchainv1alpha1.AddToScheme(scheme))

	return client.New(cfg, client.Options{Scheme: scheme})
}

// BuildWorkspacesClient builds a client that targets the Workspaces REST API server.
// It also builds a valid JWT token for authenticating the requests.
func BuildWorkspacesClient(ctx context.Context) (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(restworkspacesv1alpha1.AddToScheme(scheme))
	utilruntime.Must(toolchainv1alpha1.AddToScheme(scheme))
	utilruntime.Must(workspacesiov1alpha1.AddToScheme(scheme))

	u := tcontext.RetrieveUser(ctx)
	k := tcontext.RetrieveUnauthKubeconfig(ctx)

	t, err := auth.BuildJwtForUser(ctx, u.Status.CompliantUsername)
	if err != nil {
		return nil, err
	}
	k.BearerToken = t
	k.Host = os.Getenv("PROXY_URL")

	m, err := BuildDefaultRESTMapper()
	if err != nil {
		return nil, err
	}

	c, err := client.New(k, client.Options{Scheme: scheme, Mapper: m})
	if err != nil {
		return nil, fmt.Errorf("error building client for host %s and user %s: %w", k.Host, u.Status.CompliantUsername, err)
	}

	return c, nil
}

// BuildDefaultRESTMapper builds a RESTMapper from the default client configuration.
func BuildDefaultRESTMapper() (meta.RESTMapper, error) {
	cfg, err := NewDefaultClientConfig()
	if err != nil {
		return nil, err
	}

	return BuildRESTMapper(cfg)
}

// BuildRESTMapper builds a RESTMapper from a given configuration.
func BuildRESTMapper(cfg *rest.Config) (meta.RESTMapper, error) {
	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}

	m, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return nil, err
	}
	return m, nil
}
