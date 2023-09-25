package user

import (
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"github.com/sunweiwe/horizon/pkg/client/informers/externalversions"
	"github.com/sunweiwe/horizon/pkg/models/resources/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type usersGetter struct {
	kubeInformer     informers.SharedInformerFactory
	horizonInformers externalversions.SharedInformerFactory
}

func New(kube informers.SharedInformerFactory, horizon externalversions.SharedInformerFactory) v1alpha3.Interface {
	return &usersGetter{kubeInformer: kube, horizonInformers: horizon}
}

func (u *usersGetter) Get(_, name string) (runtime.Object, error) {
	return u.horizonInformers.Iam().V1alpha2().Users().Lister().Get(name)
}

// TODO
func (u *usersGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	return &api.ListResult{}, nil
}
