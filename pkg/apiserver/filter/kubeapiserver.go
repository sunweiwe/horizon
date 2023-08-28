package filter

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sunweiwe/horizon/pkg/apiserver/request"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type kubeAPIProxy struct {
	next          http.Handler
	kubeAPIServer *url.URL
	transport     http.RoundTripper
}

func WithKubeAPIServer(next http.Handler, config *rest.Config) http.Handler {
	kubeAPIServer, _ := url.Parse(config.Host)
	transport, err := rest.TransportFor(config)
	if err != nil {
		klog.Errorf("Unable to create transport form rest.Config: %v", err)
		return next
	}

	return &kubeAPIProxy{
		next:          next,
		kubeAPIServer: kubeAPIServer,
		transport:     transport,
	}
}

func (k kubeAPIProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	info, ok := request.RequestInfoFrom(req.Context())
	if !ok {
		responsewriters.InternalError(w, req, fmt.Errorf("no RequestInfo found in the context"))
		return
	}

	if info.KubernetesRequest {
		s := *req.URL
		s.Host = k.kubeAPIServer.Host
		s.Scheme = k.kubeAPIServer.Scheme

		req.Header.Del("Authorization")
		httpProxy := proxy.NewUpgradeAwareHandler(
			&s,
			k.transport,
			true,
			false,
			&responder{},
		)
		httpProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(k.transport, k.transport)
		httpProxy.ServeHTTP(w, req)
		return
	}

	k.next.ServeHTTP(w, req)
}
