package v1beta1

import (
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type resourceManager struct {
	client client.Client
	cache  cache.Cache
}

func New(client client.Client, cache cache.Cache) ResourceManager {
	return &resourceManager{
		client: client,
		cache:  cache,
	}
}

const labelResourceServed = "horizon.io/resource-served"

func (r *resourceManager) GetResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, name string) (client.Object, error) {
	var obj client.Object
	gvk, err := r.getGVK(gvr)
	if err != nil {
		return nil, err
	}

	if r.client.Scheme().Recognizes(gvk) {
		gvkObject, err := r.client.Scheme().New(gvk)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.Object)
	} else {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	if err := r.Get(ctx, namespace, name, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *resourceManager) IsServed(gvr schema.GroupVersionResource) (bool, error) {

	if r.client.Scheme().IsVersionRegistered(gvr.GroupVersion()) {
		return true, nil
	}

	crd := &apiextensions.CustomResourceDefinition{}
	if err := r.cache.Get(context.Background(), client.ObjectKey{Name: gvr.GroupResource().String()}, crd); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if crd.Labels["labelResourceServed"] == "true" {
		return true, nil
	}

	return false, nil
}

func (r *resourceManager) CreateObjectFromRawData(gvr schema.GroupVersionResource, rawData []byte) (client.Object, error) {
	var obj client.Object
	gvk, err := r.getGVK(gvr)
	if err != nil {
		return nil, err
	}

	if r.client.Scheme().Recognizes(gvk) {
		gvkObject, err := r.client.Scheme().New(gvk)
		if err != nil {
			return nil, err
		}
		obj = gvkObject.(client.Object)
	} else {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		obj = u
	}

	err = json.Unmarshal(rawData, obj)
	if err != nil {
		return nil, err
	}

	if obj.GetObjectKind().GroupVersionKind().String() != gvk.String() {
		return nil, errors.NewBadRequest("Wrong resource GroupVersionKind")
	}

	return obj, nil
}

func (r *resourceManager) getGVK(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	var (
		gvk schema.GroupVersionKind
		err error
	)

	gvk, err = r.client.RESTMapper().KindFor(gvr)
	if err != nil {
		return gvk, err
	}

	return gvk, nil
}

func (r *resourceManager) Get(ctx context.Context, namespace, name string, object client.Object) error {
	return r.cache.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, object)
}

func (r *resourceManager) CreateResource(ctx context.Context, object client.Object) error {
	return r.Create(ctx, object)
}

func (h *resourceManager) Create(ctx context.Context, object client.Object) error {
	return h.client.Create(ctx, object)
}
