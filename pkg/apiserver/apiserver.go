package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/apiserver/filter"
	"github.com/sunweiwe/horizon/pkg/apiserver/request"
	"github.com/sunweiwe/horizon/pkg/informers"
	"github.com/sunweiwe/horizon/pkg/models/resources/v1beta1"
	"github.com/sunweiwe/horizon/pkg/server/healthz"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"github.com/sunweiwe/horizon/pkg/simple/client/monitoring"
	"github.com/sunweiwe/horizon/pkg/utils/ip"
	"github.com/sunweiwe/horizon/pkg/utils/metrics"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	apiserverconfig "github.com/sunweiwe/horizon/pkg/apiserver/config"
	clusterv1alphal "github.com/sunweiwe/horizon/pkg/hapis/cluster/v1alpha1"
	tenantv1alpha2 "github.com/sunweiwe/horizon/pkg/hapis/tenant/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var initMetrics sync.Once

type APIServer struct {
	ServerCount int

	Server *http.Server

	Config *apiserverconfig.Config

	container *restful.Container

	KubernetesClient k8s.Client

	InformerFactory informers.InformerFactory

	RuntimeCache runtimecache.Cache

	RuntimeClient runtimeclient.Client

	MetricsClient monitoring.Interface

	MonitoringClient monitoring.Interface
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	klog.V(0).Info("Apiserver PrepareRun")
	s.container = restful.NewContainer()
	s.container.Filter(logRequest)
	s.container.Router(restful.CurlyRouter{})

	s.container.RecoverHandler(func(i interface{}, w http.ResponseWriter) {
		logStackFromRecover(i, w)
	})

	s.dynamicResourceAPI()
	s.horizonAPIs(stopCh)
	s.metricsAPI()

	urlruntime.Must(healthz.Handler(s.container, []healthz.HealthChecker{}...))

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container
	s.buildHandlerChain(stopCh)

	return nil
}

func (s *APIServer) buildHandlerChain(stopCh <-chan struct{}) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.New("api", "apis", "hapis", "hapi"),
		GroupLessAPIPrefixes: sets.New("api", "hapi"),
		GlobalResources:      []schema.GroupResource{},
	}

	handler := s.Server.Handler
	handler = filter.WithKubeAPIServer(handler, s.KubernetesClient.Config())

	handler = filter.WithRequestInfo(handler, requestInfoResolver)
	s.Server.Handler = handler
}

func (s *APIServer) Run(ctx context.Context) (err error) {
	klog.V(0).Info("Apiserver Run")

	err = s.waitForResourceSync(ctx)
	if err != nil {
		return err
	}

	shutdown, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = s.Server.Shutdown(shutdown)
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}

func (s *APIServer) waitForResourceSync(ctx context.Context) error {
	stopCh := ctx.Done()

	k8sGVRs := map[schema.GroupVersion][]string{
		{Group: "", Version: "v1"}: {
			"namespaces",
			"nodes",
			"resourcequotas",
			"pods",
			"services",
			"persistentvolumeclaims",
			"persistentvolumes",
			"secrets",
			"configmaps",
			"serviceaccounts",
		},
		{Group: "rbac.authorization.k8s.io", Version: "v1"}: {
			"roles",
			"rolebindings",
			"clusterroles",
			"clusterrolebindings",
		},
		{Group: "apps", Version: "v1"}: {
			"deployments",
			"daemonsets",
			"replicasets",
			"statefulsets",
			"controllerrevisions",
		},
		{Group: "storage.k8s.io", Version: "v1"}: {
			"storageclasses",
		},
		{Group: "batch", Version: "v1"}: {
			"jobs",
			"cronjobs",
		},
		{Group: "networking.k8s.io", Version: "v1"}: {
			"ingresses",
			"networkpolicies",
		},
		{Group: "autoscaling", Version: "v2"}: {
			"horizontalpodautoscalers",
		},
	}

	if err := waitForCacheSync(
		s.KubernetesClient.Kubernetes().Discovery(),
		s.InformerFactory.KubernetesSharedInformerFactory(),
		func(resource schema.GroupVersionResource) (interface{}, error) {
			return s.InformerFactory.KubernetesSharedInformerFactory().ForResource(resource)
		},
		k8sGVRs,
		stopCh,
	); err != nil {
		return err
	}

	hzGVRs := map[schema.GroupVersion][]string{
		{Group: "cluster.horizon.io", Version: "v1alpha1"}: {"clusters"},
		{Group: "tenant.horiozn.io", Version: "v1alpha1"}:  {"workspaces"},
	}

	if err := waitForCacheSync(
		s.KubernetesClient.Kubernetes().Discovery(),
		s.InformerFactory.HorizonSharedInformerFactory(),
		func(resource schema.GroupVersionResource) (interface{}, error) {
			return s.InformerFactory.HorizonSharedInformerFactory().ForResource(resource)
		},
		hzGVRs, stopCh,
	); err != nil {
		return err
	}

	//
	go s.RuntimeCache.Start(ctx)
	s.RuntimeCache.WaitForCacheSync(ctx)

	klog.V(0).Info("Finished caching objects")
	return nil
}

