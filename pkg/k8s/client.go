package k8s

import (
	"context"
	"fmt"
	"io"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// NewClient creates a new Kubernetes clientset from kubeconfig or in-cluster config
func NewClient(kubeconfig string) (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Warningf("Failed to get in-cluster config: %v, falling back to default kubeconfig", err)
			// Fall back to default kubeconfig location
			config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		klog.Errorf("Failed to build config: %v", err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create clientset: %v", err)
		return nil, err
	}

	return clientset, nil
}

// ListPods lists all pods in the specified namespace
func ListPods(clientset kubernetes.Interface, namespace string) ([]v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list pods in namespace %s: %v", namespace, err)
		return nil, err
	}
	return pods.Items, nil
}

// CreatePod creates a new pod in the specified namespace
func CreatePod(clientset kubernetes.Interface, namespace string, pod *v1.Pod) (*v1.Pod, error) {
	createdPod, err := clientset.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create pod %s in namespace %s: %v", pod.Name, namespace, err)
		return nil, err
	}
	return createdPod, nil
}

// UpdatePod updates an existing pod in the specified namespace
func UpdatePod(clientset kubernetes.Interface, namespace string, pod *v1.Pod) (*v1.Pod, error) {
	updatedPod, err := clientset.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update pod %s in namespace %s: %v", pod.Name, namespace, err)
		return nil, err
	}
	return updatedPod, nil
}

// DeletePod deletes a pod in the specified namespace
func DeletePod(clientset kubernetes.Interface, namespace, name string) error {
	err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete pod %s in namespace %s: %v", name, namespace, err)
		return err
	}
	return nil
}

// WatchPods watches for changes to pods in the specified namespace
func WatchPods(clientset kubernetes.Interface, namespace string) (watch.Interface, error) {
	watcher, err := clientset.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to watch pods in namespace %s: %v", namespace, err)
		return nil, err
	}
	return watcher, nil
}

// ListDeployments lists all deployments in the specified namespace
func ListDeployments(clientset kubernetes.Interface, namespace string) ([]appsv1.Deployment, error) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list deployments in namespace %s: %v", namespace, err)
		return nil, err
	}
	return deployments.Items, nil
}

// CreateDeployment creates a new deployment in the specified namespace
func CreateDeployment(clientset kubernetes.Interface, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	createdDeployment, err := clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create deployment %s in namespace %s: %v", deployment.Name, namespace, err)
		return nil, err
	}
	return createdDeployment, nil
}

// UpdateDeployment updates an existing deployment in the specified namespace
func UpdateDeployment(clientset kubernetes.Interface, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	updatedDeployment, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update deployment %s in namespace %s: %v", deployment.Name, namespace, err)
		return nil, err
	}
	return updatedDeployment, nil
}

// DeleteDeployment deletes a deployment in the specified namespace
func DeleteDeployment(clientset kubernetes.Interface, namespace, name string) error {
	err := clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete deployment %s in namespace %s: %v", name, namespace, err)
		return err
	}
	return nil
}

// ListServices lists all services in the specified namespace
func ListServices(clientset kubernetes.Interface, namespace string) ([]v1.Service, error) {
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list services in namespace %s: %v", namespace, err)
		return nil, err
	}
	return services.Items, nil
}

// CreateService creates a new service in the specified namespace
func CreateService(clientset kubernetes.Interface, namespace string, service *v1.Service) (*v1.Service, error) {
	createdService, err := clientset.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create service %s in namespace %s: %v", service.Name, namespace, err)
		return nil, err
	}
	return createdService, nil
}

// UpdateService updates an existing service in the specified namespace
func UpdateService(clientset kubernetes.Interface, namespace string, service *v1.Service) (*v1.Service, error) {
	updatedService, err := clientset.CoreV1().Services(namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update service %s in namespace %s: %v", service.Name, namespace, err)
		return nil, err
	}
	return updatedService, nil
}

// DeleteService deletes a service in the specified namespace
func DeleteService(clientset kubernetes.Interface, namespace, name string) error {
	err := clientset.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete service %s in namespace %s: %v", name, namespace, err)
		return err
	}
	return nil
}

// ListConfigMaps lists all configmaps in the specified namespace
func ListConfigMaps(clientset kubernetes.Interface, namespace string) ([]v1.ConfigMap, error) {
	configmaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list configmaps in namespace %s: %v", namespace, err)
		return nil, err
	}
	return configmaps.Items, nil
}

