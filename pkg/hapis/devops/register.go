package devops

import (
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/hapis/generic"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var devopsGroupVersions = []schema.GroupVersion{
	{Group: "devops.horizon.io", Version: "v1alpha1"},
	{Group: "gitops.horizon.io", Version: "v1alpha1"},
}

func AddToContainer(container *restful.Container, endpoint string) error {
	endpoint = strings.TrimSuffix(endpoint, "/")
	for _, groupVersion := range devopsGroupVersions {
		proxy, err := generic.NewGenericProxy(endpoint+"/hapis/"+groupVersion.Group, groupVersion.Group, groupVersion.Version)
		if err != nil {
			return err
		}
		if err = proxy.AddToContainer(container); err != nil {
			return err
		}
	}

	return nil
}
