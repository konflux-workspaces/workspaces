package cqrs

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReadWriteModel interface {
	Reader
	Writer
}

// Reader allows access to the ReadModel
type Reader interface {
	ReadUserWorkspace(ctx context.Context, user string, space string, obj *workspacesv1alpha1.Workspace, opts ...client.GetOption) error
	ListUserWorkspaces(ctx context.Context, user string, objs *workspacesv1alpha1.WorkspaceList, opts ...client.ListOption) error
}

// Writer allows access to the WriteModel
type Writer interface {
	// Create creates an object
	Create(ctx context.Context, obj client.Object) error
	// Delete deletes an object
	Delete(ctx context.Context, key client.ObjectKey) error
	// Update updates an object
	Update(ctx context.Context, obj client.Object) error
}
