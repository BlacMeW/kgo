package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetClusterMetrics(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create metrics handler
	handler := NewMetricsHandler(clientset)

	// Create a test request
	req, _ := http.NewRequest("GET", "/api/v1/metrics/cluster", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call the handler
	handler.GetClusterMetrics(c)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}
}

func TestGetNamespaceMetrics(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create metrics handler
	handler := NewMetricsHandler(clientset)

	// Create a test request
	req, _ := http.NewRequest("GET", "/api/v1/metrics/namespace/default", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "namespace", Value: "default"}}

	// Call the handler
	handler.GetNamespaceMetrics(c)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
