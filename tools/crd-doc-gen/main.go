package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	clusterv1alpha1 "github.com/sunweiwe/api/cluster/v1alpha1"
	tenantInstaller "github.com/sunweiwe/api/tenant/installer"
	tenantv1alpha1 "github.com/sunweiwe/api/tenant/v1alpha1"

	urlruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/sunweiwe/horizon/tools/lib"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/openapi-spec/swagger.json", "--output=./api/openapi-spec/swagger.json")
}

func main() {

	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	tenantInstaller.Installer(Scheme)

	urlruntime.Must(clusterv1alpha1.AddToScheme(Scheme))
	urlruntime.Must(Scheme.SetVersionPriority(clusterv1alpha1.SchemeGroupVersion))

	mapper := meta.NewDefaultRESTMapper(nil)

	mapper.AddSpecific(clusterv1alpha1.SchemeGroupVersion.WithKind(clusterv1alpha1.ResourceKindCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesSingularCluster), meta.RESTScopeRoot)

	mapper.AddSpecific(clusterv1alpha1.SchemeGroupVersion.WithKind(clusterv1alpha1.ResourceKindCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralCluster),
		clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesSingularCluster), meta.RESTScopeRoot)

	mapper.AddSpecific(tenantv1alpha1.SchemeGroupVersion.WithKind(tenantv1alpha1.ResourceKindWorkspace),
		tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourcePluralWorkspace),
		tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourceSingularWorkspace), meta.RESTScopeRoot)

	spec, err := lib.RenderOpenAPISpec(lib.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title: "Horizon",
			Contact: &spec.ContactInfo{
				Name:  "Horizon",
				URL:   "https://horizon.io/",
				Email: "sun_weiwe@163.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			clusterv1alpha1.GetOpenAPIDefinitions,
			tenantv1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []schema.GroupVersionResource{
			clusterv1alpha1.SchemeGroupVersion.WithResource(clusterv1alpha1.ResourcesPluralCluster),
			tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourcePluralWorkspace),
		},
		Mapper: mapper,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(filepath.Dir(output), 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(output, []byte(spec), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
