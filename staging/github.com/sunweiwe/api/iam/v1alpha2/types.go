package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// User is the Schema for the users API
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +kubebuilder:object:root=true
type User struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec"`
	// +optional
	Status UserState `json:"status,omitempty"`
}

type UserSpec struct {
	Email string `json:"email"`

	// +optional
	Lang string `json:"lang,omitempty"`

	// +optional
	Description string `json:"description,omitempty"`

	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// +optional
	Groups []string `json:"groups,omitempty"`

	EncryptedPassword string `json:"password,omitempty"`
}

type UserState string

type UserStatus struct {
	// +optional
	State UserState `json:"state,omitempty"`

	// +optional
	Reason string `json:"reason,omitempty"`

	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// +optional
	LastLoginTime *metav1.Time `json:"lastLoginTime,omitempty"`
}

// UserList contains a list of User
// +kubebuilder:object:root=true
type UserList struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}
