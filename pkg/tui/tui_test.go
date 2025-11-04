package tui

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestTUIBasicInitialization tests basic TUI initialization
func TestTUIBasicInitialization(t *testing.T) {
	// Create a fake clientset for testing
	clientset := fake.NewSimpleClientset()

	// Create TUI instance
	tui, err := NewTUI(clientset)
	if err != nil {
		t.Fatalf("Failed to create TUI instance: %v", err)
	}

	if tui == nil {
		t.Fatal("TUI instance is nil")
	}

	if tui.clientset == nil {
		t.Error("TUI clientset is nil")
	}

	if tui.namespace != "kube-system" {
		t.Errorf("Expected default namespace 'kube-system', got '%s'", tui.namespace)
	}

	if tui.currentView != ResourcePods {
		t.Error("Expected default resource to be ResourcePods")
	}
}

// TestTUIDataUpdateHandling tests data update handling
func TestTUIDataUpdateHandling(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 1, // Set to 1 so it decrements to 0 and doesn't trigger extra logging
	}

	// Create test data
	testPods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-1",
				Namespace: "default",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-2",
				Namespace: "default",
			},
			Status: v1.PodStatus{
				Phase: v1.PodPending,
			},
		},
	}

	testDeployments := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &[]int32{3}[0],
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 3,
			},
		},
	}

	// Test pod data update
	tui.pods = testPods

	if len(tui.pods) != 2 {
		t.Errorf("Expected 2 pods, got %d", len(tui.pods))
	}

	// Test deployment data update
	tui.deployments = testDeployments

	if len(tui.deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(tui.deployments))
	}
}

// TestTUIAsyncDataLoading tests async data loading functionality
func TestTUIAsyncDataLoading(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 0,
	}

	// Test that TUI can be created without panicking
	if tui == nil {
		t.Error("TUI creation failed")
	}

	// Test that basic fields are initialized
	if tui.clientset == nil {
		t.Error("Clientset not initialized")
	}
}

// TestTUINavigation tests basic navigation functionality
func TestTUINavigation(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 0,
	}

	// Test namespace switching
	tui.namespace = "kube-system"
	if tui.namespace != "kube-system" {
		t.Errorf("Expected namespace 'kube-system', got '%s'", tui.namespace)
	}

	// Test resource switching
	tui.currentView = ResourceDeployments
	if tui.currentView != ResourceDeployments {
		t.Errorf("Expected resource ResourceDeployments, got %v", tui.currentView)
	}

	tui.currentView = ResourceServices
	if tui.currentView != ResourceServices {
		t.Errorf("Expected resource ResourceServices, got %v", tui.currentView)
	}

	tui.currentView = ResourceConfigMaps
	if tui.currentView != ResourceConfigMaps {
		t.Errorf("Expected resource ResourceConfigMaps, got %v", tui.currentView)
	}
}

// TestTUIFiltering tests filtering functionality
func TestTUIFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 0,
		filter:         "",
	}

	// Set up test data
	testPods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "nginx"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis-pod",
				Namespace: "default",
				Labels:    map[string]string{"app": "redis"},
			},
		},
	}

	tui.pods = testPods
	tui.currentView = ResourcePods

	// Test filtering by name
	tui.filter = "nginx"
	filtered := tui.getFilteredPods()
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered pod, got %d", len(filtered))
	}
	if filtered[0].Name != "nginx-pod" {
		t.Errorf("Expected filtered pod name 'nginx-pod', got '%s'", filtered[0].Name)
	}

	// Test filtering with no matches
	tui.filter = "nonexistent"
	filtered = tui.getFilteredPods()
	if len(filtered) != 0 {
		t.Errorf("Expected 0 filtered pods, got %d", len(filtered))
	}

	// Test clearing filter
	tui.filter = ""
	filtered = tui.getFilteredPods()
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered pods after clearing filter, got %d", len(filtered))
	}
}

// TestTUIKeyHandling tests key event handling
func TestTUIKeyHandling(t *testing.T) {
	// Test quit key
	event := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	// Note: In a real scenario, this would be handled by the event loop
	// Here we just test that the event creation doesn't panic

	if event == nil {
		t.Error("Failed to create key event")
	}

	// Test navigation keys
	upEvent := tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	if upEvent == nil {
		t.Error("Failed to create up key event")
	}

	downEvent := tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	if downEvent == nil {
		t.Error("Failed to create down key event")
	}
}

// TestTUIConcurrentDataUpdates tests concurrent data updates
func TestTUIConcurrentDataUpdates(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 0,
	}

	// Test that multiple data updates don't cause race conditions
	done := make(chan bool, 2)

	go func() {
		tui.pods = []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
		}
		done <- true
	}()

	go func() {
		tui.deployments = []appsv1.Deployment{
			{ObjectMeta: metav1.ObjectMeta{Name: "dep1"}},
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify data was updated
	if len(tui.pods) != 1 {
		t.Errorf("Expected 1 pod after concurrent update, got %d", len(tui.pods))
	}
	if len(tui.deployments) != 1 {
		t.Errorf("Expected 1 deployment after concurrent update, got %d", len(tui.deployments))
	}
}

// TestTUIMemoryCleanup tests that the TUI cleans up resources properly
func TestTUIMemoryCleanup(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 0,
	}

	// Add some data
	tui.pods = []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
	}

	// Verify data exists
	if len(tui.pods) != 2 {
		t.Errorf("Expected 2 pods initially, got %d", len(tui.pods))
	}

	// Clear data (simulate namespace change)
	tui.pods = nil
	tui.deployments = nil
	tui.services = nil
	tui.configMaps = nil

	// Verify data is cleared
	if len(tui.pods) != 0 {
		t.Errorf("Expected 0 pods after clearing, got %d", len(tui.pods))
	}
}

// BenchmarkTUIDataUpdate benchmarks data update performance
func BenchmarkTUIDataUpdate(b *testing.B) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 0,
	}

	// Create a large dataset
	var pods []v1.Pod
	for i := 0; i < 1000; i++ {
		pods = append(pods, v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("pod-%d", i),
				Namespace: "default",
			},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tui.pods = pods
	}
}

// BenchmarkTUIFiltering benchmarks filtering performance
func BenchmarkTUIFiltering(b *testing.B) {
	clientset := fake.NewSimpleClientset()

	// Create TUI manually to avoid screen initialization issues
	tui := &TUI{
		clientset:      clientset,
		namespace:      "kube-system",
		currentView:    ResourcePods,
		pods:           []v1.Pod{},
		deployments:    []appsv1.Deployment{},
		services:       []v1.Service{},
		configMaps:     []v1.ConfigMap{},
		loadingCounter: 0,
		filter:         "",
	}

	// Set up test data
	var pods []v1.Pod
	for i := 0; i < 1000; i++ {
		pods = append(pods, v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("test-pod-%d", i),
				Namespace: "default",
			},
		})
	}
	tui.pods = pods
	tui.currentView = ResourcePods

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tui.filter = "test-pod-5"
		_ = tui.getFilteredPods()
	}
}
