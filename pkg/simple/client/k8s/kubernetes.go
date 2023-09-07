package k8s

import (
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/sunweiwe/horizon/pkg/client/clientset"
	apiextensionsClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	clientCmd "k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	Config() *rest.Config
	Kubernetes() kubernetes.Interface
	Horizon() clientset.Interface
}

type kubernetesClient struct {
	kube    kubernetes.Interface
	horizon clientset.Interface

	apiextensions apiextensionsClient.Interface

	master string

	config *rest.Config
}

func NewKubernetesClientOrDie(options *KubernetesOptions) Client {
	config, err := clientCmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		panic(err)
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	k := &kubernetesClient{
		kube:          kubernetes.NewForConfigOrDie(config),
		horizon:       clientset.NewForConfigOrDie(config),
		apiextensions: apiextensionsClient.NewForConfigOrDie(config),
		master:        config.Host,
		config:        config,
	}

	if options.Master != "" {
		k.master = options.Master
	}

	if !strings.HasPrefix(k.master, "http://") {
		k.master = "https://" + k.master
	}

	return k
}

func NewKubernetesClient(options *KubernetesOptions) (Client, error) {
	config, err := clientCmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		return nil, err
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	var k kubernetesClient
	k.kube, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.horizon, err = clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.apiextensions, err = apiextensionsClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.master = options.Master
	k.config = config

	return &k, nil
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}

func (k *kubernetesClient) Kubernetes() kubernetes.Interface {
	return k.kube
}

func (k *kubernetesClient) Horizon() clientset.Interface {
	return k.horizon
}
