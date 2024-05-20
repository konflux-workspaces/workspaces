package hook

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
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

	ProcessConfig(cfg)
	return tcontext.InjectUnauthKubeconfig(ctx, cfg), nil
}
