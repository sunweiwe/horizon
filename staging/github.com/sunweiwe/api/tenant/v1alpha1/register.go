// Package v1alpha1 contains API Schema definitions for the tenant v1alpha1 API group
// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true
// +k8s:conversion-gen=horizon.io/api/tenant
// +k8s:defaulter-gen=TypeMeta
// +groupName=tenant.horizon.io
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "tenant.horizon.io", Version: "v1alpha1"}

	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
