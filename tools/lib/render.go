package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/klog/v2"
	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/common/restfuladapter"
	"k8s.io/kube-openapi/pkg/validation/spec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	clientrest "k8s.io/client-go/rest"
)

type Config struct {
	Scheme *runtime.Scheme
	Codecs serializer.CodecFactory

	Info               spec.InfoProps
	OpenAPIDefinitions []common.GetOpenAPIDefinitions
	Resources          []schema.GroupVersionResource
	Mapper             *meta.DefaultRESTMapper
}

func RenderOpenAPISpec(cfg Config) (string, error) {
	metav1.AddToGroupVersion(cfg.Scheme, schema.GroupVersion{Version: "v1"})

	sc := schema.GroupVersion{Group: "", Version: "v1"}
	cfg.Scheme.AddUnversionedTypes(sc,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)

	recommendedOptions := options.NewRecommendedOptions("/registry/foo.com", cfg.Codecs.LegacyCodec())
	recommendedOptions.SecureServing.BindPort = 8443
	recommendedOptions.Etcd = nil
	recommendedOptions.Authentication = nil
	recommendedOptions.Authorization = nil
	recommendedOptions.CoreAPI = nil
	recommendedOptions.Admission = nil

	if err := recommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		log.Fatal(fmt.Errorf("error creating self-signed certificates: %v", err))
	}

	serverConfig := server.NewRecommendedConfig(cfg.Codecs)
	serverConfig.ClientConfig = &clientrest.Config{}
	serverConfig.SharedInformerFactory = informers.NewSharedInformerFactory(nil, time.Duration(time.Second))

	if err := recommendedOptions.ApplyTo(serverConfig); err != nil {
		log.Fatal(err)
		return "", err
	}

	klog.V(0).Info("get openapi config")

	serverConfig.OpenAPIConfig = server.DefaultOpenAPIConfig(cfg.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(cfg.Scheme))
	serverConfig.OpenAPIConfig.Info.InfoProps = cfg.Info
	serverConfig.OpenAPIV3Config = server.DefaultOpenAPIV3Config(cfg.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(cfg.Scheme))
	serverConfig.OpenAPIV3Config.Info.InfoProps = cfg.Info

	genericServer, err := serverConfig.Complete().New("stash-server", server.NewEmptyDelegate())
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	{
		table := map[string]map[string]ResourceInfo{}
		for _, gvr := range cfg.Resources {
			var res map[string]ResourceInfo
			if r, found := table[gvr.Group]; found {
				res = r
			} else {
				res = map[string]ResourceInfo{}
				table[gvr.Group] = res
			}

			gvk, err := cfg.Mapper.KindFor(gvr)
			if err != nil {
				log.Fatal(err)
				return "", err
			}

			obj, err := cfg.Scheme.New(gvk)
			if err != nil {
				return "", err
			}
			list, err := cfg.Scheme.New(gvk.GroupVersion().WithKind(gvk.Kind + "List"))
			if err != nil {
				log.Fatal(err)
				return "", err
			}

			res[gvr.Resource] = ResourceInfo{
				gvk:  gvk,
				obj:  obj,
				list: list,
			}
		}

		for g, tp := range table {
			apiGroupInfo := server.NewDefaultAPIGroupInfo(g, cfg.Scheme, metav1.ParameterCodec, cfg.Codecs)
			storage := map[string]map[string]rest.Storage{}
			for r, stuff := range tp {
				if storage[stuff.gvk.Version] == nil {
					storage[stuff.gvk.Version] = map[string]rest.Storage{}
				}
				storage[stuff.gvk.Version][r] = NewREST(stuff)
				storage[stuff.gvk.Version][r+"/status"] = NewStatusREST(
					StatusResourceInfo{
						gvk: struct {
							Group   string
							Version string
							Kind    string
						}{Group: stuff.gvk.Group, Version: stuff.gvk.Version, Kind: stuff.gvk.Kind},
						obj: stuff.obj,
					})
			}
			for version, s := range storage {
				apiGroupInfo.VersionedResourcesStorageMap[version] = s
			}

			if err := genericServer.InstallAPIGroup(&apiGroupInfo); err != nil {
				log.Fatal(err)
				return "", err
			}
		}
	}

	spec, err := builder.BuildOpenAPISpecFromRoutes(restfuladapter.AdaptWebServices(genericServer.Handler.GoRestfulContainer.RegisteredWebServices()), serverConfig.OpenAPIConfig)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(data), nil
}

func (c *Config) GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	out := map[string]common.OpenAPIDefinition{}
	for _, def := range c.OpenAPIDefinitions {
		for k, v := range def(ref) {
			out[k] = v
		}
	}

	return out
}
