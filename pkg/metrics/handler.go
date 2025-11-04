package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// MetricsHandler struct holds the Kubernetes clientset
type MetricsHandler struct {
	clientset kubernetes.Interface
}

// NewMetricsHandler creates a new metrics API handler
func NewMetricsHandler(clientset kubernetes.Interface) *MetricsHandler {
	return &MetricsHandler{clientset: clientset}
}

// GetClusterMetrics returns basic cluster metrics
func (h *MetricsHandler) GetClusterMetrics(c *gin.Context) {
	// Get node count
	nodes, err := h.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get pod count across all namespaces
	pods, err := h.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list pods: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get namespace count
	namespaces, err := h.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list namespaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate pod status counts
	podStatus := map[string]int{
		"running":   0,
		"pending":   0,
		"failed":    0,
		"succeeded": 0,
		"unknown":   0,
	}

	for _, pod := range pods.Items {
		status := string(pod.Status.Phase)
		if count, exists := podStatus[status]; exists {
			podStatus[status] = count + 1
		} else {
			podStatus["unknown"]++
		}
	}

	metrics := gin.H{
		"cluster": gin.H{
			"nodes":      len(nodes.Items),
			"pods":       len(pods.Items),
			"namespaces": len(namespaces.Items),
		},
		"pod_status": podStatus,
		"timestamp":  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, metrics)
}

// GetNamespaceMetrics returns metrics for a specific namespace
func (h *MetricsHandler) GetNamespaceMetrics(c *gin.Context) {
	namespace := c.Param("namespace")

	// Get pods in namespace
	pods, err := h.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list pods in namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get deployments in namespace
	deployments, err := h.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list deployments in namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get services in namespace
	services, err := h.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list services in namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate deployment status
	deploymentStatus := map[string]int{
		"available":   0,
		"unavailable": 0,
		"updating":    0,
	}

	for _, deployment := range deployments.Items {
		if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
			deploymentStatus["available"]++
		} else if deployment.Status.ReadyReplicas > 0 {
			deploymentStatus["updating"]++
		} else {
			deploymentStatus["unavailable"]++
		}
	}

	metrics := gin.H{
		"namespace": namespace,
		"pods": gin.H{
			"total":     len(pods.Items),
			"running":   countPodsByPhase(pods.Items, v1.PodRunning),
			"pending":   countPodsByPhase(pods.Items, v1.PodPending),
			"failed":    countPodsByPhase(pods.Items, v1.PodFailed),
			"succeeded": countPodsByPhase(pods.Items, v1.PodSucceeded),
		},
		"deployments": gin.H{
			"total":  len(deployments.Items),
			"status": deploymentStatus,
		},
		"services": gin.H{
			"total": len(services.Items),
		},
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, metrics)
}

// countPodsByPhase counts pods by their phase
func countPodsByPhase(pods []v1.Pod, phase v1.PodPhase) int {
	count := 0
	for _, pod := range pods {
		if pod.Status.Phase == phase {
			count++
		}
	}
	return count
}
