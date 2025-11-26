/*
Copyright 2025.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
)

// ConfigSyncReconciler reconciles a ConfigSync object
type ConfigSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=configs.example.io,resources=configsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=configs.example.io,resources=configsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=configs.example.io,resources=configsyncs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *ConfigSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Logic of our configsync controller is:

	// Step 1: You always need to Get the object from the API (r.Client.Get(ctx, req.NamespacedName, &configSync)).
	var configSync configsv1alpha1.ConfigSync
	if err := r.Get(ctx, req.NamespacedName, &configSync); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Step 2: Handling the source — Apparently there is no XOR validation in kubebuilder so doing this. Only one source can be set per controller.
	sourceSet := 0
	if configSync.Spec.Source.Git != nil {
		sourceSet++
	}
	if configSync.Spec.Source.ConfigMapRef != nil {
		sourceSet++
	}
	if configSync.Spec.Source.SecretRef != nil {
		sourceSet++
	}
	if sourceSet != 1 {
		// Invalid spec: either none or multiple sources set. You should update status conditions to reflect this error.
		return ctrl.Result{}, nil
	}

	if configSync.Spec.Source.Git != nil {
		// Handle Git source logic here
	}
	if configSync.Spec.Source.ConfigMapRef != nil {
		// Handle ConfigMap source logic here
	}
	if configSync.Spec.Source.SecretRef != nil {
		// Handle Secret source logic here
	}

	// Step 3: The “apply to cluster” logic will vary: for a Git source, you need to clone/read files, parse manifests, then Apply them (usually with client.Apply). For ConfigMap/Secret sources, just read them and apply.
	for _, target := range configSync.Spec.Targets {
		// Apply logic for each target
		_ = target // Placeholder to avoid unused variable error
	}
	// Step 4: Update status fields: LastSyncedTime, SourceRevision, AppliedTargets, Conditions. This is standard K8s convention.
	configSync.Status.LastSyncedTime = &metav1.Time{Time: ctrl.Now()}
	configSync.Status.AppliedTargets = len(configSync.Spec.Targets)
	configSync.Status.SourceRevision = revisionSHA
	// conditions: available, progressing, degraded
	// You would typically use helper functions to set conditions properly.???
	if err := r.Status().Update(ctx, &configSync); err != nil {
		return ctrl.Result{}, err
	}

	// Step 5: Requeue: you may want to requeue periodically (RefreshInterval) or immediately on failures.

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configsv1alpha1.ConfigSync{}).
		Named("configsync").
		Complete(r)
}
