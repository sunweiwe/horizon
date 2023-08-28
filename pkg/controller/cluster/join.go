package cluster

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/sunweiwe/api/types/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	namespacedPolicyRules = []rbacv1.PolicyRule{
		{
			Verbs:     []string{rbacv1.VerbAll},
			APIGroups: []string{rbacv1.APIGroupAll},
			Resources: []string{rbacv1.ResourceAll},
		},
	}
	clusterPolicyRules = []rbacv1.PolicyRule{
		namespacedPolicyRules[0],
		{
			NonResourceURLs: []string{rbacv1.NonResourceAll},
			Verbs:           []string{"get"},
		},
	}
	localSchemeBuilder = runtime.SchemeBuilder{
		kubernetesscheme.AddToScheme,
		v1beta1.AddToScheme,
	}
)

const (
	tokenKey                    = "token"
	serviceAccountSecretTimeout = 30 * time.Second
	kubefedManagedSelector      = "kubefed.io/managed=true"
)

func performPreflightChecks() error {

	return nil
}

func createKubeFedNamespace(clusterClientSet kubernetes.Interface, kubefedNamespace,
	joiningClusterName string, dryRun bool) (*corev1.Namespace, error) {
	fedNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kubefedNamespace,
		},
	}

	if dryRun {
		return fedNamespace, nil
	}

	_, err := clusterClientSet.CoreV1().Namespaces().Get(context.Background(), kubefedNamespace, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.V(2).Infof("Could not get %s namespace: %v", kubefedNamespace, err)
		return nil, err
	}

	if err == nil {
		klog.V(2).Infof("Already existing %s namespace", kubefedNamespace)
		return fedNamespace, nil
	}

	_, err = clusterClientSet.CoreV1().Namespaces().Create(context.Background(), fedNamespace, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.V(2).Infof("Could not create %s namespace: %v", kubefedNamespace, err)
		return nil, err
	}

	return fedNamespace, nil
}

func createAuthorizedServiceAccount(joiningClusterClientset kubernetes.Interface,
	namespace, joiningClusterName, hostClusterName string,
	scope apiextensionsv1.ResourceScope, dryRun, errorOnExisting bool) (string, error) {
	klog.V(2).Infof("Creating service account in joining cluster: %s", joiningClusterName)

	n, err := createServiceAccountWithSecret(joiningClusterClientset, namespace, joiningClusterName, hostClusterName, dryRun, errorOnExisting)
	if err != nil {
		klog.V(2).Infof("Error creating service account: %s in joining cluster: %s due to: %v",
			n, joiningClusterName, err)
		return "", err
	}

	klog.V(2).Infof("Created service account: %s in joining cluster: %s", n, joiningClusterName)

	if scope == apiextensionsv1.NamespaceScoped {
		klog.V(2).Infof("Creating role and binding for service account: %s in joining cluster: %s", n, joiningClusterName)

		err = createRoleAndBinding(joiningClusterClientset, n, namespace, joiningClusterName, dryRun, errorOnExisting)
		if err != nil {
			klog.V(2).Infof("Error creating role and binding for service account: %s in joining cluster: %s due to: %v", n, joiningClusterName, err)
			return "", err
		}

		klog.V(2).Infof("Created role and binding for service account: %s in joining cluster: %s",
			n, joiningClusterName)

		klog.V(2).Infof("Creating health check cluster role and binding for service account: %s in joining cluster: %s", n, joiningClusterName)

		err = createHealthCheckClusterRoleAndBinding(joiningClusterClientset, n, namespace, joiningClusterName, dryRun, errorOnExisting)
		if err != nil {
			klog.V(2).Infof("Error creating health check cluster role and binding for service account: %s in joining cluster: %s due to: %v",
				n, joiningClusterName, err)
			return "", err
		}

		klog.V(2).Infof("Created health check cluster role and binding for service account: %s in joining cluster: %s",
			n, joiningClusterName)

	} else {
		klog.V(2).Infof("Creating cluster role and binding for service account: %s in joining cluster: %s", n, joiningClusterName)

		err = createClusterRoleAndBinding(joiningClusterClientset, n, namespace, joiningClusterName, dryRun, errorOnExisting)
		if err != nil {
			klog.V(2).Infof("Error creating cluster role and binding for service account: %s in joining cluster: %s due to: %v",
				n, joiningClusterName, err)
			return "", err
		}

		klog.V(2).Infof("Created cluster role and binding for service account: %s in joining cluster: %s",
			n, joiningClusterName)
	}

	return n, nil
}