// CreateConfigMap creates a new configmap in the specified namespace
func CreateConfigMap(clientset kubernetes.Interface, namespace string, configmap *v1.ConfigMap) (*v1.ConfigMap, error) {
	createdConfigMap, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configmap, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create configmap %s in namespace %s: %v", configmap.Name, namespace, err)
		return nil, err
	}
	return createdConfigMap, nil
}

// UpdateConfigMap updates an existing configmap in the specified namespace
func UpdateConfigMap(clientset kubernetes.Interface, namespace string, configmap *v1.ConfigMap) (*v1.ConfigMap, error) {
	updatedConfigMap, err := clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configmap, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update configmap %s in namespace %s: %v", configmap.Name, namespace, err)
		return nil, err
	}
	return updatedConfigMap, nil
}

// DeleteConfigMap deletes a configmap in the specified namespace
func DeleteConfigMap(clientset kubernetes.Interface, namespace, name string) error {
	err := clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete configmap %s in namespace %s: %v", name, namespace, err)
		return err
	}
	return nil
}

// GetPodLogs retrieves logs from a pod
func GetPodLogs(clientset kubernetes.Interface, namespace, podName, containerName string, follow bool, tailLines int64) (io.ReadCloser, error) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Container: containerName,
		Follow:    follow,
		TailLines: &tailLines,
	})

	return req.Stream(context.TODO())
}

// ExecPod executes a command in a pod container
func ExecPod(clientset kubernetes.Interface, config *rest.Config, namespace, podName, containerName string, command []string) error {
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})
}

// ApplyYaml applies a YAML file to the cluster
func ApplyYaml(clientset kubernetes.Interface, namespace string, yamlFile string) error {
	// Decode YAML file
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yamlFile), nil, nil)
	if err != nil {
		return err
	}

	// Switch on the type of the object
	switch obj := obj.(type) {
	case *v1.Pod:
		_, err = CreatePod(clientset, namespace, obj)
	case *appsv1.Deployment:
		_, err = CreateDeployment(clientset, namespace, obj)
	case *v1.Service:
		_, err = CreateService(clientset, namespace, obj)
	case *v1.ConfigMap:
		_, err = CreateConfigMap(clientset, namespace, obj)
	default:
		return fmt.Errorf("unsupported object type %T", obj)
	}

	return err
}

// DeleteYaml deletes a resource defined in a YAML file from the cluster
func DeleteYaml(clientset kubernetes.Interface, namespace string, yamlFile string) error {
	// Decode YAML file
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yamlFile), nil, nil)
	if err != nil {
		return err
	}

	// Switch on the type of the object
	switch obj := obj.(type) {
	case *v1.Pod:
		err = DeletePod(clientset, namespace, obj.Name)
	case *appsv1.Deployment:
		err = DeleteDeployment(clientset, namespace, obj.Name)
	case *v1.Service:
		err = DeleteService(clientset, namespace, obj.Name)
	case *v1.ConfigMap:
		err = DeleteConfigMap(clientset, namespace, obj.Name)
	default:
		return fmt.Errorf("unsupported object type %T", obj)
	}

	return err
}

// UpdateFromYaml updates a resource defined in a YAML file in the cluster
func UpdateFromYaml(clientset kubernetes.Interface, namespace string, yamlFile string) error {
	// Decode YAML file
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yamlFile), nil, nil)
	if err != nil {
		return err
	}

	// Switch on the type of the object
	switch obj := obj.(type) {
	case *v1.Pod:
		_, err = UpdatePod(clientset, namespace, obj)
	case *appsv1.Deployment:
		_, err = UpdateDeployment(clientset, namespace, obj)
	case *v1.Service:
		_, err = UpdateService(clientset, namespace, obj)
	case *v1.ConfigMap:
		_, err = UpdateConfigMap(clientset, namespace, obj)
	default:
		return fmt.Errorf("unsupported object type %T", obj)
	}

	return err
}

// RetryOnConflict retries the operation in case of a conflict error
func RetryOnConflict(clientset kubernetes.Interface, namespace string, obj runtime.Object, updateFunc func() error) error {
	return retry.OnError(retry.DefaultRetry, func(err error) bool {
		// Check if the error is a conflict error
		return errors.IsConflict(err)
	}, func() error {
		// Retry the update operation
		return updateFunc()
	})
}

// ListNamespaces lists all namespaces in the cluster
func ListNamespaces(clientset kubernetes.Interface) ([]v1.Namespace, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list namespaces: %v", err)
		return nil, err
	}
	return namespaces.Items, nil
}
