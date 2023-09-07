package namespace

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	controllerName = "namespace-controller"
)

type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {

	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	if r.Logger.GetSink() == nil {
		r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	}

	if r.Recorder == nil {
		r.Recorder = mgr.GetEventRecorderFor(controllerName)
	}

	if r.MaxConcurrentReconciles <= 0 {
		r.MaxConcurrentReconciles = 1
	}

	return ctrl.NewControllerManagedBy(mgr).Named(controllerName).WithOptions(controller.Options{
		MaxConcurrentReconciles: r.MaxConcurrentReconciles,
	}).For(&corev1.Namespace{}).Complete(r)

}

// TODO
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(0).Infof("Namespace reconcile starting")

	return ctrl.Result{}, nil
}
