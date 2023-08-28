package informers

import (
	"reflect"
	"sync"
	"time"

	"github.com/sunweiwe/horizon/pkg/client/clientset/versioned"
	"github.com/sunweiwe/horizon/pkg/client/informers/cluster"
	"github.com/sunweiwe/horizon/pkg/client/informers/internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

type SharedInformerFactory interface {
	internal.SharedInformerFactory
	Cluster() cluster.Interface
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool

	ForResource(resource schema.GroupVersionResource) (GenericInformer, error)
}

type SharedInformerOption func(*sharedInformerFactory) *sharedInformerFactory

type sharedInformerFactory struct {
	client versioned.Interface

	namespace        string
	tweakListOptions internal.TweakListOptionsFunc

	lock sync.Mutex

	customResync  map[reflect.Type]time.Duration
	defaultResync time.Duration

	informers map[reflect.Type]cache.SharedIndexInformer

	startedInformers map[reflect.Type]bool
}

func (s *sharedInformerFactory) Cluster() cluster.Interface {
	return cluster.New(s, s.namespace, s.tweakListOptions)
}

func (s *sharedInformerFactory) InformerFor(obj runtime.Object, handler internal.NewInformerFunc) cache.SharedIndexInformer {
	informerType := reflect.TypeOf(obj)
	informer, exists := s.informers[informerType]
	if exists {
		return informer
	}

	resyncPeriod, exists := s.customResync[informerType]
	if !exists {
		resyncPeriod = s.defaultResync
	}

	informer = handler(s.client, resyncPeriod)
	s.informers[informerType] = informer

	return informer
}

func NewSharedInformerFactory(client versioned.Interface, defaultResync time.Duration) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync)
}

func (s *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for informerType, informer := range s.informers {
		if !s.startedInformers[informerType] {
			go informer.Run(stopCh)
			s.startedInformers[informerType] = true
		}
	}
}

func NewSharedInformerFactoryWithOptions(client versioned.Interface, defaultResync time.Duration, options ...SharedInformerOption) SharedInformerFactory {
	f := &sharedInformerFactory{
		client:           client,
		namespace:        v1.NamespaceAll,
		defaultResync:    defaultResync,
		informers:        make(map[reflect.Type]cache.SharedIndexInformer),
		startedInformers: make(map[reflect.Type]bool),
		customResync:     make(map[reflect.Type]time.Duration),
	}

	for _, opt := range options {
		f = opt(f)
	}

	return f
}

func (f *sharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	informers := func() map[reflect.Type]cache.SharedIndexInformer {
		f.lock.Lock()
		defer f.lock.Unlock()

		informers := map[reflect.Type]cache.SharedIndexInformer{}
		for informerType, informer := range informers {
			if f.startedInformers[informerType] {
				informers[informerType] = informer
			}
		}
		return informers
	}()

	ret := map[reflect.Type]bool{}
	for informerType, informer := range informers {
		ret[informerType] = cache.WaitForCacheSync(stopCh, informer.HasSynced)
	}
	return ret
}
