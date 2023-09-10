package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}

const (
	ResourceKindWorkspace     = "Workspace"
	ResourceSingularWorkspace = "workspace"
	ResourcePluralWorkspace   = "workspaces"
	WorkspaceLabel            = "horizon.io/workspace"
)

// +genclient
// +kubebuilder:object:root=true
// +genclient:nonNamespaced

// Workspace is the Schema for the workspaces API
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="tenant",scope="Cluster"
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceSpec   `json:"spec,omitempty"`
	Status            WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +genclient:nonNamespaced

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

type WorkspaceSpec struct {
	Manager string `json:"manager,omitempty"`
}

type WorkspaceStatus struct {
}
