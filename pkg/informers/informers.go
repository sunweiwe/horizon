package informers

import (
	"reflect"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/sunweiwe/horizon/pkg/client/clientset"
	horizonInformers "github.com/sunweiwe/horizon/pkg/client/informers"
	kubeInformers "k8s.io/client-go/informers"
)

const defaultResync = 600 * time.Second

type InformerFactory interface {
	KubernetesSharedInformerFactory() kubeInformers.SharedInformerFactory
	HorizonSharedInformerFactory() horizonInformers.SharedInformerFactory

	Start(stopCh <-chan struct{})
}

type GenericInformerFactory interface {
	Start(stopCh <-chan struct{})
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool
}

type informerFactories struct {
	informerFactory        kubeInformers.SharedInformerFactory
	horizonInformerFactory horizonInformers.SharedInformerFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}

	if f.horizonInformerFactory != nil {
		f.horizonInformerFactory.Start(stopCh)
	}
}

func NewInformerFactories(client kubernetes.Interface, horizonClient clientset.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = kubeInformers.NewSharedInformerFactory(client, defaultResync)
	}

	if horizonClient != nil {
		factory.horizonInformerFactory = horizonInformers.NewSharedInformerFactory(horizonClient, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() kubeInformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) HorizonSharedInformerFactory() horizonInformers.SharedInformerFactory {
	return f.horizonInformerFactory
}
