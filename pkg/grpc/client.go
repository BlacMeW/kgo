package grpc

import (
	"context"
	"time"

	"k8s-dashboard/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// Client wraps the gRPC client for Kubernetes operations
type Client struct {
	conn   *grpc.ClientConn
	client proto.K8SServiceClient
}

// NewClient creates a new gRPC client
func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, err
	}

	client := proto.NewK8SServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// ListPods lists pods in the specified namespace
func (c *Client) ListPods(namespace string) ([]v1.Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListPods(ctx, &proto.ListRequest{Namespace: namespace})
	if err != nil {
		klog.Errorf("Failed to list pods via gRPC: %v", err)
		return nil, err
	}

	var pods []v1.Pod
	for _, protoPod := range resp.Pods {
		pod := c.convertProtoToPod(protoPod)
		pods = append(pods, *pod)
	}

	return pods, nil
}

// ListDeployments lists deployments in the specified namespace
func (c *Client) ListDeployments(namespace string) ([]appsv1.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListDeployments(ctx, &proto.ListRequest{Namespace: namespace})
	if err != nil {
		klog.Errorf("Failed to list deployments via gRPC: %v", err)
		return nil, err
	}

	var deployments []appsv1.Deployment
	for _, protoDep := range resp.Deployments {
		dep := c.convertProtoToDeployment(protoDep)
		deployments = append(deployments, *dep)
	}

	return deployments, nil
}

// ListServices lists services in the specified namespace
func (c *Client) ListServices(namespace string) ([]v1.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListServices(ctx, &proto.ListRequest{Namespace: namespace})
	if err != nil {
		klog.Errorf("Failed to list services via gRPC: %v", err)
		return nil, err
	}

	var services []v1.Service
	for _, protoSvc := range resp.Services {
		svc := c.convertProtoToService(protoSvc)
		services = append(services, *svc)
	}

	return services, nil
}

// ListConfigMaps lists configmaps in the specified namespace
func (c *Client) ListConfigMaps(namespace string) ([]v1.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListConfigMaps(ctx, &proto.ListRequest{Namespace: namespace})
	if err != nil {
		klog.Errorf("Failed to list configmaps via gRPC: %v", err)
		return nil, err
	}

	var configMaps []v1.ConfigMap
	for _, protoCm := range resp.Configmaps {
		cm := c.convertProtoToConfigMap(protoCm)
		configMaps = append(configMaps, *cm)
	}

	return configMaps, nil
}

// ListNamespaces lists all namespaces
func (c *Client) ListNamespaces() ([]*proto.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListNamespaces(ctx, &emptypb.Empty{})
	if err != nil {
		klog.Errorf("Failed to list namespaces via gRPC: %v", err)
		return nil, err
	}

	return resp.Namespaces, nil
}

