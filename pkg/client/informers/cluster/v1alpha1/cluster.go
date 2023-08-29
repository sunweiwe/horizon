package v1alpha1

import (
	"context"
	"time"

	"github.com/sunweiwe/horizon/pkg/client/clientset"
	"github.com/sunweiwe/horizon/pkg/client/informers/internal"
	"github.com/sunweiwe/horizon/pkg/client/listers/cluster/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	clusterv1alpha1 "github.com/sunweiwe/api/cluster/v1alpha1"
)

type ClusterInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ClusterLister
}

type clusterInformer struct {
	factory          internal.SharedInformerFactory
	tweakListOptions internal.TweakListOptionsFunc
}

func (c *clusterInformer) Informer() cache.SharedIndexInformer {
	return c.factory.InformerFor(&clusterv1alpha1.Cluster{}, c.defaultInformer)
}

func (c *clusterInformer) Lister() v1alpha1.ClusterLister {
	return v1alpha1.NewClusterLister(c.Informer().GetIndexer())
}

func (f *clusterInformer) defaultInformer(client clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func NewFilteredClusterInformer(client clientset.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internal.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ClusterV1alpha1().Clusters().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ClusterV1alpha1().Clusters().Watch(context.TODO(), options)
			},
		},
		&clusterv1alpha1.Cluster{},
		resyncPeriod,
		indexers,
	)
}
