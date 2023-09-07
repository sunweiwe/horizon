package k8s

import (
	"github.com/sunweiwe/horizon/pkg/client/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type nullClient struct {
}

func NewNullClient() Client {
	return &nullClient{}
}

func (n nullClient) Kubernetes() kubernetes.Interface {
	return nil
}

func (n nullClient) Horizon() clientset.Interface {
	return nil
}

func (n nullClient) Config() *rest.Config {
	return nil
}
