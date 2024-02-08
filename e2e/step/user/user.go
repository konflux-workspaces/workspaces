package user

import "github.com/cucumber/godog"

func RegisterSteps(ctx *godog.ScenarioContext) {
	// given
	ctx.Given(`^An user is onboarded$`, givenAnUserIsOnboarded)

	// when
	ctx.When(`^An user onboards$`, whenAnUserOnboards)

	// then
}
