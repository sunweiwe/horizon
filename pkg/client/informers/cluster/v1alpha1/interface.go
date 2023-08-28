package v1alpha1

import (
	"github.com/sunweiwe/horizon/pkg/client/informers/internal"
)

type Interface interface {
	Clusters() ClusterInformer
}

type version struct {
	factory          internal.SharedInformerFactory
	namespace        string
	tweakListOptions internal.TweakListOptionsFunc
}

func New(f internal.SharedInformerFactory, ns string, tweakListOptions internal.TweakListOptionsFunc) Interface {
	return &version{
		factory:          f,
		namespace:        ns,
		tweakListOptions: tweakListOptions,
	}
}

func (v *version) Clusters() ClusterInformer {
	return &clusterInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}
