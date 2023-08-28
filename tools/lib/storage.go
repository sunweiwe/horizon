package lib

import (
	"context"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
)

type ResourceInfo struct {
	gvk  schema.GroupVersionKind
	obj  runtime.Object
	list runtime.Object
}

type StatusResourceInfo struct {
	gvk schema.GroupVersionKind
	obj runtime.Object
}

type StandardStorage struct {
	cfg ResourceInfo
}

func (r *StandardStorage) Destroy() {
}

type StatusStandardStorage struct {
	cfg StatusResourceInfo
}

func (r *StatusStandardStorage) Destroy() {
}

var _ rest.GroupVersionKindProvider = &StandardStorage{}
var _ rest.StandardStorage = &StandardStorage{}
var _ rest.GroupVersionKindProvider = &StatusStandardStorage{}
var _ rest.Patcher = &StatusStandardStorage{}

func NewREST(cfg ResourceInfo) *StandardStorage {
	return &StandardStorage{cfg}
}

func NewStatusREST(cfg StatusResourceInfo) *StatusStandardStorage {
	return &StatusStandardStorage{cfg}
}

func (r *StandardStorage) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return r.cfg.gvk
}

func (r *StandardStorage) New() runtime.Object {
	return r.cfg.obj
}

func (r *StandardStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	return r.New(), nil
}

func (r *StandardStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.New(), nil
}

func (r *StandardStorage) NewList() runtime.Object {
	return r.cfg.list
}

func (r *StandardStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	return r.NewList(), nil
}

func (r *StandardStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	return r.New(), true, nil
}

func (r *StandardStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	return r.New(), true, nil
}

func (r *StandardStorage) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	return r.NewList(), nil
}

func (r *StandardStorage) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (r *StandardStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return nil, nil
}

func (r *StandardStorage) NamespaceScoped() bool {
	return false
}

func (r *StatusStandardStorage) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return r.cfg.gvk
}

func (r *StatusStandardStorage) New() runtime.Object {
	return r.cfg.obj
}

func (r *StatusStandardStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	return r.New(), true, nil
}

func (r *StatusStandardStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.New(), nil
}
