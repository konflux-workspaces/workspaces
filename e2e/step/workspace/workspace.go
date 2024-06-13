package workspace

import "github.com/cucumber/godog"

func RegisterSteps(ctx *godog.ScenarioContext) {
	// given
	ctx.Given(`^A community workspace exists for an user$`, givenACommunityWorkspaceExists)
	ctx.Given(`^A private workspace exists for an user$`, givenAPrivateWorkspaceExists)

	ctx.Given(`^Default workspace is created for them$`, givenDefaultWorkspaceIsCreatedForThem)

	// when
	ctx.When(`^A workspace is created for an user$`, whenAWorkspaceIsCreatedForUser)
	ctx.When(`^The owner changes visibility to community$`, whenOwnerChangesVisibilityToCommunity)
	ctx.When(`^The owner changes visibility to private$`, whenOwnerChangesVisibilityToPrivate)

	// then
	ctx.Then(`^The workspace is readable for everyone$`, thenTheWorkspaceIsReadableForEveryone)
	ctx.Then(`^The workspace is readable only for the ones directly granted access to$`, thenTheWorkspaceIsReadableOnlyForGranted)
	ctx.Then(`^A community workspace is created$`, thenACommunityWorkspaceIsCreated)
	ctx.Then(`^A private workspace is created$`, thenAPrivateWorkspaceIsCreated)
	ctx.Then(`^Default workspace is created for them$`, thenDefaultWorkspaceIsCreatedForThem)
	ctx.Then(`^The owner is granted admin access to the workspace$`, thenTheOwnerIsGrantedAdminAccessToTheWorkspace)
	ctx.Then(`^The workspace visibility is set to "([^"]*)"$`, thenTheWorkspaceVisibilityIsSetTo)

	ctx.Then(`^The workspace visibility is updated to "([^"]*)"$`, thenTheWorkspaceVisibilityIsUpdatedTo)
}
