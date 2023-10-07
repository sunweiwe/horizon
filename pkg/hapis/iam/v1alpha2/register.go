package v1alpha2

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	iamv1alpha2 "github.com/sunweiwe/api/iam/v1alpha2"
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/authorization/authorizer"
	"github.com/sunweiwe/horizon/pkg/apiserver/runtime"
	"github.com/sunweiwe/horizon/pkg/constants"
	"github.com/sunweiwe/horizon/pkg/models/iam/im"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName = "iam.horizon.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, authorizer authorizer.Authorizer, im im.IdentityManagementInterface) error {
	service := runtime.NewWebService(GroupVersion)
	handler := newHandler(im, authorizer)

	// user
	service.Route(service.GET("/users").
		To(handler.ListUsers).
		Doc("List all users.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	return nil
}
