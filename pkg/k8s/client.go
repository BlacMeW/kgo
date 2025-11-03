package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
