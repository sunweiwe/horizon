package cluster

import (
	"github.com/sunweiwe/horizon/pkg/client/informers/cluster/v1alpha1"
	"github.com/sunweiwe/horizon/pkg/client/informers/internal"
)

type Interface interface {
	V1alpha1() v1alpha1.Interface
}

type group struct {
	namespace        string
	factory          internal.SharedInformerFactory
	tweakListOptions internal.TweakListOptionsFunc
}

func New(s internal.SharedInformerFactory, ns string, tweakListOptions internal.TweakListOptionsFunc) Interface {
	return &group{
		namespace:        ns,
		factory:          s,
		tweakListOptions: tweakListOptions,
	}
}

func (g *group) V1alpha1() v1alpha1.Interface {
	return v1alpha1.New(g.factory, g.namespace, g.tweakListOptions)
}
