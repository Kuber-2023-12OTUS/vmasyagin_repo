/*
Copyright 2024 Victor Masyagin.

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
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	otusv1 "mysql-operator/api/v1"
)

// MySQLReconciler reconciles a MySQL object
type MySQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=otus.homework,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=otus.homework,resources=mysqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=otus.homework,resources=mysqls/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services;persistentvolumes;persistentvolumeclaims,verbs=create;patch;update;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=create;patch;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MySQL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *MySQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	mysql := &otusv1.MySQL{}
	err := r.Get(ctx, req.NamespacedName, mysql)

	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		l.Error(err, "unable to fetch MySQL custom resource")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	if !mysql.DeletionTimestamp.IsZero() {
		l.Info(fmt.Sprintf("MySQL custom resource '%s' and its children resources are being deleted...", mysql.Name))
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if r.createPersistentVolume(ctx, mysql) != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	if r.createPersistentVolumeClaim(ctx, mysql) != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	if r.createService(ctx, mysql) != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	if r.createDeployment(ctx, mysql) != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&otusv1.MySQL{}).
		Complete(r)
}