func createServiceAccountWithSecret(clusterClientSet kubernetes.Interface, namespace,
	joiningClusterName, hostClusterName string, dryRun, errorOnExisting bool) (string, error) {
	n := fmt.Sprintf("%s-%s", joiningClusterName, hostClusterName)

	if dryRun {
		return n, nil
	}

	ctx := context.Background()
	sa, err := clusterClientSet.CoreV1().ServiceAccounts(namespace).Get(ctx, n, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			sa = &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      n,
					Namespace: namespace,
				},
			}

			sa, err = clusterClientSet.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
			switch {
			case apierrors.IsAlreadyExists(err) && errorOnExisting:
				klog.V(2).Infof("Service account %s/%s already exists in target cluster %s", namespace, n, joiningClusterName)
				return "", err
			case err != nil && apierrors.IsAlreadyExists(err):
				klog.V(2).Infof("Could not create service account %s/%s in target cluster %s due to: %v", namespace, n, joiningClusterName, err)
				return "", err
			}
		} else {
			return "", err
		}
	}

	if len(sa.Secrets) > 0 {
		return n, nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-token-", n),
			Namespace:    namespace,
			Annotations: map[string]string{
				corev1.ServiceAccountNameKey: n,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

	secret, err = clusterClientSet.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})

	if err != nil {
		klog.V(2).Infof("Could not create secret for service account %s/%s in target cluster %s due to: %v", namespace, n, joiningClusterName, err)
		return "", err
	}

	sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: secret.Name})
	_, err = clusterClientSet.CoreV1().ServiceAccounts(namespace).Update(ctx, sa, metav1.UpdateOptions{})
	switch {
	case err != nil:
		klog.Infof("Could not update service account %s/%s in target cluster %s due to: %v", namespace, n, joiningClusterName, err)
		return "", err
	default:
		return n, nil
	}
}

func createRoleAndBinding(clientset kubernetes.Interface, name, namespace, clusterName string,
	dryRun, errorOnExisting bool) error {
	if dryRun {
		return nil
	}

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: namespacedPolicyRules,
	}

	existingRole, err := clientset.RbacV1().Roles(namespace).Get(context.Background(), name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not retrieve role for service account %s in joining cluster %s due to %v", name, clusterName, err)
		return err
	case errorOnExisting && err == nil:
		return errors.Errorf("role for service account %s in joining cluster %s already exists", name, clusterName)
	case err == nil:
		existingRole.Rules = role.Rules
		_, err = clientset.RbacV1().Roles(namespace).Update(context.Background(), existingRole, metav1.UpdateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not update role for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	default:
		_, err := clientset.RbacV1().Roles(namespace).Create(context.Background(), role, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not create role for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	}

	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     name,
		},
	}

	existingBinding, err := clientset.RbacV1().RoleBindings(namespace).Get(context.Background(), binding.Name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not retrieve role binding for service account %s in joining cluster %s due to: %v",
			name, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("role binding for service account %s in joining cluster %s already exists", name, clusterName)
	case err == nil:
		if !reflect.DeepEqual(existingBinding.RoleRef, binding.RoleRef) {
			err = clientset.RbacV1().RoleBindings(namespace).Delete(context.Background(), existingBinding.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.V(2).Infof("Could not delete existing role binding for service account %s in joining cluster %s due to: %v",
					name, clusterName, err)
				return err
			}
			_, err = clientset.RbacV1().RoleBindings(namespace).Create(context.Background(), binding, metav1.CreateOptions{})
			if err != nil {
				klog.V(2).Infof("Could not create role binding for service account: %s in joining cluster: %s due to: %v",
					name, clusterName, err)
				return err
			}
		} else {
			existingBinding.Subjects = binding.Subjects
			_, err = clientset.RbacV1().RoleBindings(namespace).Update(context.Background(), existingBinding, metav1.UpdateOptions{})
			if err != nil {
				klog.V(2).Infof("Could not update role binding for service account %s in joining cluster %s due to: %v",
					name, clusterName, err)
				return err
			}
		}
	default:
		_, err = clientset.RbacV1().RoleBindings(namespace).Create(context.Background(), binding, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not create role binding for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	}

	return nil
}

