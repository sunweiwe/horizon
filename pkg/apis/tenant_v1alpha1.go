package apis

import (
	"github.com/sunweiwe/api/tenant/v1alpha1"
)

func init() {
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
