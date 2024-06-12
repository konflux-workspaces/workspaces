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

package controller

import (
	"context"
	"errors"
	"slices"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// UserSignupReconciler reconciles a Workspace object
type UserSignupReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	WorkspacesNamespace string
}

//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=usersignups,verbs=get;list;watch
//+kubebuilder:rbac:groups=workspaces.konflux.io,resources=internalworkspaces,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *UserSignupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("request", req)

	u := toolchainv1alpha1.UserSignup{}
	if err := r.Client.Get(ctx, req.NamespacedName, &u); err != nil {
		if kerrors.IsNotFound(err) {
			l.V(6).Info("UserSignup not found")
			if err := r.ensureWorkspaceIsDeleted(ctx, req.Name); err != nil {
				if errors.Is(err, ErrNonTransient) {
					l.Error(err, "can not delete workspace", "user", req.Name)
					return ctrl.Result{}, nil
				}
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		l.Error(err, "error retrieving UserSignup")
		return ctrl.Result{}, err
	}

	if err := r.ensureWorkspaceIsPresentForHomeSpace(ctx, u); err != nil {
		l.Error(err, "error ensuring InternalWorkspace is present for user's HomeSpace")
		return ctrl.Result{}, err
	}

	l.V(6).Info("InternalWorkspace is present for user")
	return ctrl.Result{}, nil
}

func (r *UserSignupReconciler) ensureWorkspaceIsPresentForHomeSpace(ctx context.Context, u toolchainv1alpha1.UserSignup) error {
	w := &workspacesv1alpha1.InternalWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      u.Status.HomeSpace,
			Namespace: r.WorkspacesNamespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, w, func() error {
		w.Spec.DisplayName = "default"
		w.Spec.Visibility = workspacesv1alpha1.InternalWorkspaceVisibilityPrivate
		w.Spec.Owner = workspacesv1alpha1.UserInfo{
			JwtInfo: workspacesv1alpha1.JwtInfo{
				Sub:               u.Spec.IdentityClaims.Sub,
				Email:             u.Spec.IdentityClaims.Email,
				UserId:            u.Spec.IdentityClaims.UserID,
				PreferredUsername: u.Spec.IdentityClaims.PreferredUsername,
				AccountId:         u.Spec.IdentityClaims.AccountID,
				Company:           u.Spec.IdentityClaims.Company,
				GivenName:         u.Spec.IdentityClaims.GivenName,
				FamilyName:        u.Spec.IdentityClaims.FamilyName,
			},
		}

		log.FromContext(ctx).Info("creating/updating workspace", "workspace", w)
		return nil
	})
	if err != nil {
		log.FromContext(ctx).Error(err, "error creating or updating workspace", "workspace", w)
	}

	// update status
	return nil
}

func (r *UserSignupReconciler) ensureWorkspaceIsDeleted(ctx context.Context, name string) error {
	// retrieve all InternalWorkspaces
	ww := workspacesv1alpha1.InternalWorkspaceList{}
	if err := r.List(ctx, &ww); err != nil {
		return err
	}

	// look for user's home InternalWorkspace
	if i := slices.IndexFunc(ww.Items, func(w workspacesv1alpha1.InternalWorkspace) bool {
		return w.Status.Space.IsHome && w.Status.Owner.Username == name
	}); i != -1 {
		// delete the user's Home InternalWorkspace
		return r.Delete(ctx, &ww.Items[i])
	}

	// workspace not found, nothing to delete
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserSignupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolchainv1alpha1.UserSignup{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
			u, ok := object.(*toolchainv1alpha1.UserSignup)
			return ok && u.Status.HomeSpace != ""
		}))).
		Complete(r)
}
