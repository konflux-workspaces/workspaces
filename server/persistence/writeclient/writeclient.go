package writeclient

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konflux-workspaces/workspaces/server/persistence/iwclient"

	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	restworkspacesv1alpha1 "github.com/konflux-workspaces/workspaces/server/api/v1alpha1"
)

// BuildClientFunc defines a function that builds a controller-runtime client
// that impersonates the given user
type BuildClientFunc func(user string) (client.Client, error)

// WriteClient implements Write primitives on Workspaces.
// Creates or updates InternalWorkspaces starting from a request on Workspaces.
type WriteClient struct {
	buildClient         BuildClientFunc
	workspacesNamespace string
	workspacesReader    *iwclient.Client
}

// New creates a new WriteClient
func New(buildClient BuildClientFunc, workspacesNamespace string, workspacesReader *iwclient.Client) *WriteClient {
	return &WriteClient{
		buildClient:         buildClient,
		workspacesNamespace: workspacesNamespace,
		workspacesReader:    workspacesReader,
	}
}

// BuildBuildClientFuncForConfig provides a configured BuildClientFunc for building a controller-runtime client
// for a given cluster and impersonating an user
func BuildBuildClientFuncForConfig(config *rest.Config) BuildClientFunc {
	newConfig := rest.CopyConfig(config)

	return func(user string) (client.Client, error) {
		newConfig.Impersonate.UserName = user

		s := runtime.NewScheme()
		if err := restworkspacesv1alpha1.AddToScheme(s); err != nil {
			return nil, err
		}
		if err := workspacesv1alpha1.AddToScheme(s); err != nil {
			return nil, err
		}

		return client.New(newConfig, client.Options{Scheme: s})
	}
}
