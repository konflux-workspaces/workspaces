package workspace

//go:generate mockgen -destination=mocks_generated_test.go -package=workspace_test . WorkspaceUpdater,WorkspaceReader,WorkspaceLister,WorkspaceCreator
