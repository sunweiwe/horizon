package am

import (
	"github.com/sunweiwe/horizon/pkg/client/clientset"
	"github.com/sunweiwe/horizon/pkg/informers"
	"k8s.io/client-go/kubernetes"
)

type AccessManagementInterface interface {
}

func NewOperator(kube kubernetes.Interface, horizon clientset.Interface, factory informers.InformerFactory) AccessManagementInterface {
	amOperator := NewReadOnlyOperator(factory).(*amOperator)
	amOperator.kube = kube
	amOperator.horizon = horizon

	return amOperator
}

type amOperator struct {
	kube    kubernetes.Interface
	horizon clientset.Interface
}

func NewReadOnlyOperator(factory informers.InformerFactory) AccessManagementInterface {
	operator := &amOperator{}

	return operator
}
