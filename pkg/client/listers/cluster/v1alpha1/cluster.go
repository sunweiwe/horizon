package v1alpha1

import (
	"github.com/sunweiwe/api/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type ClusterLister interface {
	List(selector labels.Selector) (ret []*v1alpha1.Cluster, err error)

	Get(name string) (*v1alpha1.Cluster, error)
}

type clusterLister struct {
	indexer cache.Indexer
}

func NewClusterLister(indexer cache.Indexer) ClusterLister {
	return &clusterLister{indexer: indexer}
}

func (c *clusterLister) List(selector labels.Selector) (ret []*v1alpha1.Cluster, err error) {
	err = cache.ListAll(c.indexer, selector, func(i interface{}) {
		ret = append(ret, i.(*v1alpha1.Cluster))
	})

	return ret, err
}

func (c *clusterLister) Get(name string) (*v1alpha1.Cluster, error) {
	obj, exists, err := c.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("cluster"), name)
	}

	return obj.(*v1alpha1.Cluster), nil
}
