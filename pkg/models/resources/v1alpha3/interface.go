package v1alpha3

import (
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"k8s.io/apimachinery/pkg/runtime"
)

type Interface interface {
	Get(namespace, name string) (runtime.Object, error)

	List(namespace string, query *query.Query) (*api.ListResult, error)
}
