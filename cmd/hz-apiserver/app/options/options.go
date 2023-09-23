package options

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/sunweiwe/horizon/pkg/apis"
	"github.com/sunweiwe/horizon/pkg/apiserver"
	"github.com/sunweiwe/horizon/pkg/informers"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"github.com/sunweiwe/horizon/pkg/simple/client/monitoring/metricsserver"
	"github.com/sunweiwe/horizon/pkg/simple/client/monitoring/prometheus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	apiserverconfig "github.com/sunweiwe/horizon/pkg/apiserver/config"
	genericoptions "github.com/sunweiwe/horizon/pkg/server/options"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ServerRunOptions struct {
	ConfigFile string
	schemeOnce sync.Once
	*apiserverconfig.Config
	DebugMode               bool
	GenericServerRunOptions *genericoptions.ServerRunOptions
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		schemeOnce:              sync.Once{},
		Config:                  apiserverconfig.New(),
	}

	return s
}

func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Config: s.Config,
	}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.Horizon())
	apiServer.InformerFactory = informerFactory

	if s.MonitoringOptions == nil || len(s.MonitoringOptions.Endpoint) == 0 {
		return nil, fmt.Errorf("monitor service address in configuration MUST not be empty, please check configmap/horizon-config in horizon-system namespace")
	} else {
		if apiServer.MonitoringClient, err = prometheus.NewPrometheus(s.MonitoringOptions); err != nil {
			return nil, fmt.Errorf("failed to connect to prometheus, please check prometheus status, error: %v", err)
		}
	}

	apiServer.MetricsClient = metricsserver.NewMetricsClient(kubernetesClient.Kubernetes(), s.KubernetesOptions)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	sch := scheme.Scheme
	s.schemeOnce.Do(func() {
		if err := apis.AddToScheme(sch); err != nil {
			klog.Fatalf("unable add APIs to scheme: %v", err)
		}
	})

	client, _ := rest.HTTPClientFor(apiServer.KubernetesClient.Config())
	mapper, err := apiutil.NewDynamicRESTMapper(apiServer.KubernetesClient.Config(), client)
	if err != nil {
		klog.Fatalf("unable create dynamic RESTMapper: %v", err)
	}

	apiServer.RuntimeCache, err = runtimecache.New(apiServer.KubernetesClient.Config(), runtimecache.Options{Scheme: sch, Mapper: mapper})
	if err != nil {
		klog.Fatalf("unable to create controller runtime cache: %v", err)
	}

	apiServer.RuntimeClient, err = runtimeclient.New(apiServer.KubernetesClient.Config(), runtimeclient.Options{Scheme: sch})
	if err != nil {
		klog.Fatalf("unable to create controller runtime client: %v", err)
	}

	apiServer.Server = server

	return apiServer, nil
}

func (s *ServerRunOptions) Flags() (fss flag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", true, "Don't enable this if you don't know what it means.")

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"), s.MonitoringOptions)

	return fss
}
