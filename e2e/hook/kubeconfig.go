package hook

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func injectUnauthKubeconfig(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	apiConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return nil, fmt.Errorf("error building config: %v", err)
	}

	cfg, err := clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error building config: %v", err)
	}

	c := &rest.Config{
		ContentConfig: cfg.ContentConfig,
		Host:          cfg.Host,
		APIPath:       cfg.APIPath,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile:     cfg.CAFile,
			CAData:     cfg.CAData,
			ServerName: cfg.ServerName,
			Insecure:   cfg.Insecure,
		},
		Timeout: cfg.Timeout,
	}

	return tcontext.InjectUnauthKubeconfig(ctx, c), nil
}
