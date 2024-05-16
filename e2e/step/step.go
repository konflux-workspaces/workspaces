package step

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"
	"github.com/konflux-workspaces/workspaces/e2e/step/workspace"
)

func InjectSteps(ctx *godog.ScenarioContext) {
	workspace.RegisterSteps(ctx)
	user.RegisterSteps(ctx)

	ctx.Step(`^fail$`, func(ctx context.Context) (context.Context, error) { return ctx, fmt.Errorf("fail as you ordered") })
}
