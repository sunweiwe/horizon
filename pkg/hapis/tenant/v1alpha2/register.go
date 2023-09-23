package v1alpha2

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/runtime"
	"github.com/sunweiwe/horizon/pkg/client/clientset"
	"github.com/sunweiwe/horizon/pkg/constants"
	"github.com/sunweiwe/horizon/pkg/informers"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

const (
	GroupName = "tenant.horizon.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, factory informers.InformerFactory, client kubernetes.Interface, horizon clientset.Interface) error {
	service := runtime.NewWebService(GroupVersion)
	handler := NewTenantHandler(factory, client, horizon)

	service.Route(service.GET("/clusters").
		To(handler.ListClusters).
		Doc("List clusters available to users").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserResourceTag}))

	c.Add(service)
	return nil
}
