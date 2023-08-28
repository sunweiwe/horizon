package runtime

import (
	"strings"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ApiRootPath = "/hapis"
)

var Container = restful.NewContainer()

const MimeMergePatchJson = "application/merge-patch+json"
const MimeJsonPatchJson = "application/json-patch+json"
const MimeMultipartFormData = "multipart/form-data"

func init() {
	restful.RegisterEntityAccessor(MimeMergePatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
	restful.RegisterEntityAccessor(MimeJsonPatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
}

func NewWebService(gv schema.GroupVersion) *restful.WebService {
	webService := restful.WebService{}

	webService.Path(strings.TrimRight(ApiRootPath+"/"+gv.String(), "/")).Produces(restful.MIME_JSON)
	return &webService
}
