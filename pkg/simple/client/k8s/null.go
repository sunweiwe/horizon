package k8s

import (
	horizon "github.com/sunweiwe/horizon/pkg/client/clientset/versioned"
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

func (n nullClient) Horizon() horizon.Interface {
	return nil
}

func (n nullClient) Config() *rest.Config {
	return nil
}
