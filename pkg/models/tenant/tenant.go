package tenant

import (
	"fmt"

	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/authorization/authorizer"
	"github.com/sunweiwe/horizon/pkg/apiserver/query"
	"github.com/sunweiwe/horizon/pkg/apiserver/request"
	"github.com/sunweiwe/horizon/pkg/client/clientset"
	"github.com/sunweiwe/horizon/pkg/informers"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"

	clusterv1alpha1 "github.com/sunweiwe/api/cluster/v1alpha1"
	resourcesv1alpha3 "github.com/sunweiwe/horizon/pkg/models/resources/v1alpha3/resource"
)

type Interface interface {
	ListClusters(info user.Info, params *query.Query) (*api.ListResult, error)
}

type tenantOperator struct {
	kube           kubernetes.Interface
	horizon        clientset.Interface
	authorizer     authorizer.Authorizer
	resourceGetter *resourcesv1alpha3.ResourceGetter
}

func New(informers informers.InformerFactory, client kubernetes.Interface, horizon clientset.Interface) Interface {

	return &tenantOperator{
		kube:    client,
		horizon: horizon,
	}
}

func (t *tenantOperator) ListClusters(info user.Info, params *query.Query) (*api.ListResult, error) {

	listClusters := authorizer.AtrributesRecord{
		User:            info,
		Verb:            "list",
		APIGroup:        "cluster.horizon.io",
		Resource:        "clusters",
		ResourceScope:   request.GlobalScope,
		ResourceRequest: true,
	}

	allowedListClusters, _, err := t.authorizer.Authorize(listClusters)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %s", err)
	}

	if allowedListClusters == authorizer.DecisionAllow {
		return t.resourceGetter.List(clusterv1alpha1.ResourcesPluralCluster, "", params)
	}

	return &api.ListResult{}, nil
}
