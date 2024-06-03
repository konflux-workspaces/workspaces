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

	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacescomv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	KubespaceNamespace  string
	WorkspacesNamespace string
}

var (
	ErrNonTransient = fmt.Errorf("object non reconcilable")
)

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=spaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=spacebindings,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=workspaces.konflux.io,resources=internalworkspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workspaces.konflux.io,resources=internalworkspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workspaces.konflux.io,resources=internalworkspaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	w := workspacescomv1alpha1.InternalWorkspace{}
	if err := r.Client.Get(ctx, req.NamespacedName, &w); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, r.ensureSpaceIsDeleted(ctx, req.Name)
		}
		return ctrl.Result{}, err
	}

	if err := r.ensureOwnerAccessToWorkspace(ctx, &w); err != nil {
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

func (r *WorkspaceReconciler) ensureOwnerAccessToWorkspace(ctx context.Context, w *workspacescomv1alpha1.InternalWorkspace) error {
	// create role
	rn := fmt.Sprintf("%s:owner", w.Name)
	{
		or := rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: w.Namespace,
				Name:      rn,
			},
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &or, func() error {
			or.Rules = []rbacv1.PolicyRule{
				{
					Verbs:         []string{"get", "update", "delete"},
					APIGroups:     []string{workspacesv1alpha1.GroupVersion.Group},
					Resources:     []string{"internalworkspaces"},
					ResourceNames: []string{w.Name},
				},
			}

			return controllerutil.SetOwnerReference(w, &or, r.Scheme)
		}); err != nil {
			return err
		}
	}

	// create role binding
	{
		orb := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: w.Namespace,
				Name:      rn,
			},
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &orb, func() error {
			orb.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     rn,
			}
			orb.Subjects = []rbacv1.Subject{
				{
					APIGroup:  "rbac.authorization.k8s.io",
					Kind:      "User",
					Name:      w.Spec.Owner.Id,
					Namespace: w.Namespace,
				},
			}

			return controllerutil.SetOwnerReference(w, &orb, r.Scheme)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *WorkspaceReconciler) ensureWorkspaceVisibilityIsSatisfied(ctx context.Context, w workspacescomv1alpha1.InternalWorkspace) error {
	s := toolchainv1alpha1.SpaceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-community", w.Name),
			Namespace: r.KubespaceNamespace,
		},
	}
	l := log.FromContext(ctx).WithValues(
		"workspace", w.Name,
		"workspace-namespace", w.Namespace,
		"space-binding", s.Name,
		"space-binding-namespace", s.Namespace,
	)

	switch w.Spec.Visibility {
	case workspacescomv1alpha1.InternalWorkspaceVisibilityCommunity:
		l.Info("ensuring spacebinding exists")
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &s, func() error {
			s.Spec.Space = w.Name
			s.Spec.MasterUserRecord = workspacescomv1alpha1.PublicViewerName
			s.Spec.SpaceRole = "viewer"
			return nil
		})
		return err
	case workspacescomv1alpha1.InternalWorkspaceVisibilityPrivate:
		l.Info("ensuring spacebinding doesn't exist")
		return client.IgnoreNotFound(r.Client.Delete(ctx, &s))
	default:
		return fmt.Errorf("%w: invalid workspace visibility value", ErrNonTransient)
	}
}

func (r *WorkspaceReconciler) ensureSpaceIsPresent(ctx context.Context, w workspacescomv1alpha1.InternalWorkspace) error {
	// skip if home workspace
	if ll := w.GetLabels(); ll != nil && ll[workspacesv1alpha1.LabelHomeWorkspace] != "" {
		return nil
	}

	s := toolchainv1alpha1.Space{ObjectMeta: metav1.ObjectMeta{Name: w.Name, Namespace: r.KubespaceNamespace}}
	l := log.FromContext(ctx).WithValues("space", s.Name, "space-namespace", r.KubespaceNamespace)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, &s, func() error {
		ll := s.GetLabels()
		if ll == nil {
			ll = map[string]string{}
		}
		ll[toolchainv1alpha1.SpaceCreatorLabelKey] = w.Spec.Owner.Id

		s.SetLabels(ll)
		return nil
	})
	if err != nil {
		l.Error(err, "error creating/updating space")
		return err
	}

	l.Info("space created/updated")
	return nil
}

func (r *WorkspaceReconciler) ensureSpaceIsDeleted(ctx context.Context, name string) error {
	s := toolchainv1alpha1.Space{}
	t := types.NamespacedName{Name: name, Namespace: r.KubespaceNamespace}
	l := log.FromContext(ctx).WithValues("space", t.Name, "space-namespace", t.Namespace)
	l.Info("retrieving space for deletion")
	if err := r.Get(ctx, t, &s); err != nil {
		if kerrors.IsNotFound(err) {
			l.Info("space not found")
			return nil
		}
		return err
	}

	l.Info("deleting space")
	return client.IgnoreNotFound(r.Delete(ctx, &s))
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacescomv1alpha1.InternalWorkspace{}).
		Watches(&toolchainv1alpha1.SpaceBinding{}, handler.EnqueueRequestsFromMapFunc(r.mapSpaceBindingToWorkspace)).
		Complete(r)
}

func (r *WorkspaceReconciler) mapSpaceBindingToWorkspace(ctx context.Context, o client.Object) []reconcile.Request {
	sb, ok := o.(*toolchainv1alpha1.SpaceBinding)
	if !ok {
		return nil
	}

	sn, ok := sb.GetLabels()[toolchainv1alpha1.SpaceBindingSpaceLabelKey]
	if !ok {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      sn,
				Namespace: r.WorkspacesNamespace,
			},
		},
	}
}
