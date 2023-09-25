package v1alpha2

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/apiserver/authorization/authorizer"
	"github.com/sunweiwe/horizon/pkg/models/iam/im"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName = "iam.horizon.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, authorizer authorizer.Authorizer, im im.IdentityManagementInterface) error {

	return nil
}
