# KGO - Advanced Kubernetes Dashboard

**KGO** is a comprehensive Go-based Kubernetes dashboard that provides **REST API**, **Advanced Terminal User Interface (TUI)**, and **Modern Web UI** for managing Kubernetes resources with enterprise-grade features.

## Key Highlights

- 🚀 **Asynchronous Data Loading**: Non-blocking UI with concurrent resource fetching
- 🎨 **Advanced TUI**: Feature-rich terminal interface with theming, filtering, and multi-view support
- 📊 **Full Resource Support**: Pods, Deployments, Services, ConfigMaps, and extensible architecture
- 🔄 **Real-time Updates**: WebSocket streaming and background data refresh
- 🏗️ **Modular Design**: Clean architecture with separate concerns for API, TUI, and web interfaces
- ⚡ **Performance Optimized**: Concurrent operations and efficient Kubernetes API usage

## Features

- Connect to any Kubernetes cluster using kubeconfig
- **Triple Interface**: REST API, Terminal UI, and Web UI
- **Full Resource Support**: Pods, Deployments, Services, and ConfigMaps
- **Asynchronous Data Loading**: Non-blocking UI with concurrent resource fetching
- CRUD operations on all supported resources
- Real-time event streaming via WebSocket
- REST API with Gin
- **Advanced TUI**: Interactive terminal interface with filtering, theming, and multi-view support
- Modern Web UI with HTML/CSS/JavaScript
- Proper error handling and logging with klog
- Modular structure

## Project Structure

```
```
k8s-dashboard/
├── cmd/server/main.go       # Main application with web/TUI modes
├── pkg/
│   ├── api/                 # REST API handlers for all resources
│   ├── k8s/client.go        # Kubernetes client operations
│   ├── tui/tui.go          # Advanced Terminal User Interface
│   ├── config/              # Configuration management
│   ├── metrics/             # Metrics and monitoring
│   └── grpc/                # gRPC support (optional)
├── web/
│   ├── index.html          # Web UI HTML
│   ├── style.css           # Web UI styles
│   └── app.js              # Web UI JavaScript
├── proto/                   # Protocol buffer definitions
├── go.mod                   # Dependencies
├── test_integration.sh     # Integration tests
└── README.md               # This file
```
```

## Running the Application

1. Ensure you have Go installed and a Kubernetes cluster accessible.
2. Clone or navigate to the project directory.
3. Run `go mod tidy` to install dependencies.
4. Run the server:

   ```bash
   go run cmd/server/main.go -kubeconfig=/path/to/kubeconfig
   ```

   If running in-cluster, omit the `-kubeconfig` flag.

5. The server will start on port 8080.

## Usage

### Web UI Mode (Default)

```bash
./bin/server -kubeconfig=/path/to/kubeconfig
```

Access the web interface at: http://localhost:8080

#### Web UI Features

- **Modern Interface**: Clean, responsive design with dark theme
- **Real-time Updates**: WebSocket connection for live pod changes
- **CRUD Operations**: Create, view, edit, and delete pods
- **Namespace Management**: Switch between namespaces
- **Pod Details**: View complete pod specifications in JSON format
- **Status Indicators**: Color-coded pod status (Running, Pending, Failed)
- **Responsive Design**: Works on desktop and mobile devices

#### Web UI Screenshots

The web UI includes:
- Header with namespace selector and action buttons
- Pod table with status, readiness, age, and node information
- Modal dialogs for pod creation/editing and details viewing
- Real-time status updates and connection indicators

### Terminal UI Mode

```bash
./bin/server -tui -kubeconfig=/path/to/kubeconfig
```

#### Advanced TUI Features

- **Asynchronous Data Loading**: Non-blocking UI with concurrent resource fetching
- **Multi-Resource Support**: Pods, Deployments, Services, and ConfigMaps
- **Advanced Filtering**: Regex support, case-sensitive/insensitive, inverse filtering
- **Multiple View Modes**: List, Details, YAML, Logs, and Relationships views
- **Theming**: Multiple color themes with customizable appearance
- **Split-Pane Layout**: Horizontal/vertical split views for detailed inspection
- **Real-time Updates**: Background data refresh without UI freezing
- **Interactive Navigation**: Tab-based resource switching, keyboard shortcuts
- **Resource Relationships**: Visual representation of resource connections

#### TUI Controls

