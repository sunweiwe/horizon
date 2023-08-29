package cluster

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"reflect"
	"time"

	clusterv1alpha1 "github.com/sunweiwe/api/cluster/v1alpha1"
	horizon "github.com/sunweiwe/horizon/pkg/client/clientset"
	clusterInformer "github.com/sunweiwe/horizon/pkg/client/informers/cluster/v1alpha1"
	clusterLister "github.com/sunweiwe/horizon/pkg/client/listers/cluster/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	clientCmd "k8s.io/client-go/tools/clientcmd"

	"github.com/sunweiwe/horizon/pkg/apiserver/config"
	"github.com/sunweiwe/horizon/pkg/client/clientset/scheme"
	"github.com/sunweiwe/horizon/pkg/constants"
	"github.com/sunweiwe/horizon/pkg/simple/client/multicluster"
	"github.com/sunweiwe/horizon/pkg/utils/k8sutil"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	maxRetries = 15

	kubeFedNamespace = "kube-federation-system"
	HorizonManaged   = "horizon.io/managed"

	hostClusterName = "horizon"
)

var hostCluster = &clusterv1alpha1.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		Name: "host",
		Annotations: map[string]string{
			"horizon.io/description": "The description was created by Horizon automatically. " +
				"It is recommended that you use the Host Cluster to manage clusters only " +
				"and deploy workloads on Member Clusters.",
		},
		Labels: map[string]string{
			clusterv1alpha1.HostCluster: "",
			HorizonManaged:              "true",
		},
	},
	Spec: clusterv1alpha1.ClusterSpec{
		JoinFederation: true,
		Enable:         true,
		Provider:       "horizon",
		Connection: clusterv1alpha1.Connection{
			Type: clusterv1alpha1.ConnectionTypeDirect,
		},
	},
}

type clusterController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	k8sClient  kubernetes.Interface
	hostConfig *rest.Config

	horizonClient horizon.Interface

	clusterLister    clusterLister.ClusterLister
	clusterHasSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	resyncPeriod time.Duration

	hostClusterName string
}

func NewClusterController(
	k8sClient kubernetes.Interface,
	horizonClient horizon.Interface,
	config *rest.Config,
	clusterInformer clusterInformer.ClusterInformer,
	resyncPeriod time.Duration,
	hostClusterName string,
) *clusterController {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})

	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cluster-controller"})

	c := &clusterController{
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		resyncPeriod:     resyncPeriod,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster"),
		workerLoopPeriod: time.Second,
		horizonClient:    horizonClient,
		k8sClient:        k8sClient,
		hostConfig:       config,
	}
	c.clusterLister = clusterInformer.Lister()
	c.clusterHasSynced = clusterInformer.Informer().HasSynced

	clusterInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueCluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldCluster := oldObj.(*clusterv1alpha1.Cluster)
			newCluster := newObj.(*clusterv1alpha1.Cluster)
			if !reflect.DeepEqual(oldCluster.Spec, newCluster.Spec) || newCluster.DeletionTimestamp != nil {
				c.enqueueCluster(newObj)
			}
		},
		DeleteFunc: c.enqueueCluster,
	}, resyncPeriod)

	return c
}

func (c *clusterController) Start(ctx context.Context) error {
	return c.Run(3, ctx.Done())
}

