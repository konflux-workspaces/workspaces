/*
Copyright 2024 The Workspaces Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacescomv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	ErrNonTransient = fmt.Errorf("object non reconcilable")
)

//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=spaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=spacebindings,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=workspaces.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workspaces.io,resources=workspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workspaces.io,resources=workspaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	w := workspacescomv1alpha1.Workspace{}
	if err := r.Client.Get(ctx, req.NamespacedName, &w); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, r.ensureSpaceIsDeleted(ctx, req.NamespacedName)
		}
		return ctrl.Result{}, err
	}

	if err := r.ensureSpaceIsPresent(ctx, w); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ensureWorkspaceVisibilityIsSatisfied(ctx, w); err != nil {
		if errors.Is(err, ErrNonTransient) {
			l.Error(err, "non transient error ensuring workspace visibility is satisfied")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *WorkspaceReconciler) ensureWorkspaceVisibilityIsSatisfied(ctx context.Context, w workspacescomv1alpha1.Workspace) error {

	s := toolchainv1alpha1.SpaceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-community", w.Name),
			Namespace: w.Namespace,
		},
	}
	switch w.Spec.Visibility {
	case workspacescomv1alpha1.WorkspaceVisibilityCommunity:
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &s, func() error {
			s.Spec.Space = w.Name
			s.Spec.MasterUserRecord = "public-viewer"
			s.Spec.SpaceRole = "viewer"
			return nil
		})
		return err
	case workspacescomv1alpha1.WorkspaceVisibilityPrivate:
		return nil
	default:
		return fmt.Errorf("%w: invalid workspace visibility value", ErrNonTransient)
	}
}

func (r *WorkspaceReconciler) ensureSpaceIsPresent(ctx context.Context, w workspacescomv1alpha1.Workspace) error {
	s := toolchainv1alpha1.Space{ObjectMeta: metav1.ObjectMeta{Name: w.Name, Namespace: w.Namespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &s, func() error {
		ll := s.GetLabels()
		if len(ll) == 0 {
			ll = map[string]string{}
		}
		ll[toolchainv1alpha1.SpaceCreatorLabelKey] = w.Spec.Owner.Id

		s.SetLabels(ll)
		return nil
	})
	return err
}

func (r *WorkspaceReconciler) ensureSpaceIsDeleted(ctx context.Context, nn types.NamespacedName) error {
	s := toolchainv1alpha1.Space{}
	if err := r.Get(ctx, nn, &s); err != nil {
		return client.IgnoreNotFound(err)
	}

	return client.IgnoreNotFound(r.Delete(ctx, &s))
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacescomv1alpha1.Workspace{}).
		Complete(r)
}
