/*
Copyright 2024 The UserSignups Authors.

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

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

const (
	LabelHomeWorkspace  string = "workspaces.io/home-workspace"
	LabelWorkspaceOwner string = "workspaces.io/owner"
)

// UserSignupReconciler reconciles a Workspace object
type UserSignupReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	WorkspacesNamespace string
}

//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=usersignups,verbs=get;list;watch
//+kubebuilder:rbac:groups=workspaces.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *UserSignupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	u := toolchainv1alpha1.UserSignup{}
	if err := r.Client.Get(ctx, req.NamespacedName, &u); err != nil {
		if kerrors.IsNotFound(err) {
			err := r.ensureWorkspaceIsDeleted(ctx, req.Name)
			if errors.Is(err, ErrNonTransient) {
				l.Error(err, "can not delete workspace", "user", req.Name)
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if err := r.ensureWorkspaceIsPresentForHomeSpace(ctx, u); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *UserSignupReconciler) ensureWorkspaceIsPresentForHomeSpace(ctx context.Context, u toolchainv1alpha1.UserSignup) error {
	w := &workspacesv1alpha1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: u.Status.HomeSpace, Namespace: r.WorkspacesNamespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, w, func() error {
		log.FromContext(ctx).Info("creating/updating workspace", "workspace", w)
		ll := w.GetLabels()
		if ll == nil {
			ll = map[string]string{}
		}
		ll[LabelHomeWorkspace] = u.Name
		ll[LabelWorkspaceOwner] = u.Name
		w.Labels = ll

		w.Spec.Visibility = workspacesv1alpha1.WorkspaceVisibilityPrivate
		w.Spec.Owner = workspacesv1alpha1.Owner{
			Id: u.Name,
		}
		return nil
	})
	if err != nil {
		log.FromContext(ctx).Error(err, "error creating or updating workspace", "workspace", w)
	}

	return err
}

func (r *UserSignupReconciler) ensureWorkspaceIsDeleted(ctx context.Context, name string) error {
	lr, err := labels.NewRequirement(LabelHomeWorkspace, selection.Equals, []string{name})
	if err != nil {
		return errors.Join(ErrNonTransient, err)
	}
	ls := labels.NewSelector()
	ls.Add(*lr)

	w := workspacesv1alpha1.Workspace{}
	if err := r.DeleteAllOf(ctx, &w, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: ls,
			Namespace:     r.WorkspacesNamespace,
		},
	}); err != nil {
		return client.IgnoreNotFound(err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserSignupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolchainv1alpha1.UserSignup{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
			u := object.(*toolchainv1alpha1.UserSignup)
			return u.Status.HomeSpace != ""
		}))).
		Complete(r)
}
