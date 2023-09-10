package installer

import (
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"

	tenantv1alpha1 "github.com/sunweiwe/api/tenant/v1alpha1"
)

func Installer(scheme *k8sruntime.Scheme) {
	urlruntime.Must(tenantv1alpha1.AddToScheme(scheme))
	urlruntime.Must(scheme.SetVersionPriority(tenantv1alpha1.SchemeGroupVersion))
}
