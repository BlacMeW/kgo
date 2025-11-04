package grpc

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"k8s-dashboard/pkg/k8s"
	"k8s-dashboard/proto"

	"google.golang.org/protobuf/types/known/emptypb"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Server implements the gRPC K8sService
type Server struct {
	proto.UnimplementedK8SServiceServer
	clientset *kubernetes.Clientset
}

// calculateAge calculates the age of a resource from its creation timestamp
func calculateAge(creationTime metav1.Time) string {
	duration := time.Since(creationTime.Time)

	if duration.Hours() > 24 {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	} else if duration.Hours() >= 1 {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh", hours)
	} else if duration.Minutes() >= 1 {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm", minutes)
	} else {
		seconds := int(duration.Seconds())
		return fmt.Sprintf("%ds", seconds)
	}
}

// NewServer creates a new gRPC server instance
func NewServer(clientset *kubernetes.Clientset) *Server {
	return &Server{
		clientset: clientset,
	}
}

// ListPods lists pods in the specified namespace
func (s *Server) ListPods(ctx context.Context, req *proto.ListRequest) (*proto.PodListResponse, error) {
	pods, err := k8s.ListPods(s.clientset, req.Namespace)
	if err != nil {
		klog.Errorf("Failed to list pods: %v", err)
		return nil, err
	}

	var protoPods []*proto.Pod
	for _, pod := range pods {
		protoPod := s.convertPodToProto(&pod)
		protoPods = append(protoPods, protoPod)
	}

	return &proto.PodListResponse{Pods: protoPods}, nil
}

// ListDeployments lists deployments in the specified namespace
func (s *Server) ListDeployments(ctx context.Context, req *proto.ListRequest) (*proto.DeploymentListResponse, error) {
	deployments, err := k8s.ListDeployments(s.clientset, req.Namespace)
	if err != nil {
		klog.Errorf("Failed to list deployments: %v", err)
		return nil, err
	}

	var protoDeployments []*proto.Deployment
	for _, dep := range deployments {
		protoDep := s.convertDeploymentToProto(&dep)
		protoDeployments = append(protoDeployments, protoDep)
	}

	return &proto.DeploymentListResponse{Deployments: protoDeployments}, nil
}

// ListServices lists services in the specified namespace
func (s *Server) ListServices(ctx context.Context, req *proto.ListRequest) (*proto.ServiceListResponse, error) {
	services, err := k8s.ListServices(s.clientset, req.Namespace)
	if err != nil {
		klog.Errorf("Failed to list services: %v", err)
		return nil, err
	}

	var protoServices []*proto.Service
	for _, svc := range services {
		protoSvc := s.convertServiceToProto(&svc)
		protoServices = append(protoServices, protoSvc)
	}

	return &proto.ServiceListResponse{Services: protoServices}, nil
}

// ListConfigMaps lists configmaps in the specified namespace
func (s *Server) ListConfigMaps(ctx context.Context, req *proto.ListRequest) (*proto.ConfigMapListResponse, error) {
	configmaps, err := k8s.ListConfigMaps(s.clientset, req.Namespace)
	if err != nil {
		klog.Errorf("Failed to list configmaps: %v", err)
		return nil, err
	}

	var protoConfigMaps []*proto.ConfigMap
	for _, cm := range configmaps {
		protoCm := s.convertConfigMapToProto(&cm)
		protoConfigMaps = append(protoConfigMaps, protoCm)
	}

	return &proto.ConfigMapListResponse{Configmaps: protoConfigMaps}, nil
}

// ListNamespaces lists all namespaces
func (s *Server) ListNamespaces(ctx context.Context, req *emptypb.Empty) (*proto.NamespaceListResponse, error) {
	namespaces, err := s.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list namespaces: %v", err)
		return nil, err
	}

	var protoNamespaces []*proto.Namespace
	for _, ns := range namespaces.Items {
		protoNs := &proto.Namespace{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Age:    calculateAge(ns.CreationTimestamp),
		}
		protoNamespaces = append(protoNamespaces, protoNs)
	}

	return &proto.NamespaceListResponse{Namespaces: protoNamespaces}, nil
}

