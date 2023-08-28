package informers

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	clusterv1alpha1 "github.com/sunweiwe/api/cluster/v1alpha1"
)

type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {

	switch resource {

	case clusterv1alpha1.SchemeGroupVersion.WithResource("clusters"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cluster().V1alpha1().Clusters().Informer()}, nil
	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
