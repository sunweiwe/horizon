package k8s

import (
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	horizon "github.com/sunweiwe/horizon/pkg/client/clientset/versioned"
	apiextensionsClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	clientCmd "k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	Config() *rest.Config
	Kubernetes() kubernetes.Interface
	Horizon() horizon.Interface
}

type kubernetesClient struct {
	k8s kubernetes.Interface
	hz  horizon.Interface

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
		k8s:           kubernetes.NewForConfigOrDie(config),
		hz:            horizon.NewForConfigOrDie(config),
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
	k.k8s, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.hz, err = horizon.NewForConfig(config)
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
	return k.k8s
}

func (k *kubernetesClient) Horizon() horizon.Interface {
	return k.hz
}

func NewKubernetesClientOptions() (option *KubernetesOptions) {
	option = &KubernetesOptions{
		QPS:   1e6,
		Burst: 1e6,
	}

	return
}
