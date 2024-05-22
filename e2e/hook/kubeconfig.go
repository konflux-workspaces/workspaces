package hook

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	wrest "github.com/konflux-workspaces/workspaces/e2e/pkg/rest"
)

func injectUnauthKubeconfig(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	cfg, err := wrest.NewDefaultClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error building unauthenticated config: %v", err)
	}

	cfg.KeyData = nil
	cfg.KeyFile = ""
	cfg.BearerToken = ""
	cfg.BearerTokenFile = ""

	return tcontext.InjectUnauthKubeconfig(ctx, cfg), nil
}
