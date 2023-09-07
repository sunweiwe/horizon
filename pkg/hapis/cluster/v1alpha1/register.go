package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/runtime"
	"github.com/sunweiwe/horizon/pkg/constants"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/informers"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	horizon "github.com/sunweiwe/horizon/pkg/client/clientset"
	horizonInformers "github.com/sunweiwe/horizon/pkg/client/informers/externalversions"
	clusterlister "github.com/sunweiwe/horizon/pkg/client/listers/cluster/v1alpha1"
	v1 "k8s.io/client-go/listers/core/v1"
)

const (
	GroupName = "cluster.horizon.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func AddToContainer(
	container *restful.Container,
	horizonClient horizon.Interface,
	informers informers.SharedInformerFactory,
	hzInformers horizonInformers.SharedInformerFactory,
	proxyService string,
	proxyAddress string,
	agentImage string,
) error {

	webService := runtime.NewWebService(GroupVersion)
	h := newHandler(horizonClient, informers, hzInformers, proxyService, proxyAddress, agentImage)

	webService.Route(webService.GET("/clusters/{cluster}/agent/deployment").
		Doc("Return deployment yaml for cluster agent").
		Param(webService.PathParameter("cluster", "Name of the cluster.").Required(true)).
		To(h.generateAgentDeployment).
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MultiClusterTag}))

	container.Add(webService)

	return nil
}

type handler struct {
	horizonClient   horizon.Interface
	serviceLister   v1.ServiceLister
	configMapLister v1.ConfigMapLister
	clusterLister   clusterlister.ClusterLister

	proxyService string
	proxyAddress string
	agentImage   string
	yamlPrinter  *printers.YAMLPrinter
}

func newHandler(horizonClient horizon.Interface, informers informers.SharedInformerFactory, hzInformers horizonInformers.SharedInformerFactory,
	proxyService, proxyAddress, agentImage string) *handler {

	return &handler{
		horizonClient:   horizonClient,
		serviceLister:   informers.Core().V1().Services().Lister(),
		configMapLister: informers.Core().V1().ConfigMaps().Lister(),
		clusterLister:   hzInformers.Cluster().V1alpha1().Clusters().Lister(),

		proxyService: proxyService,
		proxyAddress: proxyAddress,
		agentImage:   agentImage,
		yamlPrinter:  &printers.YAMLPrinter{},
	}
}
