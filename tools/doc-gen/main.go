package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/go-openapi/spec"
	"github.com/sunweiwe/horizon/pkg/apiserver/runtime"
	"github.com/sunweiwe/horizon/pkg/constants"
	"github.com/sunweiwe/horizon/pkg/informers"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	clusterv1alpha1 "github.com/sunweiwe/horizon/pkg/hapis/cluster/v1alpha1"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/version"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/hz-openapi-spec/swagger.json", "--output=./api.json")
}

func main() {
	flag.Parse()

	generateSwaggerJson()

}

func generateSwaggerJson() []byte {
	container := runtime.Container
	clientsets := k8s.NewNullClient()

	informerFactory := informers.NewNullInformerFactory()
	urlruntime.Must(clusterv1alpha1.AddToContainer(container, clientsets.Horizon(), informerFactory.KubernetesSharedInformerFactory(),
		informerFactory.HorizonSharedInformerFactory(), "", "", ""))

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	swagger := restfulspec.BuildSwagger(config)
	swagger.Info.Extensions = make(spec.Extensions)
	swagger.Info.Extensions.Add("x-tagGroups", []struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		{
			Name: "Authentication",
			Tags: []string{constants.AuthenticationTag},
		},
	})

	data, _ := json.MarshalIndent(swagger, "", "  ")
	err := os.WriteFile(output, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s", output)

	return data
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Horizon",
			Description: "Horizon OpenAPI",
			Version:     version.Get().GitVersion,
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "Horizon",
					URL:   "https://horizon.io/",
					Email: "sun_weiwe@163.com",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache 2.0",
					URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
		},
	}

	// setup security definitions
	swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
		"jwt": spec.APIKeyAuth("Authorization", "header"),
	}
	swo.Security = []map[string][]string{{"jwt": []string{}}}
}
