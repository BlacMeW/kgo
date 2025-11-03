package api

import (
	"net/http"

	"k8s-dashboard/pkg/k8s"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Handler struct holds the Kubernetes clientset
type Handler struct {
	clientset kubernetes.Interface
}

// NewHandler creates a new API handler with the given clientset
func NewHandler(clientset kubernetes.Interface) *Handler {
	return &Handler{clientset: clientset}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

// ListPods handles GET /api/v1/pods?namespace=default
func (h *Handler) ListPods(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	pods, err := k8s.ListPods(h.clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to list pods: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pods": pods})
}

// CreatePod handles POST /api/v1/pods/:namespace
func (h *Handler) CreatePod(c *gin.Context) {
	namespace := c.Param("namespace")

	var pod v1.Pod
	if err := c.ShouldBindJSON(&pod); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure namespace is set
	pod.Namespace = namespace

	createdPod, err := k8s.CreatePod(h.clientset, namespace, &pod)
	if err != nil {
		klog.Errorf("Failed to create pod: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdPod)
}

// UpdatePod handles PUT /api/v1/pods/:namespace/:name
func (h *Handler) UpdatePod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var pod v1.Pod
	if err := c.ShouldBindJSON(&pod); err != nil {
		klog.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Ensure name and namespace are set
	pod.Name = name
	pod.Namespace = namespace

	updatedPod, err := k8s.UpdatePod(h.clientset, namespace, &pod)
	if err != nil {
		klog.Errorf("Failed to update pod: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedPod)
}

// DeletePod handles DELETE /api/v1/pods/:namespace/:name
func (h *Handler) DeletePod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := k8s.DeletePod(h.clientset, namespace, name)
	if err != nil {
		klog.Errorf("Failed to delete pod: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pod deleted successfully"})
}

// WatchPods handles WebSocket connection for watching pod changes
func (h *Handler) WatchPods(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")

	watcher, err := k8s.WatchPods(h.clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to start watching pods: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Upgrade to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				klog.Info("Watcher channel closed")
				return
			}
			err := ws.WriteJSON(event)
			if err != nil {
				klog.Errorf("Failed to write to WebSocket: %v", err)
				return
			}
		}
	}
}
