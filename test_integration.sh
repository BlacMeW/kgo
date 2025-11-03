#!/bin/bash

# Integration Test Script for Kubernetes Dashboard Backend
# This script assumes you have a running Kubernetes cluster and the server is started

SERVER_URL="http://localhost:8080"
NAMESPACE="default"

echo "Starting integration tests..."

# Test 1: List Pods
echo "Test 1: Listing pods"
response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/v1/pods?namespace=$NAMESPACE")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" -eq 200 ]; then
    echo "✅ List pods: SUCCESS"
    echo "Response: $body"
else
    echo "❌ List pods: FAILED (HTTP $http_code)"
    echo "Response: $body"
fi

# Test 2: Create a test pod
echo "Test 2: Creating a test pod"
pod_data='{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "test-integration-pod"
  },
  "spec": {
    "containers": [{
      "name": "test-container",
      "image": "nginx:alpine",
      "ports": [{"containerPort": 80}]
    }]
  }
}'

response=$(curl -s -w "\n%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d "$pod_data" \
  "$SERVER_URL/api/v1/pods/$NAMESPACE")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" -eq 201 ]; then
    echo "✅ Create pod: SUCCESS"
else
    echo "❌ Create pod: FAILED (HTTP $http_code)"
    echo "Response: $body"
fi

# Wait a moment for pod to be created
sleep 2

# Test 3: List pods again to verify creation
echo "Test 3: Verifying pod creation"
response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/v1/pods?namespace=$NAMESPACE")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" -eq 200 ] && echo "$body" | grep -q "test-integration-pod"; then
    echo "✅ Pod creation verified: SUCCESS"
else
    echo "❌ Pod creation verification: FAILED"
fi

# Test 4: Update the pod (add a label)
echo "Test 4: Updating the pod"
update_data='{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "test-integration-pod",
    "labels": {
      "app": "test"
    }
  },
  "spec": {
    "containers": [{
      "name": "test-container",
      "image": "nginx:alpine",
      "ports": [{"containerPort": 80}]
    }]
  }
}'

response=$(curl -s -w "\n%{http_code}" -X PUT \
  -H "Content-Type: application/json" \
  -d "$update_data" \
  "$SERVER_URL/api/v1/pods/$NAMESPACE/test-integration-pod")

http_code=$(echo "$response" | tail -n1)

if [ "$http_code" -eq 200 ]; then
    echo "✅ Update pod: SUCCESS"
else
    echo "❌ Update pod: FAILED (HTTP $http_code)"
fi

# Test 5: Delete the pod
echo "Test 5: Deleting the pod"
response=$(curl -s -w "\n%{http_code}" -X DELETE \
  "$SERVER_URL/api/v1/pods/$NAMESPACE/test-integration-pod")

http_code=$(echo "$response" | tail -n1)

if [ "$http_code" -eq 200 ]; then
    echo "✅ Delete pod: SUCCESS"
else
    echo "❌ Delete pod: FAILED (HTTP $http_code)"
fi

echo "Integration tests completed!"