type informerForResourceFunc func(resource schema.GroupVersionResource) (interface{}, error)

func waitForCacheSync(discoveryClient discovery.DiscoveryInterface,
	sharedInformerFactory informers.GenericInformerFactory,
	informerForResourceFunc informerForResourceFunc, GVRs map[schema.GroupVersion][]string, stopCh <-chan struct{}) error {
	klog.V(0).Info("Apiserver Start cache objects")

	for groupVersion, resourceNames := range GVRs {
		var apiResourceList *v1.APIResourceList
		var err error
		err = retry.OnError(retry.DefaultRetry, func(err error) bool {
			return !errors.IsNotFound(err)
		}, func() error {
			apiResourceList, err = discoveryClient.ServerResourcesForGroupVersion(groupVersion.String())
			return err
		})
		if err != nil {
			if errors.IsNotFound(err) {
				klog.Warningf("group version %s not exists in the cluster", groupVersion)
				continue
			}
			return fmt.Errorf("failed to fetch group version %s: %s", groupVersion, err)
		}

		for _, resourceName := range resourceNames {
			groupVersionResource := groupVersion.WithResource(resourceName)
			if !isResourceExists(apiResourceList.APIResources, groupVersionResource) {
				klog.Warningf("resource %s not exists in the cluster", groupVersionResource)
			} else {
				if _, err = informerForResourceFunc(groupVersionResource); err != nil {
					return fmt.Errorf("failed to create informer for %s: %s", groupVersionResource, err)
				}
			}
		}
	}

	sharedInformerFactory.Start(stopCh)
	sharedInformerFactory.WaitForCacheSync(stopCh)
	klog.V(0).Info("Apiserver WaitForCacaheSync Start successful")
	return nil
}

func isResourceExists(apiResources []v1.APIResource, resource schema.GroupVersionResource) bool {
	for _, apiResource := range apiResources {
		if apiResource.Name == resource.Resource {
			return true
		}
	}

	return false
}

func logStackFromRecover(reason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", reason))
	for i := 2; ; i += 1 {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content_Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func (s *APIServer) dynamicResourceAPI() {
	dynamicResourceHandler := filter.NewDynamicResourceHandler(
		func(err restful.ServiceError, req *restful.Request, resp *restful.Response) {
			for header, values := range err.Header {
				for _, value := range values {
					resp.Header().Add(header, value)
				}
			}
			resp.WriteErrorString(err.Code, err.Message)
		}, v1beta1.New(s.RuntimeClient, s.RuntimeCache))

	s.container.ServiceErrorHandler(dynamicResourceHandler.HandleServiceError)
}

func (s *APIServer) metricsAPI() {
	initMetrics.Do(registerMetrics)
	metrics.Defaults.Install(s.container)
}

func (s *APIServer) horizonAPIs(stopCh <-chan struct{}) {
	urlruntime.Must(clusterv1alphal.AddToContainer(
		s.container,
		s.KubernetesClient.Horizon(),
		s.InformerFactory.KubernetesSharedInformerFactory(),
		s.InformerFactory.HorizonSharedInformerFactory(),
		s.Config.MultiClusterOptions.ProxyPublishService,
		s.Config.MultiClusterOptions.ProxyPublishAddress,
		s.Config.MultiClusterOptions.AgentImage))

	urlruntime.Must(tenantv1alpha2.AddToContainer(
		s.container,
		s.InformerFactory,
		s.KubernetesClient.Kubernetes(),
		s.KubernetesClient.Horizon()))

}

func logRequest(req *restful.Request, rep *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, rep)

	logWithVerbose := klog.V(4)
	if rep.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		ip.RemoteIp(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		rep.StatusCode(),
		rep.ContentLength(),
		time.Since(start)/time.Millisecond,
	)

}
