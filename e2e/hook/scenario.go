package hook

import (
	"context"

	"github.com/cucumber/godog"
	tcontext "github.com/konflux-workspaces/workspaces/e2e/pkg/context"
)

func injectScenarioId(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	return tcontext.InjectScenarioId(ctx, sc.Id), nil
}
