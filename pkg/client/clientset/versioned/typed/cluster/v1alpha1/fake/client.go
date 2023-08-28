package fake

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	v1alpha1 "github.com/sunweiwe/horizon/pkg/client/clientset/versioned/typed/cluster/v1alpha1"
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