// CreatePod creates a new pod
func (s *Server) CreatePod(ctx context.Context, req *proto.CreatePodRequest) (*proto.PodResponse, error) {
	// Convert proto spec to Kubernetes pod spec
	podSpec := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Spec.Name,
			Namespace: req.Namespace,
			Labels:    req.Spec.Labels,
		},
		Spec: v1.PodSpec{},
	}

	// Add containers
	for _, containerSpec := range req.Spec.Containers {
		container := v1.Container{
			Name:  containerSpec.Name,
			Image: containerSpec.Image,
		}

		// Add ports
		for _, portSpec := range containerSpec.Ports {
			container.Ports = append(container.Ports, v1.ContainerPort{
				ContainerPort: int32(portSpec.ContainerPort),
				Protocol:      v1.Protocol(portSpec.Protocol),
			})
		}

		podSpec.Spec.Containers = append(podSpec.Spec.Containers, container)
	}

	pod, err := s.clientset.CoreV1().Pods(req.Namespace).Create(ctx, podSpec, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create pod: %v", err)
		return nil, err
	}

	return &proto.PodResponse{Pod: s.convertPodToProto(pod)}, nil
}

// UpdatePod updates an existing pod
func (s *Server) UpdatePod(ctx context.Context, req *proto.UpdatePodRequest) (*proto.PodResponse, error) {
	// Get existing pod
	existingPod, err := s.clientset.CoreV1().Pods(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Update labels
	if req.Spec.Labels != nil {
		existingPod.Labels = req.Spec.Labels
	}

	// Update containers
	if len(req.Spec.Containers) > 0 {
		var containers []v1.Container
		for _, containerSpec := range req.Spec.Containers {
			container := v1.Container{
				Name:  containerSpec.Name,
				Image: containerSpec.Image,
			}

			for _, portSpec := range containerSpec.Ports {
				container.Ports = append(container.Ports, v1.ContainerPort{
					ContainerPort: int32(portSpec.ContainerPort),
					Protocol:      v1.Protocol(portSpec.Protocol),
				})
			}

			containers = append(containers, container)
		}
		existingPod.Spec.Containers = containers
	}

	pod, err := s.clientset.CoreV1().Pods(req.Namespace).Update(ctx, existingPod, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update pod: %v", err)
		return nil, err
	}

	return &proto.PodResponse{Pod: s.convertPodToProto(pod)}, nil
}

// DeletePod deletes a pod
func (s *Server) DeletePod(ctx context.Context, req *proto.DeleteRequest) (*emptypb.Empty, error) {
	err := s.clientset.CoreV1().Pods(req.Namespace).Delete(ctx, req.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete pod: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateDeployment creates a new deployment
func (s *Server) CreateDeployment(ctx context.Context, req *proto.CreateDeploymentRequest) (*proto.DeploymentResponse, error) {
	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Spec.Name,
			Namespace: req.Namespace,
			Labels:    req.Spec.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &req.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: req.Spec.Template.Labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: req.Spec.Template.Labels,
				},
				Spec: v1.PodSpec{},
			},
		},
	}

	// Add containers to pod template
	for _, containerSpec := range req.Spec.Template.Containers {
		container := v1.Container{
			Name:  containerSpec.Name,
			Image: containerSpec.Image,
		}

		for _, portSpec := range containerSpec.Ports {
			container.Ports = append(container.Ports, v1.ContainerPort{
				ContainerPort: int32(portSpec.ContainerPort),
				Protocol:      v1.Protocol(portSpec.Protocol),
			})
		}

		deploymentSpec.Spec.Template.Spec.Containers = append(deploymentSpec.Spec.Template.Spec.Containers, container)
	}

	deployment, err := s.clientset.AppsV1().Deployments(req.Namespace).Create(ctx, deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create deployment: %v", err)
		return nil, err
	}

	return &proto.DeploymentResponse{Deployment: s.convertDeploymentToProto(deployment)}, nil
}

// UpdateDeployment updates an existing deployment
func (s *Server) UpdateDeployment(ctx context.Context, req *proto.UpdateDeploymentRequest) (*proto.DeploymentResponse, error) {
	// Get existing deployment
	existingDep, err := s.clientset.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Update spec
	if req.Spec.Replicas != 0 {
		existingDep.Spec.Replicas = &req.Spec.Replicas
	}

	if req.Spec.Labels != nil {
		existingDep.Labels = req.Spec.Labels
	}

	if req.Spec.Template != nil && req.Spec.Template.Labels != nil {
		existingDep.Spec.Template.Labels = req.Spec.Template.Labels
		existingDep.Spec.Selector.MatchLabels = req.Spec.Template.Labels
	}

	deployment, err := s.clientset.AppsV1().Deployments(req.Namespace).Update(ctx, existingDep, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update deployment: %v", err)
		return nil, err
	}

	return &proto.DeploymentResponse{Deployment: s.convertDeploymentToProto(deployment)}, nil
}

