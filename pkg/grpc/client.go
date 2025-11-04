package grpc

import (
	"context"
	"time"

	"k8s-dashboard/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
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
func (c *Client) ListDeployments(namespace string) ([]*proto.Deployment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListDeployments(ctx, &proto.ListRequest{Namespace: namespace})
	if err != nil {
		klog.Errorf("Failed to list deployments via gRPC: %v", err)
		return nil, err
	}

	return resp.Deployments, nil
}

// ListServices lists services in the specified namespace
func (c *Client) ListServices(namespace string) ([]*proto.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListServices(ctx, &proto.ListRequest{Namespace: namespace})
	if err != nil {
		klog.Errorf("Failed to list services via gRPC: %v", err)
		return nil, err
	}

	return resp.Services, nil
}

// ListConfigMaps lists configmaps in the specified namespace
func (c *Client) ListConfigMaps(namespace string) ([]*proto.ConfigMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.ListConfigMaps(ctx, &proto.ListRequest{Namespace: namespace})
	if err != nil {
		klog.Errorf("Failed to list configmaps via gRPC: %v", err)
		return nil, err
	}

	return resp.Configmaps, nil
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
