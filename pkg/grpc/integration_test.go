package grpc

import (
	"context"
	"testing"
	"time"

	"k8s-dashboard/proto"
)

// TestGRPCIntegration tests the full gRPC client-server communication
func TestGRPCIntegration(t *testing.T) {
	// This test requires a running gRPC server
	// In a real CI/CD environment, you'd start a test server

	client, err := NewClient("localhost:50051")
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to gRPC server: %v", err)
		return
	}
	defer client.Close()

	// Test ListPods
	pods, err := client.ListPods("default")
	if err != nil {
		t.Errorf("ListPods failed: %v", err)
	} else {
		t.Logf("Successfully listed %d pods", len(pods))
	}

	// Test ListDeployments
	deployments, err := client.ListDeployments("default")
	if err != nil {
		t.Errorf("ListDeployments failed: %v", err)
	} else {
		t.Logf("Successfully listed %d deployments", len(deployments))
	}

	// Test ListServices
	services, err := client.ListServices("default")
	if err != nil {
		t.Errorf("ListServices failed: %v", err)
	} else {
		t.Logf("Successfully listed %d services", len(services))
	}

	// Test ListConfigMaps
	configmaps, err := client.ListConfigMaps("default")
	if err != nil {
		t.Errorf("ListConfigMaps failed: %v", err)
	} else {
		t.Logf("Successfully listed %d configmaps", len(configmaps))
	}

	// Test new resources
	// Test ListStatefulSets
	statefulsets, err := client.ListStatefulSets("default")
	if err != nil {
		t.Errorf("ListStatefulSets failed: %v", err)
	} else {
		t.Logf("Successfully listed %d StatefulSets", len(statefulsets))
	}

	// Test ListDaemonSets
	daemonsets, err := client.ListDaemonSets("default")
	if err != nil {
		t.Errorf("ListDaemonSets failed: %v", err)
	} else {
		t.Logf("Successfully listed %d DaemonSets", len(daemonsets))
	}

	// Test ListJobs
	jobs, err := client.ListJobs("default")
	if err != nil {
		t.Errorf("ListJobs failed: %v", err)
	} else {
		t.Logf("Successfully listed %d Jobs", len(jobs))
	}

	// Test ListCronJobs
	cronjobs, err := client.ListCronJobs("default")
	if err != nil {
		t.Errorf("ListCronJobs failed: %v", err)
	} else {
		t.Logf("Successfully listed %d CronJobs", len(cronjobs))
	}

	// Test ListIngresses
	ingresses, err := client.ListIngresses("default")
	if err != nil {
		t.Errorf("ListIngresses failed: %v", err)
	} else {
		t.Logf("Successfully listed %d Ingresses", len(ingresses))
	}

	// Test ListPVCs
	pvcs, err := client.ListPVCs("default")
	if err != nil {
		t.Errorf("ListPVCs failed: %v", err)
	} else {
		t.Logf("Successfully listed %d PVCs", len(pvcs))
	}

	// Test ListSecrets
	secrets, err := client.ListSecrets("default")
	if err != nil {
		t.Errorf("ListSecrets failed: %v", err)
	} else {
		t.Logf("Successfully listed %d Secrets", len(secrets))
	}

	// Test ListServiceAccounts
	serviceaccounts, err := client.ListServiceAccounts("default")
	if err != nil {
		t.Errorf("ListServiceAccounts failed: %v", err)
	} else {
		t.Logf("Successfully listed %d ServiceAccounts", len(serviceaccounts))
	}
}

// TestGRPCClientTimeout tests client timeout behavior
func TestGRPCClientTimeout(t *testing.T) {
	client, err := NewClient("localhost:50051")
	if err != nil {
		t.Skipf("Skipping timeout test: cannot connect to gRPC server: %v", err)
		return
	}
	defer client.Close()

	// Test with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	_, err = client.client.ListPods(ctx, &proto.ListRequest{Namespace: "default"})
	if err == nil {
		t.Error("Expected timeout error, but got none")
	}
}

// TestGRPCConnectionFailure tests behavior when server is unreachable
func TestGRPCConnectionFailure(t *testing.T) {
	// Try to connect to a non-existent server
	client, err := NewClient("localhost:12345")
	if err != nil {
		// This is expected - connection should fail
		return
	}
	defer client.Close()

	// If we get here, try a call that should fail
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = client.client.ListPods(ctx, &proto.ListRequest{Namespace: "default"})
	if err == nil {
		t.Error("Expected connection error, but call succeeded")
	}
}

// TestGRPCClientReconnection tests client reconnection logic
func TestGRPCClientReconnection(t *testing.T) {
	client, err := NewClient("localhost:50051")
	if err != nil {
		t.Skipf("Skipping reconnection test: cannot connect to gRPC server: %v", err)
		return
	}

	// Close and try to reconnect
	client.Close()

	// Try to create a new connection
	newClient, err := NewClient("localhost:50051")
	if err != nil {
		t.Skipf("Skipping reconnection test: cannot reconnect to gRPC server: %v", err)
		return
	}
	defer newClient.Close()

	// Test that the new client works
	pods, err := newClient.ListPods("default")
	if err != nil {
		t.Errorf("Reconnected client failed: %v", err)
	} else {
		t.Logf("Reconnected client successfully listed %d pods", len(pods))
	}
}
