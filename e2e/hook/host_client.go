package hook

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/konflux-workspaces/workspaces/e2e/pkg/cli"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
	wrest "github.com/konflux-workspaces/workspaces/e2e/pkg/rest"
)

func injectHostClient(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	c, err := wrest.BuildDefaultHostClient()
	if err != nil {
		return nil, fmt.Errorf("error building client: %v", err)
	}

	tc := cli.New(c, sc.Id)
	return tcontext.InjectHostClient(ctx, tc), nil
}
