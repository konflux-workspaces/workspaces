package cache

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// IndexKeyInternalWorkspaceDisplayName key for InternalWorkspace's indexer on field DisplayName
	IndexKeyInternalWorkspaceDisplayName string = "display-name"
	// IndexKeyInternalWorkspaceVisibility key for InternalWorkspace's indexer on field Visibility
	IndexKeyInternalWorkspaceVisibility string = "visibility"
	// IndexKeyInternalWorkspaceOwnerUsername key for InternalWorkspace's indexer on field for Owner's Username
	IndexKeyInternalWorkspaceOwnerUsername string = "owner.username"
	// IndexKeyInternalWorkspaceOwnerEmail key for InternalWorkspace's indexer on field for Owner's Email
	IndexKeyInternalWorkspaceOwnerEmail string = "owner.email"
	// IndexKeyInternalWorkspaceOwnerSub key for InternalWorkspace's indexer on field for Owner's Sub
	IndexKeyInternalWorkspaceOwnerSub string = "owner.sub"
	// IndexKeyInternalWorkspaceSpaceName key for InternalWorkspace's indexer on field for Space's name
	IndexKeyInternalWorkspaceSpaceName string = "space.name"

	// IndexKeyUserComplaintName key for InternalWorkspace's indexer on field for UserSignup's ComplaintName
	IndexKeyUserComplaintName string = "status.complaintName"
)

var UserSignupIndexers = map[string]client.IndexerFunc{
	IndexKeyUserComplaintName: newSingleFieldIndexer(func(u *toolchainv1alpha1.UserSignup) string {
		return u.Status.CompliantUsername
	}),
}

var InternalWorkspacesIndexers = map[string]client.IndexerFunc{
	IndexKeyInternalWorkspaceDisplayName: newSingleFieldIndexer(func(w *workspacesv1alpha1.InternalWorkspace) string {
		return w.Spec.DisplayName
	}),
	IndexKeyInternalWorkspaceVisibility: newSingleFieldIndexer(func(w *workspacesv1alpha1.InternalWorkspace) string {
		return string(w.Spec.Visibility)
	}),
	IndexKeyInternalWorkspaceOwnerUsername: newSingleFieldIndexer(func(w *workspacesv1alpha1.InternalWorkspace) string {
		return w.Status.Owner.Username
	}),
	IndexKeyInternalWorkspaceSpaceName: newSingleFieldIndexer(func(w *workspacesv1alpha1.InternalWorkspace) string {
		return w.Status.Space.Name
	}),
	IndexKeyInternalWorkspaceOwnerEmail: newSingleFieldIndexer(func(w *workspacesv1alpha1.InternalWorkspace) string {
		return w.Spec.Owner.JwtInfo.Email
	}),
	IndexKeyInternalWorkspaceOwnerSub: newSingleFieldIndexer(func(w *workspacesv1alpha1.InternalWorkspace) string {
		return w.Spec.Owner.JwtInfo.Sub
	}),
}

func newSingleFieldIndexer[T client.Object](f func(T) string) func(client.Object) []string {
	return func(obj client.Object) []string {
		t, ok := obj.(T)
		if !ok {
			return nil
		}

		return []string{f(t)}
	}
}