func (c *clusterController) Run(workers int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.V(0).Info("starting cluster controller")
	defer klog.Info("shutting down cluster controller")

	if !cache.WaitForCacheSync(stopCh, c.clusterHasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	go wait.Until(
		func() {
			if err := c.reconcileHostCluster(); err != nil {
				klog.Errorf("Error create host cluster, error %v", err)
			}

			if err := c.resyncClusters(); err != nil {
				klog.Errorf("Failed to reconcile cluster ready status, err: %v", err)
			}

		}, c.resyncPeriod, stopCh)

	<-stopCh

	return nil
}

func (c *clusterController) worker() {
	for c.processNextItem() {

	}
}

func (c *clusterController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.syncCluster(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *clusterController) reconcileHostCluster() error {
	clusters, err := c.clusterLister.List(labels.SelectorFromSet(labels.Set{clusterv1alpha1.HostCluster: ""}))
	if err != nil {
		return err
	}

	hostKubeConfig, err := buildKubeConfigFromRestConfig(c.hostConfig)
	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		hostCluster.Spec.Connection.KubeConfig = hostKubeConfig
		hostCluster.Name = c.hostClusterName
		// TODO
	} else if len(clusters) > 1 {
		return fmt.Errorf("there MUST not be more than one host clusters, while there are %d", len(clusters))
	}

	cluster := clusters[0].DeepCopy()
	managedByHorizon, ok := cluster.Labels[HorizonManaged]
	if !ok || managedByHorizon != "true" {
		return nil
	}

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		cluster.Spec.Connection.KubeConfig = hostKubeConfig
	} else {
		if bytes.Equal(cluster.Spec.Connection.KubeConfig, hostKubeConfig) {
			return nil
		}
	}

	return nil
}

func (c *clusterController) resyncClusters() error {
	clusters, err := c.clusterLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		key, _ := cache.MetaNamespaceKeyFunc(cluster)
		c.queue.Add(key)
	}

	return nil
}

func (c *clusterController) syncCluster(key string) error {
	klog.V(5).Infof("starting to sync cluster %s", key)
	startTime := time.Now()

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("Not valid controller key %s, %#v", key, err)
		return err
	}

	defer func() {
		klog.V(4).Infof("Finished syncing cluster %s in %s", name, time.Since(startTime))
	}()

	cluster, err := c.clusterLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("Failed to get cluster with name %s, %#v", name, err)
		return err
	}

	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sets.New(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.HostCluster) {
			cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterv1alpha1.Finalizer)
			if cluster, err = c.horizonClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	} else {
		if sets.New(cluster.ObjectMeta.Finalizers...).Has(clusterv1alpha1.Finalizer) {

			finalizers := sets.New(cluster.ObjectMeta.Finalizers...)
			finalizers.Delete(clusterv1alpha1.Finalizer)
			cluster.ObjectMeta.Finalizers = finalizers.UnsortedList()

			if cluster, err = c.horizonClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	}

	oldCluster := cluster.DeepCopy()

	if !cluster.Spec.JoinFederation {
		klog.V(5).Infof("Skipping to join cluster %s cause it is not expected to join", cluster.Name)
		return nil
	}

	if len(cluster.Spec.Connection.KubeConfig) == 0 {
		klog.V(5).Infof("Skipping to join cluster %s cause the kubeconfig is empty", cluster.Name)
		return nil
	}

	clusterConfig, err := clientCmd.RESTConfigFromKubeConfig(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		return fmt.Errorf("Failed to create cluster config for %s: %s", cluster.Name, err)
	}

	clusterClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return fmt.Errorf("Failed to create cluster client for %s: %s", cluster.Name, err)
	}

	if len(cluster.Spec.Connection.KubernetesAPIEndpoint) == 0 {
		cluster.Spec.Connection.KubernetesAPIEndpoint = clusterConfig.Host
	}

	serverVersion, err := clusterClient.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("Failed to get kubernetes version, %#v", err)
		return err
	}
	cluster.Status.KubernetesVersion = serverVersion.GitVersion

	nodes, err := clusterClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get cluster nodes, %#v", err)
		return err
	}
	cluster.Status.NodeCount = len(nodes.Items)

	kubeSystem, err := clusterClient.CoreV1().Namespaces().Get(context.TODO(), metav1.NamespaceSystem, metav1.GetOptions{})
	if err != nil {
		return err
	}
	cluster.Status.UID = kubeSystem.UID

	readyCondition := clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterReady,
		Status:             v1.ConditionTrue,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             string(clusterv1alpha1.ClusterReady),
		Message:            "Cluster is available now",
	}
	c.updateClusterCondition(cluster, readyCondition)
	if err = c.updateKubeConfigExpirationDateCondition(cluster); err != nil {
		klog.Warningf("sync kubeconfig expiration date for cluster %s failed: %v", cluster.Name, err)
	}

	if !reflect.DeepEqual(oldCluster.Status, cluster.Status) {
		_, err = c.horizonClient.ClusterV1alpha1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Failed to update cluster status, %#v", err)
			return err
		}
	}

	if err = c.setClusterNameInConfigMap(clusterClient, cluster.Name); err != nil {
		return err
	}

	return nil
}

