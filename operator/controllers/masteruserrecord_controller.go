/*
Copyright 2024 The MasterUserRecords Authors.

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

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	workspacesv1alpha1 "github.com/konflux-workspaces/workspaces/operator/api/v1alpha1"
)

// MasterUserRecordReconciler reconciles a Workspace object
type MasterUserRecordReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	WorkspacesNamespace string
}

//+kubebuilder:rbac:groups=toolchain.dev.openshift.com,resources=masteruserrecords,verbs=get;list;watch
//+kubebuilder:rbac:groups=workspaces.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *MasterUserRecordReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	m := toolchainv1alpha1.MasterUserRecord{}
	if err := r.Client.Get(ctx, req.NamespacedName, &m); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, r.ensureWorkspaceIsDeleted(ctx, req.Name)
		}
		return ctrl.Result{}, err
	}

	if err := r.ensureWorkspaceIsPresent(ctx, m); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *MasterUserRecordReconciler) ensureWorkspaceIsPresent(ctx context.Context, m toolchainv1alpha1.MasterUserRecord) error {
	w := &workspacesv1alpha1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: m.Name, Namespace: r.WorkspacesNamespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, w, func() error {
		log.FromContext(ctx).Info("creating/updating workspace", "workspace", w)
		w.Spec.Visibility = workspacesv1alpha1.WorkspaceVisibilityPrivate
		w.Spec.Owner = workspacesv1alpha1.Owner{
			Id: m.Name,
		}
		return nil
	})
	if err != nil {
		log.FromContext(ctx).Error(err, "error creating or updating workspace", "workspace", w)
	}

	return err
}

func (r *MasterUserRecordReconciler) ensureWorkspaceIsDeleted(ctx context.Context, name string) error {
	w := workspacesv1alpha1.Workspace{}
	t := types.NamespacedName{Name: name, Namespace: r.WorkspacesNamespace}
	if err := r.Get(ctx, t, &w); err != nil {
		return client.IgnoreNotFound(err)
	}

	return client.IgnoreNotFound(r.Delete(ctx, &w))
}

// SetupWithManager sets up the controller with the Manager.
func (r *MasterUserRecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolchainv1alpha1.MasterUserRecord{}).
		Complete(r)
}
