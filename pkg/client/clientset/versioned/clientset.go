package versioned

import (
	"fmt"

	clusterv1alpha1 "github.com/sunweiwe/horizon/pkg/client/clientset/versioned/typed/cluster/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	ClusterV1alpha1() clusterv1alpha1.ClusterV1alpha1Interface
}

type Clientset struct {
	clusterv1alpha1 *clusterv1alpha1.ClusterV1alpha1Client
}

func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.clusterv1alpha1 = &clusterv1alpha1.ClusterV1alpha1Client{}

	return &cs
}

func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, c.Burst)
	}

	var cs Clientset
	var err error

	cs.clusterv1alpha1, err = clusterv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func (c *Clientset) ClusterV1alpha1() clusterv1alpha1.ClusterV1alpha1Interface {
	return c.clusterv1alpha1
}
