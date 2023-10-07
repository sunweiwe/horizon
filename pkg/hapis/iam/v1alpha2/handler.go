package v1alpha2

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/authorization/authorizer"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"github.com/sunweiwe/horizon/pkg/models/iam/im"

	iamv1alpha2 "github.com/sunweiwe/api/iam/v1alpha2"
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
	queryParam := query.ParseQueryParameter(request)
	data, err := h.im.ListUsers(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	for i, item := range data.Items {
		user := item.(*iamv1alpha2.User)
		user = user.DeepCopy()

		data.Items[i] = user
	}

	u := iamv1alpha2.User{}
	u.Name = "sun"
	data.Items = append(data.Items, u)

	response.WriteEntity(nil)
}
