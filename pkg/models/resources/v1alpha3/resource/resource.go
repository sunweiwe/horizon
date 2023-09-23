package resource

import (
	"errors"

	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"github.com/sunweiwe/horizon/pkg/models/resources/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type ResourceGetter struct {
	clusterResourceGetters    map[schema.GroupVersionResource]v1alpha3.Interface
	namespacedResourceGetters map[schema.GroupVersionResource]v1alpha3.Interface
}

func (r *ResourceGetter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {
	cluster := namespace == ""
	getter := r.TryResource(cluster, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.List(namespace, query)
}

func (r *ResourceGetter) TryResource(clusterScope bool, resource string) v1alpha3.Interface {
	if clusterScope {
		for k, v := range r.clusterResourceGetters {
			if k.Resource == resource {
				return v
			}
		}
	}
	for k, v := range r.namespacedResourceGetters {
		if k.Resource == resource {
			return v
		}
	}

	return nil
}