// DeleteDeployment deletes a deployment
func (s *Server) DeleteDeployment(ctx context.Context, req *proto.DeleteRequest) (*emptypb.Empty, error) {
	err := s.clientset.AppsV1().Deployments(req.Namespace).Delete(ctx, req.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete deployment: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateService creates a new service
func (s *Server) CreateService(ctx context.Context, req *proto.CreateServiceRequest) (*proto.ServiceResponse, error) {
	serviceSpec := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Spec.Name,
			Namespace: req.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceType(req.Spec.Type),
			Selector: req.Spec.Selector,
		},
	}

	// Add ports
	for _, portSpec := range req.Spec.Ports {
		servicePort := v1.ServicePort{
			Port:       int32(portSpec.ContainerPort),
			TargetPort: intOrString(portSpec.ContainerPort),
			Protocol:   v1.Protocol(portSpec.Protocol),
		}
		serviceSpec.Spec.Ports = append(serviceSpec.Spec.Ports, servicePort)
	}

	service, err := s.clientset.CoreV1().Services(req.Namespace).Create(ctx, serviceSpec, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create service: %v", err)
		return nil, err
	}

	return &proto.ServiceResponse{Service: s.convertServiceToProto(service)}, nil
}

// UpdateService updates an existing service
func (s *Server) UpdateService(ctx context.Context, req *proto.UpdateServiceRequest) (*proto.ServiceResponse, error) {
	// Get existing service
	existingSvc, err := s.clientset.CoreV1().Services(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Update spec
	if req.Spec.Type != "" {
		existingSvc.Spec.Type = v1.ServiceType(req.Spec.Type)
	}

	if req.Spec.Selector != nil {
		existingSvc.Spec.Selector = req.Spec.Selector
	}

	service, err := s.clientset.CoreV1().Services(req.Namespace).Update(ctx, existingSvc, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update service: %v", err)
		return nil, err
	}

	return &proto.ServiceResponse{Service: s.convertServiceToProto(service)}, nil
}

// DeleteService deletes a service
func (s *Server) DeleteService(ctx context.Context, req *proto.DeleteRequest) (*emptypb.Empty, error) {
	err := s.clientset.CoreV1().Services(req.Namespace).Delete(ctx, req.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete service: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateConfigMap creates a new configmap
func (s *Server) CreateConfigMap(ctx context.Context, req *proto.CreateConfigMapRequest) (*proto.ConfigMapResponse, error) {
	configMapSpec := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Spec.Name,
			Namespace: req.Namespace,
			Labels:    req.Spec.Labels,
		},
		Data: req.Spec.Data,
	}

	configMap, err := s.clientset.CoreV1().ConfigMaps(req.Namespace).Create(ctx, configMapSpec, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create configmap: %v", err)
		return nil, err
	}

	return &proto.ConfigMapResponse{Configmap: s.convertConfigMapToProto(configMap)}, nil
}

// UpdateConfigMap updates an existing configmap
func (s *Server) UpdateConfigMap(ctx context.Context, req *proto.UpdateConfigMapRequest) (*proto.ConfigMapResponse, error) {
	// Get existing configmap
	existingCm, err := s.clientset.CoreV1().ConfigMaps(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Update data and labels
	if req.Spec.Data != nil {
		existingCm.Data = req.Spec.Data
	}

	if req.Spec.Labels != nil {
		existingCm.Labels = req.Spec.Labels
	}

	configMap, err := s.clientset.CoreV1().ConfigMaps(req.Namespace).Update(ctx, existingCm, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update configmap: %v", err)
		return nil, err
	}

	return &proto.ConfigMapResponse{Configmap: s.convertConfigMapToProto(configMap)}, nil
}

// DeleteConfigMap deletes a configmap
func (s *Server) DeleteConfigMap(ctx context.Context, req *proto.DeleteRequest) (*emptypb.Empty, error) {
	err := s.clientset.CoreV1().ConfigMaps(req.Namespace).Delete(ctx, req.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete configmap: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetPodLogs retrieves logs from a pod
func (s *Server) GetPodLogs(ctx context.Context, req *proto.PodLogsRequest) (*proto.LogsResponse, error) {
	logOptions := &v1.PodLogOptions{
		Container: req.ContainerName,
		Follow:    req.Follow,
		TailLines: func() *int64 {
			if req.TailLines != 0 {
				v := int64(req.TailLines)
				return &v
			} else {
				return nil
			}
		}(),
	}

	reqLog := s.clientset.CoreV1().Pods(req.Namespace).GetLogs(req.PodName, logOptions)
	logs, err := reqLog.Stream(ctx)
	if err != nil {
		klog.Errorf("Failed to get pod logs: %v", err)
		return nil, err
	}
	defer logs.Close()

	var logData strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			logData.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return &proto.LogsResponse{Logs: logData.String()}, nil
}

// ExecPod executes a command in a pod (streaming)
func (s *Server) ExecPod(req *proto.ExecRequest, stream proto.K8SService_ExecPodServer) error {
	// This is a simplified implementation - in a real scenario you'd use the Kubernetes exec API
	// For now, just return a message
	response := &proto.ExecResponse{
		Output:  fmt.Sprintf("Executing command '%s' in pod %s/%s", req.Command, req.Namespace, req.PodName),
		IsError: false,
	}

	if err := stream.Send(response); err != nil {
		return err
	}

	return nil
}

// Helper functions for converting Kubernetes objects to protobuf

func (s *Server) convertPodToProto(pod *v1.Pod) *proto.Pod {
	protoPod := &proto.Pod{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Node:      pod.Spec.NodeName,
		Age:       calculateAge(pod.CreationTimestamp),
		Labels:    pod.Labels,
	}

	// Convert containers
	for _, container := range pod.Spec.Containers {
		protoContainer := &proto.Container{
			Name:  container.Name,
			Image: container.Image,
			Status: func() string {
				for _, status := range pod.Status.ContainerStatuses {
					if status.Name == container.Name {
						if status.Ready {
							return "Running"
						}
						return "Waiting"
					}
				}
				return "Unknown"
			}(),
		}

		// Convert ports
		for _, port := range container.Ports {
			protoPort := &proto.Port{
				Protocol:      string(port.Protocol),
				ContainerPort: int32(port.ContainerPort),
			}
			protoContainer.Ports = append(protoContainer.Ports, protoPort)
		}

		protoPod.Containers = append(protoPod.Containers, protoContainer)
	}

	return protoPod
}

func (s *Server) convertDeploymentToProto(dep *appsv1.Deployment) *proto.Deployment {
	return &proto.Deployment{
		Name:              dep.Name,
		Namespace:         dep.Namespace,
		Replicas:          *dep.Spec.Replicas,
		ReadyReplicas:     dep.Status.ReadyReplicas,
		AvailableReplicas: dep.Status.AvailableReplicas,
		Age:               calculateAge(dep.CreationTimestamp),
		Labels:            dep.Labels,
	}
}

func (s *Server) convertServiceToProto(svc *v1.Service) *proto.Service {
	protoSvc := &proto.Service{
		Name:       svc.Name,
		Namespace:  svc.Namespace,
		Type:       string(svc.Spec.Type),
		ClusterIp:  svc.Spec.ClusterIP,
		ExternalIp: getExternalIP(svc),
		Age:        calculateAge(svc.CreationTimestamp),
		Labels:     svc.Labels,
	}

	// Convert ports
	for _, port := range svc.Spec.Ports {
		portStr := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		protoSvc.Ports = append(protoSvc.Ports, portStr)
	}

	return protoSvc
}

func (s *Server) convertConfigMapToProto(cm *v1.ConfigMap) *proto.ConfigMap {
	return &proto.ConfigMap{
		Name:      cm.Name,
		Namespace: cm.Namespace,
		Data:      cm.Data,
		Age:       calculateAge(cm.CreationTimestamp),
		Labels:    cm.Labels,
	}
}

// Helper functions

func getExternalIP(svc *v1.Service) string {
	if svc.Spec.Type == v1.ServiceTypeLoadBalancer {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				return ingress.IP
			}
			if ingress.Hostname != "" {
				return ingress.Hostname
			}
		}
	}
	return "<none>"
}

func intOrString(i int32) intstr.IntOrString {
	return intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: i,
	}
}
