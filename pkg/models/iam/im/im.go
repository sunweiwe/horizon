package im

import (
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"k8s.io/klog"

	resources "github.com/sunweiwe/horizon/pkg/models/resources/v1alpha3"
)

type IdentityManagementInterface interface {
	ListUsers(query *query.Query) (*api.ListResult, error)
}

type imOperator struct {
	userGetter resources.Interface
}

func NewOperator(userGetter resources.Interface) IdentityManagementInterface {
	return &imOperator{
		userGetter: userGetter,
	}
}

func (im *imOperator) ListUsers(query *query.Query) (ret *api.ListResult, err error) {
	ret, err = im.userGetter.List("", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return
}