func createHealthCheckClusterRoleAndBinding(clientset kubernetes.Interface, name, namespace string,
	clusterName string, dryRun, errorOnExisting bool) error {
	if dryRun {
		return nil
	}

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:           []string{"Get"},
				NonResourceURLs: []string{"/healthz"},
			},
			{
				Verbs:     []string{"list"},
				APIGroups: []string{""},
				Resources: []string{"nodes"},
			},
		},
	}

	existingRole, err := clientset.RbacV1().ClusterRoles().Get(context.Background(), role.Name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get health check cluster role for service account %s in joining cluster %s due to %v",
			name, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("health check cluster role for service account %s in joining cluster %s already exists", name, clusterName)
	case err == nil:
		existingRole.Rules = role.Rules
		_, err := clientset.RbacV1().ClusterRoles().Update(context.Background(), existingRole, metav1.UpdateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not update health check cluster role for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	default:
		_, err := clientset.RbacV1().ClusterRoles().Create(context.Background(), role, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not create health check cluster role for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	}

	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     name,
		},
	}

	existingBinding, err := clientset.RbacV1().ClusterRoleBindings().Get(context.Background(), binding.Name, metav1.GetOptions{})
	switch {
	case err != nil && apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get health check cluster role binding for service account %s in joining cluster %s due to %v",
			name, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("health check cluster role binding for service account %s in joining cluster %s already exists", name, clusterName)
	case err == nil:
		if !reflect.DeepEqual(existingBinding.RoleRef, binding.RoleRef) {
			err = clientset.RbacV1().ClusterRoleBindings().Delete(context.Background(), existingBinding.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.V(2).Infof("Could not delete existing health check cluster role binding for service account %s in joining cluster %s due to: %v",
					name, clusterName, err)
				return err
			}
			_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), binding, metav1.CreateOptions{})
			if err != nil {
				klog.V(2).Infof("Could not create health check cluster role binding for service account: %s in joining cluster: %s due to: %v",
					name, clusterName, err)
				return err
			}
		} else {
			existingBinding.Subjects = binding.Subjects
			_, err := clientset.RbacV1().ClusterRoleBindings().Update(context.Background(), existingBinding, metav1.UpdateOptions{})
			if err != nil {
				klog.V(2).Infof("Could not update health check cluster role binding for service account: %s in joining cluster: %s due to: %v",
					name, clusterName, err)
				return err
			}
		}
	default:
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), binding, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not create health check cluster role binding for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	}

	return nil
}

func createClusterRoleAndBinding(clientset kubernetes.Interface, name, namespace string,
	clusterName string, dryRun, errorOnExisting bool) error {
	if dryRun {
		return nil
	}

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: clusterPolicyRules,
	}
	existingRole, err := clientset.RbacV1().ClusterRoles().Get(context.Background(), name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get cluster role for service account %s in joining cluster %s due to %v",
			name, clusterName, err)
		return err
	case err == nil && errorOnExisting:
		return errors.Errorf("cluster role for service account %s in joining cluster %s already exists", name, clusterName)
	case err == nil:
		existingRole.Rules = role.Rules
		_, err := clientset.RbacV1().ClusterRoles().Update(context.Background(), existingRole, metav1.UpdateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not update cluster role for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}

	default:
		_, err := clientset.RbacV1().ClusterRoles().Create(context.Background(), role, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not create cluster role for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	}

	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     name,
		},
	}
	existingBinding, err := clientset.RbacV1().ClusterRoleBindings().Get(context.Background(), binding.Name, metav1.GetOptions{})
	switch {
	case err != nil && !apierrors.IsNotFound(err):
		klog.V(2).Infof("Could not get cluster role binding for service account %s in joining cluster %s due to %v",
			name, clusterName, err)
	case err == nil && errorOnExisting:
		return errors.Errorf("cluster role binding for service account %s in joining cluster %s already  exists", name, clusterName)
	case err == nil:
		if !reflect.DeepEqual(existingBinding.RoleRef, binding.RoleRef) {
			err = clientset.RbacV1().ClusterRoleBindings().Delete(context.Background(), existingBinding.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.V(2).Infof("Could not delete existing cluster role binding for service account %s in joining cluster %s due to:%v",
					name, clusterName, err)
				return err
			}
			_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), binding, metav1.CreateOptions{})
			if err != nil {
				klog.V(2).Infof("Could not create cluster role binding for service account: %s in joining cluster: %s due to: %v",
					name, clusterName, err)
				return err
			}
		} else {
			existingBinding.Subjects = binding.Subjects
			_, err := clientset.RbacV1().ClusterRoleBindings().Update(context.Background(), existingBinding, metav1.UpdateOptions{})
			if err != nil {
				klog.V(2).Infof("Could not update cluster role binding for service account: %s in joining cluster: %s due to: %v",
					name, clusterName, err)
				return err
			}
		}
	default:
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), existingBinding, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Could not create cluster role binding for service account: %s in joining cluster: %s due to: %v",
				name, clusterName, err)
			return err
		}
	}

	return nil
}
