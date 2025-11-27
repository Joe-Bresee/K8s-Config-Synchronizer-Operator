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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SourceSpec describes the source of configuration data for a ConfigSync.
// Only one of the fields should be set.
type SourceSpec struct {
	// Git references a Git repository and path to read the configuration from.
	Git *GitSource `json:"git,omitempty"`
	// ConfigMapRef points to a ConfigMap in the cluster to use as the source. Must be in form key: filename, data: file contents. Use data fields and NOT binaryData.
	ConfigMapRef *ObjectRef `json:"configMapRef,omitempty"`
	// SecretRef points to a Secret in the cluster to use as the source.
	SecretRef *ObjectRef `json:"secretRef,omitempty"`
}

type GitSource struct {
	// Repo is the HTTPS or SSH URL of the Git repository to clone (for example
	// `https://github.com/myorg/configs.git`). This field is required when `git` is used.
	// +kubebuilder:validation:Required
	RepoURL string `json:"repoURL"`

	// Path is the repository-relative path to the file containing the configuration
	// (for example `config/app.yaml`). This field is required when `git` is used.
	// +kubebuilder:validation:Required
	Path string `json:"path"`

	// Branch is the Git branch to checkout. If unspecified, the operator will
	// default to the repository's default branch.
	// +optional
	Branch string `json:"branch,omitempty"`

	// Revision is an optional Git revision (branch, tag, or commit SHA). If
	// unspecified, the operator will default to the repository's default branch.
	// +optional
	Revision string `json:"revision,omitempty"`

	// AuthMethod controls how the operator authenticates to the Git repository.
	// Allowed values are `ssh`, `https`, or `none`.
	// +kubebuilder:validation:Enum=ssh;https;none
	AuthMethod string `json:"authMethod,omitempty"`

	// AuthSecretRef references a Secret that contains an SSH private key or basic auth credentials when
	// `AuthMethod=ssh` or `AuthMethod=https`. The Secret should contain the key under a standard key
	// name (e.g., `id_rsa` for SSH or `username` and `password` for HTTPS).
	// +optional
	AuthSecretRef *ObjectRef `json:"authSecretRef,omitempty"`
}

type TargetRef struct {
	// Namespace is the namespace in which the target resource should be
	// created or updated.
	Namespace string `json:"namespace"`

	// Name is the name of the target ConfigMap or Secret.
	Name string `json:"name"`

	// Type is the type of the Kubernetes resource to write. Valid values are
	// `ConfigMap` or `Secret`.
	// +kubebuilder:validation:Enum=ConfigMap;Secret
	Type string `json:"type"`
}

type ObjectRef struct {
	// Name is the name of the referenced object (ConfigMap or Secret).
	Name string `json:"name"`

	// Namespace is the namespace of the referenced object.
	Namespace string `json:"namespace,omitempty"`
}

// ConfigSyncSpec defines the desired state of ConfigSync
type ConfigSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ConfigSync. Edit configsync_types.go to remove/update
	// Source defines where to fetch configuration data from. Only one source
	// field may be set (git, configMapRef, or secretRef).
	// +optional
	Source SourceSpec `json:"source"`

	// Targets is the list of target resources to apply the rendered
	// configuration to. Each target contains `namespace`, `name`, and `type`.
	// +kubebuilder:validation:MinItems=1
	Targets []TargetRef `json:"targets"`

	// RefreshInterval controls how frequently the operator should re-fetch
	// the source and re-apply targets. It uses Kubernetes duration format
	// (e.g. `10m`, `1h`). If omitted, the operator's default behavior applies.
	// +kubebuilder:validation:Pattern=^([0-9]+(\\.[0-9]+)?(ns|us|ms|s|m|h))+$
	RefreshInterval string `json:"refreshInterval,omitempty"`
}

// ConfigSyncStatus defines the observed state of ConfigSync.
type ConfigSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ConfigSync resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +optional
	// LastSyncedTime is the timestamp of the last successful sync operation.
	// +optional
	LastSyncedTime *metav1.Time `json:"lastSyncedTime,omitempty"`

	// SourceRevision records the source revision (for example a Git SHA) that
	// was applied during the last sync.
	// +optional
	SourceRevision string `json:"sourceRevision,omitempty"`

	// AppliedTargets is the number of targets that were successfully
	// created or updated during the last sync.
	// +optional
	AppliedTargets int `json:"appliedTargets,omitempty"`

	// Conditions represent the current state of the ConfigSync resource.
	// This follows the Kubernetes condition convention (type, status, reason,
	// message, lastTransitionTime).
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ConfigSync is the Schema for the configsyncs API
type ConfigSync struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of ConfigSync
	// +required
	Spec ConfigSyncSpec `json:"spec"`

	// status defines the observed state of ConfigSync
	// +optional
	Status ConfigSyncStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ConfigSyncList contains a list of ConfigSync
type ConfigSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ConfigSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigSync{}, &ConfigSyncList{})
}
