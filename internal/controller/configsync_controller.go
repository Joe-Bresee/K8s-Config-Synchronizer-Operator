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
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	apply "github.com/joe-bresee/config-synchronizer-operator/internal/apply"
	source "github.com/joe-bresee/config-synchronizer-operator/internal/sources"
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
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

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
	log := logf.FromContext(ctx)

	// Fetch the CR
	var configSync configsv1alpha1.ConfigSync
	if err := r.Get(ctx, req.NamespacedName, &configSync); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// --------------------------------------------------------------
	// Step 1: Fetch source and determine revision
	// --------------------------------------------------------------
	revisionSHA, sourcePath, err := source.FetchSource(&configSync, ctx, r.Client)
	if err != nil {
		setCondition(&configSync.Status, "Degraded", metav1.ConditionTrue, "SourceFetchFailed", err.Error())
		_ = r.Status().Update(ctx, &configSync)
		return ctrl.Result{}, err
	}

	// Cleanup temp directories if not Git
	if configSync.Spec.Source.Git == nil && sourcePath != "" {
		defer os.RemoveAll(sourcePath)
	}

	// --------------------------------------------------------------
	// Step 2: Determine if we need to apply
	// --------------------------------------------------------------
	previousRevision := configSync.Status.SourceRevision
	shouldApply := previousRevision != revisionSHA

	if shouldApply {
		log.Info("Source revision changed — applying", "old", previousRevision, "new", revisionSHA)

		// Apply to all targets
		for _, target := range configSync.Spec.Targets {
			if err := apply.ApplyTarget(ctx, r.Client, sourcePath, target); err != nil {
				setCondition(&configSync.Status, "Degraded", metav1.ConditionTrue, "ApplyFailed", err.Error())
				_ = r.Status().Update(ctx, &configSync)
				return ctrl.Result{}, err
			}
		}

		// Apply succeeded — mark condition
		setCondition(&configSync.Status, "Degraded", metav1.ConditionFalse, "ApplySucceeded", "All targets applied successfully")
	} else {
		log.Info("No changes detected — skipping apply", "revision", revisionSHA)
	}

	// --------------------------------------------------------------
	// Step 3: Always update status AFTER apply decision
	// --------------------------------------------------------------
	configSync.Status.LastSyncedTime = &metav1.Time{Time: time.Now()}
	configSync.Status.AppliedTargets = len(configSync.Spec.Targets)
	configSync.Status.SourceRevision = revisionSHA
	configSync.Status.SourcePath = sourcePath

	if err := r.Status().Update(ctx, &configSync); err != nil {
		return ctrl.Result{}, err
	}

	// --------------------------------------------------------------
	// Step 4: Compute requeue interval
	// --------------------------------------------------------------
	requeueAfter := 30 * time.Second // default
	if configSync.Spec.RefreshInterval != "" {
		d, err := time.ParseDuration(configSync.Spec.RefreshInterval)
		if err != nil {
			setCondition(&configSync.Status, "Degraded", metav1.ConditionTrue, "InvalidRefreshInterval", "RefreshInterval must be a valid duration (e.g. 30s, 5m)")
			_ = r.Status().Update(ctx, &configSync)
			return ctrl.Result{}, nil
		}
		requeueAfter = d
	}

	log.Info("Reconcile completed", "requeueAfter", requeueAfter.String())
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// SetupWithManager
// --------------------------------------------------------------
func (r *ConfigSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configsv1alpha1.ConfigSync{}).
		Named("configsync").
		Complete(r)
}

// Condition helper
// --------------------------------------------------------------
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
