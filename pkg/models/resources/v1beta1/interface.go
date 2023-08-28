package v1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceManager interface {
	IsServed(schema.GroupVersionResource) (bool, error)
	CreateObjectFromRawData(gvr schema.GroupVersionResource, rawData []byte) (client.Object, error)

	GetResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, name string) (client.Object, error)
	CreateResource(ctx context.Context, object client.Object) error
}
