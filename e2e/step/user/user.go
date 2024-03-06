package user

import "github.com/cucumber/godog"

func RegisterSteps(ctx *godog.ScenarioContext) {
	// given
	ctx.Given(`^An user is onboarded$`, givenAnUserIsOnboarded)

	// when
	ctx.When(`^An user onboards$`, whenAnUserOnboards)

	ctx.When(`^The user requests the list of workspaces$`, whenUserRequestsTheListOfWorkspaces)
	ctx.When(`^The user requests their default workspace$`, whenUserRequestsTheirDefaultWorkspace)

	ctx.When(`^The user changes workspace visibility to "([^"]*)"$`, whenTheUserChangesWorkspaceVisibilityTo)
	ctx.When(`^The user requests a new community workspace$`, whenUserRequestsANewCommunityWorkspace)
	ctx.When(`^The user requests a new private workspace$`, whenUserRequestsANewPrivateWorkspace)

	// then
	ctx.Then(`^The user retrieves a list of workspaces containing just the default one$`, thenTheUserRetrievesAListOfWorkspacesContainingJustTheDefaultOne)
	ctx.Then(`^The user retrieves their default workspace$`, thenTheUserRetrievesTheirDefaultWorkspace)
}
