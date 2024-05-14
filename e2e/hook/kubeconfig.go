package hook

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	"github.com/konflux-workspaces/workspaces/e2e/pkg/kubeconfig"
)

func injectUnauthKubeconfig(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	cfg, err := kubeconfig.BuildRESTConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("error building config: %v", err))
	}

	return tcontext.InjectUnauthKubeconfig(ctx, cfg), nil
}
