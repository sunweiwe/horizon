package app

import (
	"github.com/sunweiwe/horizon/cmd/controller-manager/app/options"
	"github.com/sunweiwe/horizon/pkg/controller/cluster"
	"github.com/sunweiwe/horizon/pkg/controller/namespace"
	"github.com/sunweiwe/horizon/pkg/informers"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	ctrl "sigs.k8s.io/controller-runtime"
)

var allControllers = []string{
	"cluster",
	"namespace",
}

var addSuccessfullyControllers = sets.New[string]()

func addAllControllers(mgr manager.Manager, client k8s.Client, informerFactory informers.InformerFactory,
	cmOptions *options.HorizonControllerManagerOptions,
	stopCh <-chan struct{}) error {

	horizonInformer := informerFactory.HorizonSharedInformerFactory()
	if cmOptions.GetControllerEnabled("cluster") {
		if cmOptions.MultiClusterOptions.Enable {
			clusterController := cluster.NewClusterController(
				client.Kubernetes(),
				client.Horizon(),
				client.Config(),
				horizonInformer.Cluster().V1alpha1().Clusters(),
				cmOptions.MultiClusterOptions.ClusterControllerResyncPeriod,
				cmOptions.MultiClusterOptions.HostClusterName,
			)
			addController(mgr, "cluster", clusterController)
		}
	}

	if cmOptions.GetControllerEnabled("namespace") {
		namespaceReconciler := &namespace.Reconciler{}
		addControllerWithSetup(mgr, "namespace", namespaceReconciler)
	}

	// log all controllers process result
	for _, name := range allControllers {
		if cmOptions.GetControllerEnabled(name) {
			if addSuccessfullyControllers.Has(name) {
				klog.Infof("%s controller is enabled and added successfully.", name)
			} else {
				klog.Infof("%s controller is enabled but is not going to run due to its dependent component being disabled.", name)
			}
		} else {
			klog.Infof("%s controller is disabled by controller selectors.", name)
		}
	}

	return nil
}

func addController(mgr manager.Manager, name string, controller manager.Runnable) {
	if err := mgr.Add(controller); err != nil {
		klog.Fatalf("Unable to create %v controller: %v", name, err)
	}
	addSuccessfullyControllers.Insert(name)
}

type setupableController interface {
	SetupWithManager(mgr ctrl.Manager) error
}

func addControllerWithSetup(mgr manager.Manager, name string, controller setupableController) {
	if err := controller.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create %v controller: %v", name, err)
	}
	addSuccessfullyControllers.Insert(name)
}
