// Copyright 2021 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

package v1alpha1

import (
	"encoding/json"
	"fmt"

	helmv2beta2 "github.com/fluxcd/helm-controller/api/v2beta2"
	"github.com/fluxcd/pkg/apis/meta"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redpanda-data/redpanda-operator/src/go/k8s/api/redpanda/v1alpha2"
)

var RedpandaChartRepository = "https://charts.redpanda.com/"

// RedpandaSpec defines the desired state of the Redpanda cluster.
type RedpandaSpec struct {
	// Defines chart details, including the version and repository.
	ChartRef v1alpha2.ChartRef `json:"chartRef,omitempty"`
	// Defines the Helm values to use to deploy the cluster.
	ClusterSpec *RedpandaClusterSpec `json:"clusterSpec,omitempty"`
	// Migration flag that adjust Kubernetes core resources with annotation and labels, so
	// flux controller can import resources.
	// Doc: https://docs.redpanda.com/current/upgrade/migrate/kubernetes/operator/
	Migration *v1alpha2.Migration `json:"migration,omitempty"`
}

type RemediationStrategy string

// HelmUpgrade configures the behavior and strategy for Helm chart upgrades.
type HelmUpgrade struct {
	// Specifies the actions to take on upgrade failures. See https://pkg.go.dev/github.com/fluxcd/helm-controller/api/v2beta1#UpgradeRemediation.
	Remediation *helmv2beta2.UpgradeRemediation `json:"remediation,omitempty"`
	// Enables forceful updates during an upgrade.
	Force *bool `json:"force,omitempty"`
	// Specifies whether to preserve user-configured values during an upgrade.
	PreserveValues *bool `json:"preserveValues,omitempty"`
	// Specifies whether to perform cleanup in case of failed upgrades.
	CleanupOnFail *bool `json:"cleanupOnFail,omitempty"`
}

// Redpanda defines the CRD for Redpanda clusters.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=redpandas
// +kubebuilder:resource:shortName=rp
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
type Redpanda struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the Redpanda cluster.
	Spec RedpandaSpec `json:"spec,omitempty"`
	// Represents the current status of the Redpanda cluster.
	Status v1alpha2.RedpandaStatus `json:"status,omitempty"`
}

// RedpandaList contains a list of Redpanda objects.
// +kubebuilder:object:root=true
type RedpandaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Specifies a list of Redpanda resources.
	Items []Redpanda `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redpanda{}, &RedpandaList{})
}

func (in *Redpanda) GetHelmReleaseName() string {
	return in.Name
}

func (in *Redpanda) GetHelmRepositoryName() string {
	helmRepository := in.Spec.ChartRef.HelmRepositoryName
	if helmRepository == "" {
		helmRepository = "redpanda-repository"
	}
	return helmRepository
}

func (in *Redpanda) ValuesJSON() (*apiextensionsv1.JSON, error) {
	vyaml, err := json.Marshal(in.Spec.ClusterSpec)
	if err != nil {
		return nil, fmt.Errorf("could not convert spec to yaml: %w", err)
	}
	values := &apiextensionsv1.JSON{Raw: vyaml}

	return values, nil
}

// RedpandaReady registers a successful reconciliation of the given HelmRelease.
func RedpandaReady(rp *Redpanda) *Redpanda {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  metav1.ConditionTrue,
		Reason:  "RedpandaClusterDeployed",
		Message: "Redpanda reconciliation succeeded",
	}
	apimeta.SetStatusCondition(rp.GetConditions(), newCondition)
	rp.Status.LastAppliedRevision = rp.Status.LastAttemptedRevision
	return rp
}

// RedpandaNotReady registers a failed reconciliation of the given Redpanda.
func RedpandaNotReady(rp *Redpanda, reason, message string) *Redpanda {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	}
	apimeta.SetStatusCondition(rp.GetConditions(), newCondition)
	return rp
}

// RedpandaProgressing resets any failures and registers progress toward
// reconciling the given Redpanda by setting the meta.ReadyCondition to
// 'Unknown' for meta.ProgressingReason.
func RedpandaProgressing(rp *Redpanda) *Redpanda {
	rp.Status.Conditions = []metav1.Condition{}
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  metav1.ConditionUnknown,
		Reason:  meta.ProgressingReason,
		Message: "Reconciliation in progress",
	}
	apimeta.SetStatusCondition(rp.GetConditions(), newCondition)
	return rp
}

// GetConditions returns the status conditions of the object.
func (in *Redpanda) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *Redpanda) OwnerShipRefObj() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: in.APIVersion,
		Kind:       in.Kind,
		Name:       in.Name,
		UID:        in.UID,
	}
}

// GetMigrationConsoleName returns Console custom resource namespace which will be taken out from
// old reconciler, so that underlying resources could be migrated.
func (in *Redpanda) GetMigrationConsoleName() string {
	if in.Spec.Migration == nil {
		return ""
	}
	name := in.Spec.Migration.ConsoleRef.Name
	if name == "" {
		name = in.Name
	}
	return name
}

// GetMigrationConsoleNamespace returns Console custom resource name which will be taken out from
// old reconciler, so that underlying resources could be migrated.
func (in *Redpanda) GetMigrationConsoleNamespace() string {
	if in.Spec.Migration == nil {
		return ""
	}
	namespace := in.Spec.Migration.ConsoleRef.Namespace
	if namespace == "" {
		namespace = in.Namespace
	}
	return namespace
}

// GetMigrationClusterName returns Cluster custom resource namespace which will be taken out from
// old reconciler, so that underlying resources could be migrated.
func (in *Redpanda) GetMigrationClusterName() string {
	if in.Spec.Migration == nil {
		return ""
	}
	name := in.Spec.Migration.ClusterRef.Name
	if name == "" {
		name = in.Name
	}
	return name
}

// GetMigrationClusterNamespace returns Cluster custom resource name which will be taken out from
// old reconciler, so that underlying resources could be migrated.
func (in *Redpanda) GetMigrationClusterNamespace() string {
	if in.Spec.Migration == nil {
		return ""
	}
	namespace := in.Spec.Migration.ClusterRef.Namespace
	if namespace == "" {
		namespace = in.Namespace
	}
	return namespace
}
