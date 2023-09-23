package v1alpha2

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"github.com/sunweiwe/horizon/pkg/client/clientset"
	"github.com/sunweiwe/horizon/pkg/informers"
	"github.com/sunweiwe/horizon/pkg/models/tenant"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type tenantHandler struct {
	tenant tenant.Interface
}

func NewTenantHandler(factory informers.InformerFactory, client kubernetes.Interface, horizon clientset.Interface) *tenantHandler {

	return &tenantHandler{
		tenant: tenant.New(factory, client, horizon),
	}
}

func (h *tenantHandler) ListClusters(r *restful.Request, response *restful.Response) {
	user, ok := request.UserFrom(r.Request.Context())

	if !ok {
		response.WriteEntity([]interface{}{})
		return
	}

	queryParam := query.ParseQueryParameter(r)
	data, err := h.tenant.ListClusters(user, queryParam)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, r, err)
			return
		}
		api.HandleInternalError(response, r, err)
		return
	}

	response.WriteEntity(data)
}
