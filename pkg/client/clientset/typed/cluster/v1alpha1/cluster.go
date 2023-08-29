package v1alpha1

import (
	"context"
	"time"

	"github.com/sunweiwe/api/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type ClustersGetter interface {
	Clusters() ClusterInterface
}

type ClusterInterface interface {
	Update(ctx context.Context, cluster *v1alpha1.Cluster, opts v1.UpdateOptions) (*v1alpha1.Cluster, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ClusterList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
}

type clusters struct {
	client rest.Interface
}

func newClusters(c *ClusterV1alpha1Client) *clusters {
	return &clusters{
		client: c.RESTClient(),
	}
}

func (c *clusters) Update(ctx context.Context, cluster *v1alpha1.Cluster, opts v1.UpdateOptions) (ret *v1alpha1.Cluster, err error) {
	ret = &v1alpha1.Cluster{}
	err = c.client.Put().
		Resource("clusters").
		Name(cluster.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(cluster).
		Do(ctx).
		Into(ret)

	return
}

func (c *clusters) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ClusterList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	ret := &v1alpha1.ClusterList{}
	err = c.client.Get().
		Resource("clusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(ret)

	return
}

func (c *clusters) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("clusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}
