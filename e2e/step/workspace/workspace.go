package workspace

import "github.com/cucumber/godog"

func RegisterSteps(ctx *godog.ScenarioContext) {
	// given
	ctx.Given(`^A community workspace exists for an user$`, givenACommunityWorkspaceExists)
	ctx.Given(`^A private workspace exists for an user$`, givenAPrivateWorkspaceExists)

	ctx.Given(`^Default workspace is created for them$`, givenDefaultWorkspaceIsCreatedForThem)
	ctx.Given(`^Default workspace is created for "([^"]*)"$`, givenDefaultWorkspaceIsCreatedForCustomUser)
	ctx.Given(`^Workspace\'s Space has cluster URL set$`, givenWorkspaceHasClusterURLSet)
	ctx.Given(`^Workspace\'s Space has no cluster URL set$`, givenWorkspaceHasNoClusterURLSet)

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
	ctx.Then(`^"([^"]*)" can not change workspace visibility to "([^"]*)"$`, thenUserCanNotChangeVisibilityTo)
	ctx.Then(`^"([^"]*)" can not patch workspace visibility to "([^"]*)"$`, thenUserCanNotPatchVisibilityTo)

	ctx.Then(`^The workspace visibility is updated to "([^"]*)"$`, thenTheWorkspaceVisibilityIsUpdatedTo)
	ctx.Then(`^Workspace has cluster URL in status$`, thenDefaultWorkspaceHasClusterURLInStatus)
	ctx.Then(`^Workspace has no cluster URL in status$`, thenDefaultWorkspaceHasNoClusterURLInStatus)
}
