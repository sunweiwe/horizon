package informers

import (
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	fakeclientset "github.com/sunweiwe/horizon/pkg/client/clientset/fake"
	horizoninformers "github.com/sunweiwe/horizon/pkg/client/informers"
)

type nullInformerFactory struct {
	fakeK8sInformerFactory     informers.SharedInformerFactory
	fakeHorizonInformerFactory horizoninformers.SharedInformerFactory
}

func NewNullInformerFactory() InformerFactory {
	fakeClient := fake.NewSimpleClientset()
	fakeInformerFactory := informers.NewSharedInformerFactory(fakeClient, time.Minute*10)

	fakeHorizonClient := fakeclientset.NewSimpleClientset()
	fakeHorizonInformerFactory := horizoninformers.NewSharedInformerFactory(fakeHorizonClient, time.Minute*10)

	return &nullInformerFactory{
		fakeK8sInformerFactory:     fakeInformerFactory,
		fakeHorizonInformerFactory: fakeHorizonInformerFactory,
	}
}

func (n nullInformerFactory) KubernetesSharedInformerFactory() informers.SharedInformerFactory {
	return n.fakeK8sInformerFactory
}

func (n nullInformerFactory) HorizonSharedInformerFactory() horizoninformers.SharedInformerFactory {
	return n.fakeHorizonInformerFactory
}

func (n nullInformerFactory) Start(stopCh <-chan struct{}) {
}
