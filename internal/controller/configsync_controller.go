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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	sources "github.com/joe-bresee/config-synchronizer-operator/internal/sources/git"
)

// ConfigSyncReconciler reconciles a ConfigSync object
type ConfigSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=configs.example.io,resources=configsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=configs.example.io,resources=configsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=configs.example.io,resources=configsyncs/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

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
		setCondition(&configSync.Status, "Degraded", metav1.ConditionTrue, "InvalidSource", "Exactly one source must be specified")
		_ = r.Status().Update(ctx, &configSync)
		return ctrl.Result{}, nil
	}

	var revisionSHA string
	if configSync.Spec.Source.Git != nil {
		var err error
		revisionSHA, err = sources.CloneOrUpdate(
			ctx,
			r.Client,
			configSync.Spec.Source.Git.RepoURL,
			configSync.Spec.Source.Git.Revision,
			configSync.Spec.Source.Git.Branch,
			configSync.Spec.Source.Git.AuthMethod,
			configSync.Spec.Source.Git.AuthSecretRef,
		)
		if err != nil {
			setCondition(&configSync.Status, "Degraded", metav1.ConditionTrue, "GitCloneFailed", err.Error())
			_ = r.Status().Update(ctx, &configSync)
			return ctrl.Result{}, err
		}

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
	configSync.Status.LastSyncedTime = &metav1.Time{Time: metav1.Now().Time}
	configSync.Status.AppliedTargets = len(configSync.Spec.Targets)
	configSync.Status.SourceRevision = revisionSHA

	if err := r.Status().Update(ctx, &configSync); err != nil {
		return ctrl.Result{}, err
	}
	// Step 5: Requeue: you may want to requeue periodically (RefreshInterval) or immediately on failures.
	var requeueAfter time.Duration
	if configSync.Spec.RefreshInterval != "" {
		d, err := time.ParseDuration(configSync.Spec.RefreshInterval)
		if err != nil {
			// Invalid duration string — mark degraded and do not requeue.
			setCondition(&configSync.Status, "Degraded", metav1.ConditionTrue, "InvalidRefreshInterval", "RefreshInterval must be a valid duration string (e.g. \"30s\", \"5m\")")
			_ = r.Status().Update(ctx, &configSync)
			return ctrl.Result{}, nil
		}
		requeueAfter = d
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configsv1alpha1.ConfigSync{}).
		Named("configsync").
		Complete(r)
}

func setCondition(status *configsv1alpha1.ConfigSyncStatus, conditionType string, statusValue metav1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	for i, c := range status.Conditions {
		if c.Type == conditionType {
			status.Conditions[i] = metav1.Condition{
				Type:               conditionType,
				Status:             statusValue,
				Reason:             reason,
				Message:            message,
				LastTransitionTime: now,
			}
			return
		}
	}
	status.Conditions = append(status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             statusValue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	})
}

// DONT FORGET ADD ERR HANDLING TO DEGRADED FOR ANY STEP POSSIBLE OF FAILING
