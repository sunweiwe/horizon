package im

import (
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"k8s.io/klog"

	iamv1alpha2 "github.com/sunweiwe/api/iam/v1alpha2"
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

	data := make([]interface{}, 0)
	for _, item := range ret.Items {
		user := item.(*iamv1alpha2.User)
		out := ensurePasswordNotOutput(user)
		data = append(data, out)
	}

	ret.Items = data
	return ret, nil
}

func ensurePasswordNotOutput(user *iamv1alpha2.User) *iamv1alpha2.User {
	out := user.DeepCopy()
	out.Spec.EncryptedPassword = ""
	return out
}
