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
	// Get
	Get(ctx context.Context, obj client.Object, name string) error
	// List
	List(ctx context.Context, objs client.ObjectList) error
	// ListOwned
	ListOwned(ctx context.Context, user string, objs client.ObjectList) error
	// ListShared
	ListShared(ctx context.Context, user string, objs client.ObjectList) error
	// ListOwnedAndShared returns all the
	ListOwnedAndShared(ctx context.Context, user string, objs client.ObjectList) error
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
