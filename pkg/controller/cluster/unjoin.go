package cluster

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteServiceAccount(clusterClientset kubernetes.Interface, name, namespace, unJoiningClusterName string, dryRun bool) error {
	if dryRun {
		return nil
	}

	klog.V(2).Infof("Deleting service account \"%s/%s\" in unjoining cluster %q.", namespace, name, unJoiningClusterName)

	err := clusterClientset.CoreV1().ServiceAccounts(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("Service account  \"%s/%s\" does not exist.", namespace, name)
	} else if err != nil {
		return errors.Wrapf(err, "Could not delete service account \"%s/%s\"", namespace, name)
	} else {
		klog.V(2).Infof("Deleted service account \"%s/%s\" in unjoining cluster %q.", namespace, name, unJoiningClusterName)
	}

	return nil
}

func deleteFedNSFromUnJoinCluster(hostClientset, unJoiningClusterClientset kubernetes.Interface,
	kubefedNamespace, unjoiningClusterName string, dryRun bool) error {

	if dryRun {
		return nil
	}

	hostClusterNamespace, err := hostClientset.CoreV1().Namespaces().Get(context.Background(), kubefedNamespace, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error retrieving namespace %q from host cluster", kubefedNamespace)
	}

	unjoiningClusterNamespace, err := unJoiningClusterClientset.CoreV1().Namespaces().Get(context.Background(), kubefedNamespace, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error retrieving namespace %q from unjoining cluster %q", kubefedNamespace, unjoiningClusterName)
	}

	if IsPrimaryCluster(hostClusterNamespace, unjoiningClusterNamespace) {
		klog.V(2).Infof("The kubefed namespace %q does not need to be deleted from the host cluster by unJoin.", kubefedNamespace)
		return nil
	}

	klog.V(2).Infof("Deleting kubefed namespace %q from unjoining cluster %q.", kubefedNamespace, unjoiningClusterName)
	err = unJoiningClusterClientset.CoreV1().Namespaces().Delete(context.Background(), kubefedNamespace, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		klog.V(2).Infof("The kubefed namespace %q no longer exists in unjoining cluster %q.", kubefedNamespace, unjoiningClusterName)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "Could not delete kubefed namespace %q from unjoining cluster %q", kubefedNamespace, unjoiningClusterName)
	} else {
		klog.V(2).Infof("Deleted kubefed namespace %q from unjoining cluster %q.", kubefedNamespace, unjoiningClusterName)
	}

	return nil
}
