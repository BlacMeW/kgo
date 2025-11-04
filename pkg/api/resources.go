package api

import (
	"net/http"

	"k8s-dashboard/pkg/k8s"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// ResourceHandler struct holds the Kubernetes clientset
type ResourceHandler struct {
	clientset kubernetes.Interface
}

// NewResourceHandler creates a new resource API handler
func NewResourceHandler(clientset kubernetes.Interface) *ResourceHandler {
	return &ResourceHandler{clientset: clientset}
}

// ListDeployments handles GET /api/v1/deployments?namespace=default
func (h *ResourceHandler) ListDeployments(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	deployments, err := k8s.ListDeployments(h.clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to list deployments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deployments": deployments})
}

// CreateDeployment handles POST /api/v1/deployments/:namespace
func (h *ResourceHandler) CreateDeployment(c *gin.Context) {
	namespace := c.Param("namespace")

	var deployment appsv1.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure namespace is set
	deployment.Namespace = namespace

	createdDeployment, err := k8s.CreateDeployment(h.clientset, namespace, &deployment)
	if err != nil {
		klog.Errorf("Failed to create deployment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdDeployment)
}

// UpdateDeployment handles PUT /api/v1/deployments/:namespace/:name
func (h *ResourceHandler) UpdateDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var deployment appsv1.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure name and namespace are set
	deployment.Name = name
	deployment.Namespace = namespace

	updatedDeployment, err := k8s.UpdateDeployment(h.clientset, namespace, &deployment)
	if err != nil {
		klog.Errorf("Failed to update deployment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedDeployment)
}

// DeleteDeployment handles DELETE /api/v1/deployments/:namespace/:name
func (h *ResourceHandler) DeleteDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := k8s.DeleteDeployment(h.clientset, namespace, name)
	if err != nil {
		klog.Errorf("Failed to delete deployment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deployment deleted successfully"})
}

// ListServices handles GET /api/v1/services?namespace=default
func (h *ResourceHandler) ListServices(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	services, err := k8s.ListServices(h.clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to list services: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"services": services})
}

// CreateService handles POST /api/v1/services/:namespace
func (h *ResourceHandler) CreateService(c *gin.Context) {
	namespace := c.Param("namespace")

	var service v1.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure namespace is set
	service.Namespace = namespace

	createdService, err := k8s.CreateService(h.clientset, namespace, &service)
	if err != nil {
		klog.Errorf("Failed to create service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdService)
}

// UpdateService handles PUT /api/v1/services/:namespace/:name
func (h *ResourceHandler) UpdateService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var service v1.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure name and namespace are set
	service.Name = name
	service.Namespace = namespace

	updatedService, err := k8s.UpdateService(h.clientset, namespace, &service)
	if err != nil {
		klog.Errorf("Failed to update service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedService)
}

// DeleteService handles DELETE /api/v1/services/:namespace/:name
func (h *ResourceHandler) DeleteService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := k8s.DeleteService(h.clientset, namespace, name)
	if err != nil {
		klog.Errorf("Failed to delete service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}

// ListConfigMaps handles GET /api/v1/configmaps?namespace=default
func (h *ResourceHandler) ListConfigMaps(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	configmaps, err := k8s.ListConfigMaps(h.clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to list configmaps: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configmaps": configmaps})
}

// CreateConfigMap handles POST /api/v1/configmaps/:namespace
func (h *ResourceHandler) CreateConfigMap(c *gin.Context) {
	namespace := c.Param("namespace")

	var configmap v1.ConfigMap
	if err := c.ShouldBindJSON(&configmap); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure namespace is set
	configmap.Namespace = namespace

	createdConfigMap, err := k8s.CreateConfigMap(h.clientset, namespace, &configmap)
	if err != nil {
		klog.Errorf("Failed to create configmap: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdConfigMap)
}

// UpdateConfigMap handles PUT /api/v1/configmaps/:namespace/:name
func (h *ResourceHandler) UpdateConfigMap(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var configmap v1.ConfigMap
	if err := c.ShouldBindJSON(&configmap); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure name and namespace are set
	configmap.Name = name
	configmap.Namespace = namespace

	updatedConfigMap, err := k8s.UpdateConfigMap(h.clientset, namespace, &configmap)
	if err != nil {
		klog.Errorf("Failed to update configmap: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedConfigMap)
}

// DeleteConfigMap handles DELETE /api/v1/configmaps/:namespace/:name
func (h *ResourceHandler) DeleteConfigMap(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := k8s.DeleteConfigMap(h.clientset, namespace, name)
	if err != nil {
		klog.Errorf("Failed to delete configmap: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ConfigMap deleted successfully"})
}

// GetPodLogs handles GET /api/v1/pods/:namespace/:name/logs
func (h *ResourceHandler) GetPodLogs(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	container := c.DefaultQuery("container", "")
	follow := c.DefaultQuery("follow", "false") == "true"
	tailLines := int64(100)

	logStream, err := k8s.GetPodLogs(h.clientset, namespace, name, container, follow, tailLines)
	if err != nil {
		klog.Errorf("Failed to get pod logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer logStream.Close()

	// Set headers for streaming
	c.Header("Content-Type", "text/plain")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Stream the logs
	buf := make([]byte, 4096)
	for {
		n, err := logStream.Read(buf)
		if n > 0 {
			c.Writer.Write(buf[:n])
			c.Writer.Flush()
		}
		if err != nil {
			break
		}
	}
}

// ExecPod handles WebSocket connection for pod exec
func (h *ResourceHandler) ExecPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	container := c.DefaultQuery("container", "")
	command := c.DefaultQuery("command", "/bin/sh")

	// Upgrade to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	// Create a command array
	cmd := []string{"/bin/sh", "-c", command}
	if command == "/bin/sh" || command == "/bin/bash" {
		cmd = []string{command}
	}

	// Get the REST config for exec
	config, err := rest.InClusterConfig()
	if err != nil {
		// Try to get config from client
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			klog.Errorf("Failed to get config for exec: %v", err)
			ws.WriteJSON(gin.H{"error": "Failed to get cluster config"})
			return
		}
	}

	// Start exec session
	err = k8s.ExecPod(h.clientset, config, namespace, name, container, cmd)
	if err != nil {
		klog.Errorf("Failed to exec pod: %v", err)
		ws.WriteJSON(gin.H{"error": err.Error()})
		return
	}

	// Send completion message
	ws.WriteJSON(gin.H{"status": "completed"})
}
