package filter

import (
	"fmt"
	"io"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/request"
	"github.com/sunweiwe/horizon/pkg/models/resources/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var NotSupportedVerbError = fmt.Errorf("Not support verb")

type DynamicResourceHandler struct {
	v1beta1.ResourceManager
	serviceErrorHandleFallback restful.ServiceErrorHandleFunction
}

func NewDynamicResourceHandler(
	serviceErrorHandleFallback restful.ServiceErrorHandleFunction,
	resourceGetter v1beta1.ResourceManager) *DynamicResourceHandler {
	return &DynamicResourceHandler{
		ResourceManager:            resourceGetter,
		serviceErrorHandleFallback: serviceErrorHandleFallback,
	}
}

func (d *DynamicResourceHandler) HandleServiceError(serviceError restful.ServiceError, req *restful.Request, w *restful.Response) {
	if serviceError.Code != http.StatusNotFound {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	requestInfo, exist := request.RequestInfoFrom(req.Request.Context())
	if !exist {
		responsewriters.InternalError(w, req.Request, fmt.Errorf("No RequestInfo found in the context!"))
		return
	}

	if requestInfo.KubernetesRequest {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    requestInfo.APIGroup,
		Version:  requestInfo.APIVersion,
		Resource: requestInfo.Resource,
	}

	if gvr.Group == "" || gvr.Resource == "" || gvr.Version == "" {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	served, err := d.IsServed(gvr)
	if err != nil {
		responsewriters.InternalError(w, req.Request, err)
		return
	}

	if !served {
		d.serviceErrorHandleFallback(serviceError, req, w)
		return
	}

	var object client.Object
	if requestInfo.Verb == request.VerbCreate || requestInfo.Verb == request.VerbUpdate || requestInfo.Verb == request.VerbPatch {
		rawData, err := io.ReadAll(req.Request.Body)
		if err != nil {
			api.HandleError(w, req, err)
			return
		}

		object, err = d.CreateObjectFromRawData(gvr, rawData)
		if err != nil {
			api.HandleError(w, req, err)
			return
		}
	}

	var result interface{}
	switch requestInfo.Verb {
	case request.VerbGet:
		result, err = d.GetResource(req.Request.Context(), gvr, requestInfo.Namespace, requestInfo.Name)
	case request.VerbCreate:
		err = d.CreateResource(req.Request.Context(), object)
	default:
		err = NotSupportedVerbError
	}

	if err != nil {
		if meta.IsNoMatchError(err) {
			d.serviceErrorHandleFallback(serviceError, req, w)
			return
		}
		api.HandleError(w, req, err)
		return
	}

	w.WriteAsJson(result)
}
