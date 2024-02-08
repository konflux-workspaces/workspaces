package hook

import (
	"github.com/cucumber/godog"
)

func InjectHooks(ctx *godog.ScenarioContext) {
	injectHookSetup(ctx)
	injectHookCleanup(ctx)
}

func injectHookSetup(ctx *godog.ScenarioContext) {
	ctx.Before(injectHostClient)
	ctx.Before(createAndInjectTestNamespace)
}

func injectHookCleanup(ctx *godog.ScenarioContext) {
	ctx.After(deleteTestNamespace)
}
