#!/bin/bash

echo "=== Testing Kubernetes Dashboard with Minikube ==="
echo ""

# Check if minikube is installed
if ! command -v minikube &> /dev/null; then
    echo "❌ Minikube not found. Install it first:"
    echo "curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64"
    echo "sudo install minikube-linux-amd64 /usr/local/bin/minikube"
    exit 1
fi

# Check if minikube is running
if ! minikube status | grep -q "Running"; then
    echo "Starting minikube cluster..."
    minikube start
    if [ $? -ne 0 ]; then
        echo "❌ Failed to start minikube"
        exit 1
    fi
fi

echo "✅ Minikube is running"

# Start the dashboard server
echo "Starting dashboard server..."
./bin/server -kubeconfig=$HOME/.kube/config &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test basic connectivity
echo "Testing API connectivity..."
if curl -s http://localhost:8080/api/v1/pods?namespace=default > /dev/null; then
    echo "✅ Server is responding"
else
    echo "❌ Server not responding"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Run integration tests
echo "Running integration tests..."
./test_integration.sh

# Cleanup
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null

echo "✅ Testing completed!"