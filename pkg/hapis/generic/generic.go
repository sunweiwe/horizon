package generic

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/horizon/pkg/api"
	"github.com/sunweiwe/horizon/pkg/apiserver/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/klog/v2"
)

type genericProxy struct {
	Endpoint *url.URL

	GroupName string

	Version string
}

func NewGenericProxy(endpoint string, groupName string, version string) (*genericProxy, error) {
	parse, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	parse.Path = strings.Trim(parse.Path, "/")

	return &genericProxy{
		Endpoint:  parse,
		GroupName: groupName,
		Version:   version,
	}, nil
}

func (g *genericProxy) AddToContainer(container *restful.Container) error {
	webService := runtime.NewWebService(schema.GroupVersion{
		Group:   g.GroupName,
		Version: g.Version,
	})

	webService.Route(webService.GET("/{path:*}").To(g.handler).
		Returns(http.StatusOK, api.StatusOK, nil))

	container.Add(webService)
	return nil
}

func (g *genericProxy) handler(request *restful.Request, response *restful.Response) {
	u := g.makeURL(request)

	httpProxy := proxy.NewUpgradeAwareHandler(u, http.DefaultTransport, false, false, &errorResponder{})
	httpProxy.ServeHTTP(response, request.Request)
}

func (g *genericProxy) makeURL(request *restful.Request) *url.URL {
	u := *(request.Request.URL)
	u.Host = g.Endpoint.Host
	u.Scheme = g.Endpoint.Scheme
	u.Path = strings.Replace(request.Request.URL.Path, fmt.Sprintf("/hapis/%s", g.GroupName), "", 1)

	if len(g.Endpoint.Host) != 0 {
		u.Path = fmt.Sprintf("/%s%s", g.Endpoint.Path, u.Path)
	}

	return &u
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
}
