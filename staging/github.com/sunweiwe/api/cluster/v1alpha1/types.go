package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	ResourceKindCluster      = "Cluster"
	ResourcesSingularCluster = "cluster"
	ResourcesPluralCluster   = "clusters"

	HostCluster = "cluster-role.horizon.io/host"

	Finalizer = "finalizer.cluster.horizon.io"
)

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}

// +genclient
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +genclient:nonNamespaced
// +kubebuilder:printcolumn:name="Federated",type="boolean",JSONPath=".spec.joinFederation"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="Active",type="boolean",JSONPath=".spec.enable"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.kubernetesVersion"
// +kubebuilder:resource:scope=Cluster

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

type ClusterSpec struct {
	JoinFederation bool `json:"joinFederation,omitempty"`

	Enable bool `json:"enable,omitempty"`

	Provider string `json:"provider,omitempty"`

	Connection Connection `json:"connection,omitempty"`

	ExternalKubeAPIEnabled bool `json:"externalKubeAPIEnabled,omitempty"`
}

type ConnectionType string

const (
	ConnectionTypeDirect ConnectionType = "direct"
	ConnectionTypeProxy  ConnectionType = "proxy"
)

type Connection struct {
	Type ConnectionType `json:"type,omitempty"`

	HorizonAPIEndpoint string `json:"horizonAPIEndpoint,omitempty"`

	KubernetesAPIEndpoint string `json:"kubernetesAPIEndpoint,omitempty"`

	ExternalKubernetesAPIEndpoint string `json:"externalKubernetesAPIEndpoint,omitempty"`

	KubeConfig []byte `json:"kubeconfig,omitempty"`

	Token string `json:"token,omitempty"`

	KubernetesAPIServerPort uint16 `json:"kubernetesAPIServerPort,omitempty"`

	HorizonAPIServerPort uint16 `json:"horizonAPIServerPort,omitempty"`
}

type ClusterConditionType string

const (
	ClusterFederated ClusterConditionType = "Federated"

	ClusterReady ClusterConditionType = "Ready"

	ClusterKubeConfigCertExpiresInSevenDays ClusterConditionType = "KubeConfigCertExpiresInSevenDays"
)

type ClusterCondition struct {
	Type ClusterConditionType `json:"type"`

	Status v1.ConditionStatus `json:"status"`

	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	Reason string `json:"reason,omitempty"`

	Message string `json:"message,omitempty"`
}

type ClusterStatus struct {
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	HorizonVersion string `json:"horizonVersion,omitempty"`

	NodeCount int `json:"nodeCount,omitempty"`

	// +optional
	Zones []string `json:"zones,omitempty"`

	// +optional
	Region *string `json:"region,omitempty"`

	// +optional
	Configz map[string]bool `json:"configz,omitempty"`

	UID types.UID `json:"uid,omitempty"`
}

// +kubebuilder:object:root=true
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}
