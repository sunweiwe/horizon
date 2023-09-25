package v1alpha2

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/apiserver/authorization/authorizer"
	"github.com/sunweiwe/horizon/pkg/models/iam/im"
)

type iamHandler struct {
	im         im.IdentityManagementInterface
	authorizer authorizer.Authorizer
}

func newHandler(im im.IdentityManagementInterface, authorizer authorizer.Authorizer) *iamHandler {
	return &iamHandler{
		im:         im,
		authorizer: authorizer,
	}
}

func (h *iamHandler) ListUsers(request *restful.Request, response *restful.Response) {

}
