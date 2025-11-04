package grpc

import (
	"testing"

	"k8s-dashboard/proto"

	v1 "k8s.io/api/core/v1"
)

func TestConvertProtoToPod(t *testing.T) {
	client := &Client{}

	protoPod := &proto.Pod{
		Name:      "test-pod",
		Namespace: "default",
		Status:    "Running",
		Node:      "node-1",
		Labels:    map[string]string{"app": "test"},
		Containers: []*proto.Container{
			{
				Name:   "nginx",
				Image:  "nginx:1.20",
				Status: "Running",
				Ports: []*proto.Port{
					{
						Protocol:      "TCP",
						ContainerPort: 80,
					},
				},
			},
		},
	}

	pod := client.convertProtoToPod(protoPod)

	if pod.Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got '%s'", pod.Name)
	}
	if pod.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", pod.Namespace)
	}
	if pod.Status.Phase != v1.PodRunning {
		t.Errorf("Expected status 'Running', got '%s'", pod.Status.Phase)
	}
	if pod.Spec.NodeName != "node-1" {
		t.Errorf("Expected node 'node-1', got '%s'", pod.Spec.NodeName)
	}
	if len(pod.Spec.Containers) != 1 {
		t.Errorf("Expected 1 container, got %d", len(pod.Spec.Containers))
	}
	if pod.Spec.Containers[0].Image != "nginx:1.20" {
		t.Errorf("Expected container image 'nginx:1.20', got '%s'", pod.Spec.Containers[0].Image)
	}
	if len(pod.Spec.Containers[0].Ports) != 1 {
		t.Errorf("Expected 1 port, got %d", len(pod.Spec.Containers[0].Ports))
	}
	if pod.Spec.Containers[0].Ports[0].ContainerPort != 80 {
		t.Errorf("Expected port 80, got %d", pod.Spec.Containers[0].Ports[0].ContainerPort)
	}
}

func TestConvertProtoToDeployment(t *testing.T) {
	client := &Client{}

	protoDep := &proto.Deployment{
		Name:              "test-deployment",
		Namespace:         "default",
		Replicas:          3,
		ReadyReplicas:     3,
		AvailableReplicas: 3,
		Labels:            map[string]string{"app": "test"},
	}

	dep := client.convertProtoToDeployment(protoDep)

	if dep.Name != "test-deployment" {
		t.Errorf("Expected deployment name 'test-deployment', got '%s'", dep.Name)
	}
	if dep.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", dep.Namespace)
	}
	if *dep.Spec.Replicas != 3 {
		t.Errorf("Expected replicas 3, got %d", *dep.Spec.Replicas)
	}
	if dep.Status.ReadyReplicas != 3 {
		t.Errorf("Expected ready replicas 3, got %d", dep.Status.ReadyReplicas)
	}
}

func TestConvertProtoToService(t *testing.T) {
	client := &Client{}

	protoSvc := &proto.Service{
		Name:      "test-service",
		Namespace: "default",
		Type:      "ClusterIP",
		ClusterIp: "10.0.0.1",
		Labels:    map[string]string{"app": "test"},
		Ports:     []string{"80/TCP"},
	}

	svc := client.convertProtoToService(protoSvc)

	if svc.Name != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", svc.Name)
	}
	if svc.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", svc.Namespace)
	}
	if svc.Spec.Type != v1.ServiceTypeClusterIP {
		t.Errorf("Expected type 'ClusterIP', got '%s'", svc.Spec.Type)
	}
	if svc.Spec.ClusterIP != "10.0.0.1" {
		t.Errorf("Expected cluster IP '10.0.0.1', got '%s'", svc.Spec.ClusterIP)
	}
}

func TestConvertProtoToConfigMap(t *testing.T) {
	client := &Client{}

	protoCm := &proto.ConfigMap{
		Name:      "test-configmap",
		Namespace: "default",
		Data:      map[string]string{"key": "value"},
		Labels:    map[string]string{"app": "test"},
	}

	cm := client.convertProtoToConfigMap(protoCm)

	if cm.Name != "test-configmap" {
		t.Errorf("Expected configmap name 'test-configmap', got '%s'", cm.Name)
	}
	if cm.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", cm.Namespace)
	}
	if cm.Data["key"] != "value" {
		t.Errorf("Expected data 'key=value', got '%s'", cm.Data["key"])
	}
}
