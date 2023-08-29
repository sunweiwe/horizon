package fake

import (
	"github.com/sunweiwe/horizon/pkg/client/clientset/typed/cluster/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

type FakeClusterV1alpha1 struct {
	*testing.Fake
}

func (c *FakeClusterV1alpha1) Clusters() v1alpha1.ClusterInterface {
	return &FakeClusters{c}
}

func (c *FakeClusterV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