// CreatePod creates a new pod
func (c *Client) CreatePod(namespace string, spec *proto.PodSpec) (*proto.Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.CreatePod(ctx, &proto.CreatePodRequest{
		Namespace: namespace,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to create pod via gRPC: %v", err)
		return nil, err
	}

	return resp.Pod, nil
}

// UpdatePod updates an existing pod
func (c *Client) UpdatePod(namespace, name string, spec *proto.PodSpec) (*proto.Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdatePod(ctx, &proto.UpdatePodRequest{
		Namespace: namespace,
		Name:      name,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to update pod via gRPC: %v", err)
		return nil, err
	}

	return resp.Pod, nil
}

// DeletePod deletes a pod
func (c *Client) DeletePod(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.client.DeletePod(ctx, &proto.DeleteRequest{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil {
		klog.Errorf("Failed to delete pod via gRPC: %v", err)
		return err
	}

	return nil
}

// CreateDeployment creates a new deployment
func (c *Client) CreateDeployment(namespace string, spec *proto.DeploymentSpec) (*proto.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.CreateDeployment(ctx, &proto.CreateDeploymentRequest{
		Namespace: namespace,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to create deployment via gRPC: %v", err)
		return nil, err
	}

	return resp.Deployment, nil
}

// UpdateDeployment updates an existing deployment
func (c *Client) UpdateDeployment(namespace, name string, spec *proto.DeploymentSpec) (*proto.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdateDeployment(ctx, &proto.UpdateDeploymentRequest{
		Namespace: namespace,
		Name:      name,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to update deployment via gRPC: %v", err)
		return nil, err
	}

	return resp.Deployment, nil
}

// DeleteDeployment deletes a deployment
func (c *Client) DeleteDeployment(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.client.DeleteDeployment(ctx, &proto.DeleteRequest{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil {
		klog.Errorf("Failed to delete deployment via gRPC: %v", err)
		return err
	}

	return nil
}

// CreateService creates a new service
func (c *Client) CreateService(namespace string, spec *proto.ServiceSpec) (*proto.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.CreateService(ctx, &proto.CreateServiceRequest{
		Namespace: namespace,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to create service via gRPC: %v", err)
		return nil, err
	}

	return resp.Service, nil
}

// UpdateService updates an existing service
func (c *Client) UpdateService(namespace, name string, spec *proto.ServiceSpec) (*proto.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdateService(ctx, &proto.UpdateServiceRequest{
		Namespace: namespace,
		Name:      name,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to update service via gRPC: %v", err)
		return nil, err
	}

	return resp.Service, nil
}

// DeleteService deletes a service
func (c *Client) DeleteService(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.client.DeleteService(ctx, &proto.DeleteRequest{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil {
		klog.Errorf("Failed to delete service via gRPC: %v", err)
		return err
	}

	return nil
}

// CreateConfigMap creates a new configmap
func (c *Client) CreateConfigMap(namespace string, spec *proto.ConfigMapSpec) (*proto.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.CreateConfigMap(ctx, &proto.CreateConfigMapRequest{
		Namespace: namespace,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to create configmap via gRPC: %v", err)
		return nil, err
	}

	return resp.Configmap, nil
}

// UpdateConfigMap updates an existing configmap
func (c *Client) UpdateConfigMap(namespace, name string, spec *proto.ConfigMapSpec) (*proto.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdateConfigMap(ctx, &proto.UpdateConfigMapRequest{
		Namespace: namespace,
		Name:      name,
		Spec:      spec,
	})
	if err != nil {
		klog.Errorf("Failed to update configmap via gRPC: %v", err)
		return nil, err
	}

	return resp.Configmap, nil
}

// DeleteConfigMap deletes a configmap
func (c *Client) DeleteConfigMap(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.client.DeleteConfigMap(ctx, &proto.DeleteRequest{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil {
		klog.Errorf("Failed to delete configmap via gRPC: %v", err)
		return err
	}

	return nil
}

// GetPodLogs retrieves logs from a pod
func (c *Client) GetPodLogs(namespace, podName, containerName string, tailLines int32, follow bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetPodLogs(ctx, &proto.PodLogsRequest{
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		TailLines:     tailLines,
		Follow:        follow,
	})
	if err != nil {
		klog.Errorf("Failed to get pod logs via gRPC: %v", err)
		return "", err
	}

	return resp.Logs, nil
}

// Conversion functions from protobuf to Kubernetes types

func (c *Client) convertProtoToPod(protoPod *proto.Pod) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      protoPod.Name,
			Namespace: protoPod.Namespace,
			Labels:    protoPod.Labels,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPhase(protoPod.Status),
		},
		Spec: v1.PodSpec{
			NodeName: protoPod.Node,
		},
	}

	// Convert containers
	for _, protoContainer := range protoPod.Containers {
		container := v1.Container{
			Name:  protoContainer.Name,
			Image: protoContainer.Image,
		}

		// Convert ports
		for _, protoPort := range protoContainer.Ports {
			port := v1.ContainerPort{
				ContainerPort: protoPort.ContainerPort,
				Protocol:      v1.Protocol(protoPort.Protocol),
			}
			container.Ports = append(container.Ports, port)
		}

		pod.Spec.Containers = append(pod.Spec.Containers, container)

		// Add container status
		status := v1.ContainerStatus{
			Name:  protoContainer.Name,
			Ready: protoContainer.Status == "Running",
		}
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, status)
	}

	return pod
}

func (c *Client) convertProtoToDeployment(protoDep *proto.Deployment) *appsv1.Deployment {
	replicas := protoDep.Replicas
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      protoDep.Name,
			Namespace: protoDep.Namespace,
			Labels:    protoDep.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     protoDep.ReadyReplicas,
			AvailableReplicas: protoDep.AvailableReplicas,
		},
	}
}

func (c *Client) convertProtoToService(protoSvc *proto.Service) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      protoSvc.Name,
			Namespace: protoSvc.Namespace,
			Labels:    protoSvc.Labels,
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceType(protoSvc.Type),
			ClusterIP: protoSvc.ClusterIp,
		},
	}

	// Convert ports (simple conversion from string format)
	for _, portStr := range protoSvc.Ports {
		// Parse "port/protocol" format
		if portStr != "" {
			// This is a simplified conversion - in a real implementation,
			// you'd want more robust parsing
			port := v1.ServicePort{
				Port:     80, // Default
				Protocol: v1.ProtocolTCP,
			}
			svc.Spec.Ports = append(svc.Spec.Ports, port)
		}
	}

	return svc
}

func (c *Client) convertProtoToConfigMap(protoCm *proto.ConfigMap) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      protoCm.Name,
			Namespace: protoCm.Namespace,
			Labels:    protoCm.Labels,
		},
		Data: protoCm.Data,
	}
}