func (c *clusterController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		klog.V(2).Infof("Error syncing cluster %s, retrying, %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	klog.V(4).Infof("Dropping cluster %s out of the queue.", key)
	c.queue.Forget(key)
	runtime.HandleError(err)
}

func (c *clusterController) updateClusterCondition(cluster *clusterv1alpha1.Cluster, condition clusterv1alpha1.ClusterCondition) {
	if cluster.Status.Conditions == nil {
		cluster.Status.Conditions = make([]clusterv1alpha1.ClusterCondition, 0)
	}

	newConditions := make([]clusterv1alpha1.ClusterCondition, 0)
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == condition.Type {
			continue
		}
		newConditions = append(newConditions, cond)
	}

	newConditions = append(newConditions, condition)
	cluster.Status.Conditions = newConditions
}

func (c *clusterController) updateKubeConfigExpirationDateCondition(cluster *clusterv1alpha1.Cluster) error {
	if _, ok := cluster.Labels[clusterv1alpha1.HostCluster]; ok {
		return nil
	}

	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy {
		return nil
	}

	klog.V(5).Infof("sync KubeConfig expiration date for cluster %s", cluster.Name)
	notAfter, err := parseKubeConfigExpirationDate(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		return fmt.Errorf("parseKubeConfigExpirationDate for cluster %s failed: %v", cluster.Name, err)
	}

	expiresInSevenDays := v1.ConditionFalse
	expirationDate := ""
	if !notAfter.IsZero() {
		expirationDate = notAfter.String()
		if time.Now().AddDate(0, 0, 7).Sub(notAfter) > 0 {
			expiresInSevenDays = v1.ConditionTrue
		}
	}

	c.updateClusterCondition(cluster, clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterKubeConfigCertExpiresInSevenDays,
		Status:             expiresInSevenDays,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             string(clusterv1alpha1.ClusterKubeConfigCertExpiresInSevenDays),
		Message:            expirationDate,
	})

	return nil
}

func (c *clusterController) setClusterNameInConfigMap(client kubernetes.Interface, name string) error {
	cm, err := client.CoreV1().ConfigMaps(constants.HorizonNamespace).Get(context.TODO(), constants.HorizonConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	configData, err := config.GetFromConfigMap(cm)
	if err != nil {
		return err
	}
	if configData.MultiClusterOptions == nil {
		configData.MultiClusterOptions = &multicluster.Options{}
	}
	if configData.MonitoringOptions.ClusterName == name {
		return nil
	}

	configData.MultiClusterOptions.ClusterName = name
	newConfigData, err := yaml.Marshal(configData)
	if err != nil {
		return err
	}
	cm.Data[constants.HorizonConfigMapDataKey] = string(newConfigData)
	if _, err = client.CoreV1().ConfigMaps(constants.HorizonNamespace).Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func parseKubeConfigExpirationDate(kubeconfig []byte) (time.Time, error) {
	config, err := k8sutil.LoadKubeConfigFromBytes(kubeconfig)
	if err != nil {
		return time.Time{}, err
	}

	if config.CertData == nil {
		return time.Time{}, nil
	}

	block, _ := pem.Decode(config.CertData)
	if block == nil {
		return time.Time{}, fmt.Errorf("pem.Decode failed, got empty block data")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

func (c *clusterController) enqueueCluster(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilRuntime.HandleError(fmt.Errorf("get cluster key %s failed", cluster.Name))
		return
	}

	c.queue.Add(key)
}
