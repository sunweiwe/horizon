package metricsserver

import (
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"github.com/sunweiwe/horizon/pkg/simple/client/monitoring"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

func NewMetricsClient(k kubernetes.Interface, options *k8s.KubernetesOptions) monitoring.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		klog.Error(err)
		return nil
	}

	discoveryClient := k.Discovery()
	apiGroups, err := discoveryClient.ServerGroups()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsAPIAvailable := metricsAPISupported(apiGroups)
	if !metricsAPIAvailable {
		klog.Warningf("Metrics API not available.")
		return nil
	}

	metricsClient, err := metricsclient.NewForConfig(config)
	if err != nil {
		klog.Error(err)
		return nil
	}

	return NewMetricsServer(k, metricsAPIAvailable, metricsClient)
}

var (
	supportedMetricsAPIs = map[string]bool{
		"v1beta1": true,
	}
)

func metricsAPISupported(discoveredAPIGroups *metav1.APIGroupList) bool {
	for _, group := range discoveredAPIGroups.Groups {
		if group.Name != metrics.GroupName {
			continue
		}

		for _, version := range group.Versions {
			if _, found := supportedMetricsAPIs[version.Version]; found {
				return true
			}
		}
	}
	return false
}

type metricsServer struct {
	metricsAPIAvailable bool
	metricsClient       metricsclient.Interface
	k8s                 kubernetes.Interface
}

func NewMetricsServer(k kubernetes.Interface, available bool, m metricsclient.Interface) monitoring.Interface {
	var metricsserver metricsServer

	metricsserver.k8s = k
	metricsserver.metricsAPIAvailable = available
	metricsserver.metricsClient = m

	return metricsserver
}