- **↑↓/←→** Navigate through resources
- **Enter** Show resource details
- **Tab** Switch between resource types (Pods/Deployments/Services/ConfigMaps)
- **r/F5** Refresh data asynchronously
- **d** Delete resource (with confirmation)
- **n** Change namespace
- **/** Advanced search/filtering
- **f** Clear filters
- **v** Cycle through view modes (List/Details/YAML/Logs/Relationships)
- **y** Toggle YAML view in details mode
- **j** Show logs for pods
- **s** Toggle split-pane view
- **S** Switch split layout (horizontal/vertical)
- **1-4** Quick switch to resource types
- **c** Create new pod (basic)
- **h/?** Show help
- **q** Quit

### API Mode (Programmatic Access)

Use the REST API directly for automation and integration:

```bash
# List pods
curl http://localhost:8080/api/v1/pods?namespace=default

# Create pod
curl -X POST -H "Content-Type: application/json" \
  -d '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod"},"spec":{"containers":[{"name":"nginx","image":"nginx"}]}}' \
  http://localhost:8080/api/v1/pods/default
```

#### TUI Features

- Real-time pod status display
- Interactive pod details view
- Namespace switching
- Pod deletion with confirmation
- Color-coded interface

## API Endpoints

### Pods
- `GET /api/v1/pods?namespace=default` - List pods in namespace
- `POST /api/v1/pods/:namespace` - Create a pod in namespace
- `PUT /api/v1/pods/:namespace/:name` - Update a pod
- `DELETE /api/v1/pods/:namespace/:name` - Delete a pod
- `GET /api/v1/pods/watch?namespace=default` - Watch pod changes (WebSocket)
- `GET /api/v1/pods/:namespace/:name/logs` - Get pod logs
- `GET /api/v1/pods/:namespace/:name/exec` - Execute commands in pod

### Deployments
- `GET /api/v1/deployments?namespace=default` - List deployments in namespace
- `POST /api/v1/deployments/:namespace` - Create a deployment in namespace
- `PUT /api/v1/deployments/:namespace/:name` - Update a deployment
- `DELETE /api/v1/deployments/:namespace/:name` - Delete a deployment

### Services
- `GET /api/v1/services?namespace=default` - List services in namespace
- `POST /api/v1/services/:namespace` - Create a service in namespace
- `PUT /api/v1/services/:namespace/:name` - Update a service
- `DELETE /api/v1/services/:namespace/:name` - Delete a service

### ConfigMaps
- `GET /api/v1/configmaps?namespace=default` - List configmaps in namespace
- `POST /api/v1/configmaps/:namespace` - Create a configmap in namespace
- `PUT /api/v1/configmaps/:namespace/:name` - Update a configmap
- `DELETE /api/v1/configmaps/:namespace/:name` - Delete a configmap

### Metrics
- `GET /api/v1/metrics/cluster` - Get cluster-wide metrics
- `GET /api/v1/metrics/namespace/:namespace` - Get namespace-specific metrics

## React Frontend Integration

### CRUD Operations

Use `fetch` for REST API calls:

```javascript
// List pods
fetch('/api/v1/pods?namespace=default')
  .then(res => res.json())
  .then(data => setPods(data.pods));

// Create pod
const podSpec = {
  apiVersion: 'v1',
  kind: 'Pod',
  metadata: { name: 'my-pod' },
  spec: {
    containers: [{
      name: 'my-container',
      image: 'nginx'
    }]
  }
};

fetch('/api/v1/pods/default', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(podSpec)
});

// Update pod
fetch('/api/v1/pods/default/my-pod', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(updatedPodSpec)
});

// Delete pod
fetch('/api/v1/pods/default/my-pod', {
  method: 'DELETE'
});
```

### Watching Changes

Use WebSocket for real-time updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/pods/watch?namespace=default');

ws.onopen = () => {
  console.log('Connected to pod watcher');
};

ws.onmessage = (event) => {
  const change = JSON.parse(event.data);
  console.log('Pod change:', change.type, change.object.metadata.name);
  // Update UI based on change.type (ADDED, MODIFIED, DELETED)
};

ws.onclose = () => {
  console.log('Disconnected from pod watcher');
};
```

## Asynchronous Data Loading Architecture

The TUI implements a sophisticated asynchronous data loading system to prevent UI freezing:

### Key Components

- **Concurrent Goroutines**: Each resource type loads in parallel goroutines
- **Channel Communication**: Thread-safe data updates via Go channels
- **Background Processing**: Dedicated goroutine handles data updates
- **Loading State Management**: Counter-based tracking of completion
- **Non-blocking UI**: User interactions remain responsive during data loading

### Benefits

- **Responsive Interface**: UI never freezes during data operations
- **Concurrent Loading**: All resources load simultaneously for faster overall loading
- **Error Resilience**: Individual failures don't block other resource loading
- **Real-time Updates**: Data refreshes in background without user intervention

## Extending for Other Resources

The application already supports Pods, Deployments, Services, and ConfigMaps. To add support for additional resources (e.g., StatefulSets, Jobs, etc.):

1. Add corresponding functions in `pkg/k8s/client.go` (e.g., `ListStatefulSets`, `CreateStatefulSet`)
2. Add handlers in `pkg/api/` directory
3. Add routes in `cmd/server/main.go`
4. Update TUI in `pkg/tui/tui.go` to handle the new resource type
5. Update protobuf definitions in `proto/k8s.proto` if using gRPC

## Authentication

Currently supports kubeconfig-based authentication. For production:

- Use Kubernetes service account tokens
- Implement JWT authentication
- Add middleware for authorization

## Testing

### Unit Tests
Run unit tests with mocked Kubernetes client:
```bash
go test ./... -v
```

### Integration Tests
For full integration testing with a real Kubernetes cluster:

1. Start the server:
   ```bash
   ./bin/server -kubeconfig=/path/to/kubeconfig
   ```

2. Run the integration test script:
   ```bash
   chmod +x test_integration.sh
   ./test_integration.sh
   ```

The integration script tests all CRUD operations and verifies responses.

### Manual Testing with curl

```bash
# List resources
curl http://localhost:8080/api/v1/pods?namespace=default
curl http://localhost:8080/api/v1/deployments?namespace=default
curl http://localhost:8080/api/v1/services?namespace=default
curl http://localhost:8080/api/v1/configmaps?namespace=default

# Create resources
curl -X POST -H "Content-Type: application/json" \
  -d '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod"},"spec":{"containers":[{"name":"nginx","image":"nginx"}]}}' \
  http://localhost:8080/api/v1/pods/default

curl -X POST -H "Content-Type: application/json" \
  -d '{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"test-config"},"data":{"key":"value"}}' \
  http://localhost:8080/api/v1/configmaps/default

# Update resources
curl -X PUT -H "Content-Type: application/json" \
  -d '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","labels":{"app":"test"}},"spec":{"containers":[{"name":"nginx","image":"nginx"}]}}' \
  http://localhost:8080/api/v1/pods/default/test-pod

# Delete resources
curl -X DELETE http://localhost:8080/api/v1/pods/default/test-pod
curl -X DELETE http://localhost:8080/api/v1/configmaps/default/test-config
```

### WebSocket Testing
Test real-time pod watching:
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/pods/watch?namespace=default');
ws.onmessage = (event) => console.log('Pod change:', JSON.parse(event.data));
```

## gRPC Implementation Status

**✅ FULLY FUNCTIONAL** - The gRPC implementation has been completed and tested.

### What Was Fixed:

- ✅ **Missing Conversion Functions**: Added `convertProtoToPod`, `convertProtoToDeployment`, `convertProtoToService`, `convertProtoToConfigMap`
- ✅ **Type Consistency**: All client methods now return Kubernetes types instead of protobuf types
- ✅ **Import Dependencies**: Added required imports (`appsv1`, `metav1`)
- ✅ **Protobuf Generation**: Regenerated fresh protobuf Go files with correct type definitions
- ✅ **Compilation**: All gRPC packages now compile successfully

### Architecture Options:

The project now supports two architectural approaches:

#### **Current: Direct Client (Default)**
- Uses goroutines + channels for async UI
- No network overhead
- Simpler deployment
- **Status**: ✅ Active

#### **Alternative: gRPC Client**
- Distributed architecture support
- Network-enabled operations
- Multi-server scalability
- **Status**: ✅ Ready for use

### Switching to gRPC Mode:

To use gRPC instead of direct client calls:

1. **Modify `cmd/server/main.go`**:
   ```go
   // Start gRPC server
   grpcServer := grpc.NewServer(clientset)
   go grpcServer.Start(":50051")
   
   // Use gRPC client in TUI
   grpcClient, _ := grpc.NewClient("localhost:50051")
   tui, _ := tui.NewTUI(grpcClient)
   ```

2. **Benefits of gRPC mode**:
   - Separate TUI and API server processes
   - Load balancing across multiple API servers
   - Network-based architecture
   - Better for microservices deployments

## Production Considerations

- **Asynchronous Architecture**: Non-blocking data loading prevents UI freezing under load
- **Resource Scaling**: Support for multiple resource types with efficient concurrent loading
- Add TLS/HTTPS for secure communication
- Implement rate limiting for API endpoints
- Add comprehensive error handling and logging with klog
- Use structured logging with appropriate log levels
- Add metrics and monitoring (already partially implemented in `pkg/metrics/`)
- Implement graceful shutdown handling
- Consider connection pooling for Kubernetes API calls
- Add caching layer for frequently accessed resources
- Implement authentication and authorization middleware