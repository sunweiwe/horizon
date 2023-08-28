package v1alpha1

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/sunweiwe/api/cluster/v1alpha1"
	"github.com/sunweiwe/horizon/pkg/api"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var errClusterConnectionIsNotProxy = fmt.Errorf("cluster is not using proxy connection")

func (h *handler) generateAgentDeployment(request *restful.Request, response *restful.Response) {
	clusterName := request.PathParameter("cluster")

	cluster, err := h.clusterLister.Get(clusterName)
	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	if cluster.Spec.Connection.Type != v1alpha1.ConnectionTypeProxy {
		api.HandleNotFound(response, request, fmt.Errorf("cluster %s is not using proxy connection", cluster.Name))
		return
	}

	if len(h.proxyAddress) == 0 {
		err = h.populateProxyAddress()
		if err != nil {
			api.HandleNotFound(response, request, err)
			return
		}
	}
	var buf bytes.Buffer

	err = h.generateDefaultDeployment(cluster, &buf)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.Write(buf.Bytes())
}

func (h *handler) populateProxyAddress() error {
	if len(h.proxyAddress) == 0 {
		return fmt.Errorf("neither proxy address nor proxy service provided")
	}

	namespace := "horizon-system"
	parts := strings.Split(h.proxyService, ".")
	if len(parts) > 1 && len(parts[1]) != 0 {
		namespace = parts[1]
	}

	service, err := h.serviceLister.Services(namespace).Get(parts[0])
	if err != nil {
		return fmt.Errorf("service %s not found in namespace %s", parts[0], namespace)
	}

	if len(service.Spec.Ports) == 0 {
		return fmt.Errorf("there are no ports in proxy service %s spec", h.proxyService)
	}

	port := service.Spec.Ports[0].Port

	var serviceAddress string
	for _, ingress := range service.Status.LoadBalancer.Ingress {
		if len(ingress.Hostname) != 0 {
			serviceAddress = fmt.Sprintf("http://%s:%d", ingress.Hostname, port)
		}

		if len(ingress.IP) != 0 {
			serviceAddress = fmt.Sprintf("http://%s:%d", ingress.IP, port)
		}
	}

	if len(serviceAddress) == 0 {
		return fmt.Errorf("cannot generate agent deployment yaml for member cluster "+
			" because %s service has no public address, please check %s status, or set address "+
			" manually in ClusterConfiguration", h.proxyService, h.proxyService)
	}

	h.proxyAddress = serviceAddress
	return nil
}

func (h *handler) generateDefaultDeployment(cluster *v1alpha1.Cluster, w io.Writer) error {
	_, err := url.Parse(h.proxyAddress)
	if err != nil {
		return fmt.Errorf("invalid proxy address %s, should format like http[s]://1.2.3.4;123", h.proxyAddress)
	}

	if cluster.Spec.Connection.Type == v1alpha1.ConnectionTypeDirect {
		return errClusterConnectionIsNotProxy
	}

	agent := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-agent",
			Namespace: "horizon-system",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                       "agent",
					"app.kubernetes.io/part-of": "tower",
				},
			},
			Strategy: appsv1.DeploymentStrategy{},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                       "agent",
						"app.kubernetes.io/part-of": "tower",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "agent",
							Command: []string{
								"/agent",
								fmt.Sprintf("--name=%s", cluster.Name),
								fmt.Sprintf("--token=%s", cluster.Spec.Connection.Token),
								fmt.Sprintf("--proxy-server=%s", h.proxyAddress),
								"--keepalive=10s",
								"--horizon-service=ks-apiserver.horizon-system.svc:80",
								"--kubernetes-service=kubernetes.default.svc:443",
								"--v=0",
							},
							Image: h.agentImage,
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("200M"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100M"),
									v1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
					ServiceAccountName: "horizon",
				},
			},
		},
	}

	return h.yamlPrinter.PrintObj(&agent, w)
}
