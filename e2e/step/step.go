package step

import (
	"github.com/cucumber/godog"
	"github.com/konflux-workspaces/workspaces/e2e/step/user"
	"github.com/konflux-workspaces/workspaces/e2e/step/workspace"
)

func InjectSteps(ctx *godog.ScenarioContext) {
	workspace.RegisterSteps(ctx)
	user.RegisterSteps(ctx)
}
