package fake

import (
	"context"

	"github.com/sunweiwe/api/cluster/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/testing"
)

type FakeClusters struct {
	Fake *FakeClusterV1alpha1
}

var clustersResource = schema.GroupVersionResource{Group: "cluster.horizon.io", Version: "v1alpha1", Resource: "clusters"}

func (c *FakeClusters) Update(ctx context.Context, cluster *v1alpha1.Cluster, opts v1.UpdateOptions) (result *v1alpha1.Cluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(clustersResource, cluster), &v1alpha1.Cluster{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cluster), err
}
