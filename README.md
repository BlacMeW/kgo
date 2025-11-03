# Kubernetes Dashboard Backend

This is a Go backend for a Kubernetes dashboard that provides both **REST API** and **Terminal User Interface (TUI)** for managing Kubernetes resources.

## Features

- Connect to any Kubernetes cluster using kubeconfig
- **Triple Interface**: REST API, Terminal UI, and Web UI
- CRUD operations on Pods
- Real-time event streaming via WebSocket
- REST API with Gin
- Interactive TUI with tcell/termbox-go
- Modern Web UI with HTML/CSS/JavaScript
- Proper error handling and logging with klog
- Modular structure

## Project Structure

```
```
k8s-dashboard/
├── cmd/server/main.go       # Main application with web/TUI modes
├── pkg/
│   ├── api/handlers.go      # REST API handlers
│   ├── k8s/client.go        # Kubernetes client operations
│   ├── tui/tui.go          # Terminal User Interface
│   └── utils/               # Utility functions
├── web/
│   ├── index.html          # Web UI HTML
│   ├── style.css           # Web UI styles
│   └── app.js              # Web UI JavaScript
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

#### TUI Controls

- **↑↓** Navigate through pods
- **Enter** Show pod details
- **r** Refresh pod list
- **d** Delete pod (with confirmation)
- **n** Change namespace
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

- `GET /api/v1/pods?namespace=default` - List pods in namespace
- `POST /api/v1/pods/:namespace` - Create a pod in namespace
- `PUT /api/v1/pods/:namespace/:name` - Update a pod
- `DELETE /api/v1/pods/:namespace/:name` - Delete a pod
- `GET /api/v1/pods/watch?namespace=default` - Watch pod changes (WebSocket)

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

## Extending for Other Resources

To add support for Deployments, Services, ConfigMaps, etc.:

1. Add corresponding functions in `pkg/k8s/client.go` (e.g., `ListDeployments`, `CreateDeployment`)
2. Add handlers in `pkg/api/handlers.go`
3. Add routes in `cmd/server/main.go`

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
# List pods
curl http://localhost:8080/api/v1/pods?namespace=default

# Create pod
curl -X POST -H "Content-Type: application/json" \
  -d '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod"},"spec":{"containers":[{"name":"nginx","image":"nginx"}]}}' \
  http://localhost:8080/api/v1/pods/default

# Update pod
curl -X PUT -H "Content-Type: application/json" \
  -d '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","labels":{"app":"test"}},"spec":{"containers":[{"name":"nginx","image":"nginx"}]}}' \
  http://localhost:8080/api/v1/pods/default/test-pod

# Delete pod
curl -X DELETE http://localhost:8080/api/v1/pods/default/test-pod
```

### WebSocket Testing
Test real-time pod watching:
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/pods/watch?namespace=default');
ws.onmessage = (event) => console.log('Pod change:', JSON.parse(event.data));
```

## Production Considerations

- Add TLS/HTTPS
- Implement rate limiting
- Add comprehensive error handling
- Use structured logging
- Add metrics and monitoring
- Implement graceful shutdown