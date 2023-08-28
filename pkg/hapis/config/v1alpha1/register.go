package v1alpha1

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/apiserver/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	horizonConfig "github.com/sunweiwe/horizon/pkg/apiserver/config"
)

const (
	GroupName = "config.horizon.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func AddToContainer(c *restful.Container, config horizonConfig.Config) error {
	service := runtime.NewWebService(GroupVersion)

	service.Route(service.GET("/configs/oauth").
		Doc("Information about the authorization server are published.").
		To(func(request *restful.Request, response *restful.Response) {
			response.WriteEntity(config.KubernetesOptions)
		}))

	return nil
}
