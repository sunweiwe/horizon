package app

import (
	"github.com/sunweiwe/horizon/cmd/controller-manager/app/options"
	"github.com/sunweiwe/horizon/pkg/controller/cluster"
	"github.com/sunweiwe/horizon/pkg/informers"
	"github.com/sunweiwe/horizon/pkg/simple/client/k8s"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var allControllers = []string{
	"namespace",
	"user",
	"cluster",
	"workspace",
}

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

	return nil
}

func addController(mgr manager.Manager, name string, controller manager.Runnable) {
	if err := mgr.Add(controller); err != nil {
		klog.Fatalf("Unable to create %v controller: %v", name, err)
	}
}
