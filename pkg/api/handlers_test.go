package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewHandler(t *testing.T) {
	// Create a fake clientset for testing
	fakeClientset := fake.NewSimpleClientset()

	handler := NewHandler(fakeClientset)

	if handler == nil {
		t.Error("NewHandler returned nil")
	}

	if handler.clientset == nil {
		t.Error("Handler clientset is nil")
	}
}

func TestListPods(t *testing.T) {
	// Create a fake pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "test-container",
					Image: "nginx",
				},
			},
		},
	}

	// Create fake clientset with the pod
	fakeClientset := fake.NewSimpleClientset(pod)
	handler := NewHandler(fakeClientset)

	// Create a gin router
	r := gin.Default()
	r.GET("/pods", handler.ListPods)

	// Create a test request
	req, _ := http.NewRequest("GET", "/pods?namespace=default", nil)
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse the response
	var response map[string][]v1.Pod
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	pods, ok := response["pods"]
	if !ok {
		t.Error("Response does not contain 'pods' key")
	}

	if len(pods) != 1 {
		t.Errorf("Expected 1 pod, got %d", len(pods))
	}

	if pods[0].Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got '%s'", pods[0].Name)
	}
}

func TestCreatePod(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	handler := NewHandler(fakeClientset)

	r := gin.Default()
	r.POST("/pods/:namespace", handler.CreatePod)

	podSpec := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-pod",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "test-container",
					Image: "nginx",
				},
			},
		},
	}

	podJSON, _ := json.Marshal(podSpec)
	req, _ := http.NewRequest("POST", "/pods/default", bytes.NewBuffer(podJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var createdPod v1.Pod
	if err := json.Unmarshal(w.Body.Bytes(), &createdPod); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if createdPod.Name != "new-pod" {
		t.Errorf("Expected pod name 'new-pod', got '%s'", createdPod.Name)
	}
}
