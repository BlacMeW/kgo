# KGO Developer Guide

## Overview

KGO is a comprehensive Kubernetes dashboard built in Go that provides multiple interfaces (REST API and Terminal UI) for managing Kubernetes resources. This guide provides detailed information for developers who want to contribute to, extend, or maintain the project.

## Architecture

### Core Components

```
k8s-dashboard/
├── cmd/server/           # Application entry points
├── pkg/
│   ├── api/             # REST API handlers and routing
│   ├── k8s/             # Kubernetes client operations
│   ├── tui/             # Terminal User Interface
│   ├── grpc/            # gRPC client/server (optional)
│   ├── config/          # Configuration management
│   ├── metrics/         # Metrics collection and reporting
│   └── utils/           # Shared utilities
├── proto/               # Protocol buffer definitions
└── test scripts         # Integration and testing scripts
```

### Design Principles

- **Modular Architecture**: Clean separation of concerns with independent packages
- **Asynchronous Operations**: Non-blocking UI with concurrent data loading
- **Extensible Design**: Easy to add new resource types and interfaces
- **Multiple Interfaces**: REST API and TUI for different use cases
- **Production Ready**: Proper error handling, logging, and monitoring

## Development Setup

### Prerequisites

- Go 1.24+
- Kubernetes cluster (local or remote)
- kubectl configured for cluster access
- Protocol buffer compiler (for gRPC development)

### Getting Started

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd k8s-dashboard
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Build the application**
   ```bash
   go build -o kgo ./cmd/server
   ```

4. **Run in development mode**
   ```bash
   # Terminal UI mode
   ./kgo -tui -kubeconfig=/path/to/kubeconfig
   ```

### Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and test**
   ```bash
   go test ./...
   go build ./cmd/server
   ```

3. **Run integration tests**
   ```bash
   ./test_integration.sh
   ```

4. **Submit a pull request**

## Package Structure

### pkg/api/

Contains REST API handlers and routing logic.

**Key Files:**
- `handler.go` - Main API handler with pod operations
- `resource_handler.go` - Generic resource operations (deployments, services, configmaps)
- `websocket.go` - WebSocket streaming for real-time updates

**Adding New Endpoints:**
```go
// Add to main.go routing
v1.GET("/newresource", newHandler.ListNewResources)
v1.POST("/newresource/:namespace", newHandler.CreateNewResource)

// Implement in resource_handler.go
func (h *ResourceHandler) ListNewResources(c *gin.Context) {
    // Implementation
}
```

### pkg/k8s/

Kubernetes client operations and resource management.

**Key Files:**
- `client.go` - Kubernetes client initialization and resource operations
- `types.go` - Custom types and conversions

**Adding New Resource Types:**
```go
// Add to client.go
func (c *Client) ListNewResources(namespace string) ([]NewResource, error) {
    // Implementation using c.clientset
}

// Add conversion functions if using gRPC
func convertProtoToNewResource(proto *proto.NewResource) *NewResource {
    // Implementation
}
```

### pkg/tui/

Advanced Terminal User Interface implementation.

**Key Features:**
- Asynchronous data loading
- Multiple view modes (List, Details, YAML, Logs, Relationships)
- Theming support (9 color themes)
- Split-pane layouts
- Advanced filtering and search

**Key Files:**
- `tui.go` - Main TUI implementation with event handling and drawing
- `themes.go` - Color theme definitions

**Adding New View Modes:**
```go
const (
    ViewModeNew ViewMode = iota
    // ... existing modes
)

// Add to draw() method
case ViewModeNew:
    t.drawNewView(width, height)

// Implement drawNewView method
func (t *TUI) drawNewView(width, height int) {
    // Implementation
}
```

### pkg/config/

Configuration management with support for multiple sources.

**Key Files:**
- `config.go` - Configuration loading and validation

**Configuration Structure:**
```go
type Config struct {
    Server     ServerConfig     `yaml:"server"`
    Kubernetes KubernetesConfig `yaml:"kubernetes"`
    Logging    LoggingConfig    `yaml:"logging"`
}
```

### pkg/metrics/

Metrics collection and monitoring.

**Key Files:**
- `handler.go` - Metrics API endpoints
- `collector.go` - Metrics collection logic

### pkg/grpc/

Optional gRPC implementation for distributed architectures.

**Key Files:**
- `client.go` - gRPC client implementation
- `server.go` - gRPC server implementation

## Adding New Resource Types

### 1. Update Kubernetes Client (pkg/k8s/client.go)

```go
// Add resource type constant
const (
    ResourceNewResource ResourceType = iota
    // ... existing types
)

// Add CRUD operations
func (c *Client) ListNewResources(namespace string) ([]NewResource, error) {
    // Implementation
}

func (c *Client) CreateNewResource(namespace string, resource *NewResource) error {
    // Implementation
}

func (c *Client) UpdateNewResource(namespace, name string, resource *NewResource) error {
    // Implementation
}

func (c *Client) DeleteNewResource(namespace, name string) error {
    // Implementation
}
```

### 2. Update API Handlers (pkg/api/)

```go
// Add to resource_handler.go
func (h *ResourceHandler) ListNewResources(c *gin.Context) {
    namespace := c.Query("namespace")
    if namespace == "" {
        namespace = "default"
    }

    resources, err := h.clientset.ListNewResources(namespace)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"resources": resources})
}

// Add other CRUD handlers...
```

### 3. Update Routes (cmd/server/main.go)

```go
// Add to v1 group
v1.GET("/newresources", resourceHandler.ListNewResources)
v1.POST("/newresources/:namespace", resourceHandler.CreateNewResource)
v1.PUT("/newresources/:namespace/:name", resourceHandler.UpdateNewResource)
v1.DELETE("/newresources/:namespace/:name", resourceHandler.DeleteNewResource)
```

### 4. Update TUI (pkg/tui/tui.go)

```go
// Add to ResourceType enum
const (
    ResourceNewResource ResourceType = iota
    // ... existing types
)

// Add to DisplayName method
func (rt ResourceType) DisplayName() string {
    switch rt {
    case ResourceNewResource:
        return "NewResources"
    // ... existing cases
    }
}

// Update getTableHeaders and related methods
func (t *TUI) getTableHeaders() []string {
    switch t.currentView {
    case ResourceNewResource:
        return []string{"Name", "Status", "Age"}
    // ... existing cases
    }
}

// Add to keyboard shortcuts (numbers 1-5 are taken, use letters)
case 'x':
    t.currentView = ResourceNewResource
    t.selected = 0
```

### 5. Update Protocol Buffers (proto/k8s.proto)

```protobuf
message NewResource {
    string name = 1;
    string namespace = 2;
    // ... other fields
}

service KubernetesService {
    // ... existing rpcs
    rpc ListNewResources(ListRequest) returns (NewResourceList);
    rpc CreateNewResource(NewResource) returns (NewResource);
    rpc UpdateNewResource(NewResource) returns (NewResource);
    rpc DeleteNewResource(DeleteRequest) returns (Empty);
}
```

## Testing

### Unit Tests

Run unit tests for all packages:
```bash
go test ./... -v
```

Run tests for specific package:
```bash
go test ./pkg/k8s -v
```

### Integration Tests

Use the provided integration test script:
```bash
chmod +x test_integration.sh
./test_integration.sh
```

### Manual Testing

Test API endpoints:
```bash
# List resources
curl http://localhost:8080/api/v1/pods?namespace=default

# Create resource
curl -X POST -H "Content-Type: application/json" \
  -d '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod"},"spec":{"containers":[{"name":"nginx","image":"nginx"}]}}' \
  http://localhost:8080/api/v1/pods/default
```

Test WebSocket connections:
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/pods/watch?namespace=default');
ws.onmessage = (event) => console.log('Change:', JSON.parse(event.data));
```

## Asynchronous Architecture

### Data Loading Pattern

The TUI uses goroutines and channels for non-blocking operations:

```go
// Start async loading
go t.loadPodsAsync()
go t.loadDeploymentsAsync()
// ... other resources

// Handle updates in main goroutine
go t.handleDataUpdates()

func (t *TUI) handleDataUpdates() {
    for update := range t.dataChan {
        // Update UI with new data
        t.updateUI(update)
    }
}
```

### Benefits

- **Responsive UI**: Never blocks on network operations
- **Concurrent Loading**: Multiple resources load simultaneously
- **Error Isolation**: One resource failure doesn't affect others
- **Real-time Updates**: Background refresh without UI freezing

## Theming System

### Adding New Themes

```go
func NewAwesomeTheme() Theme {
    return Theme{
        background: tcell.ColorBlack,
        foreground: tcell.ColorWhite,
        header:     tcell.ColorBlue,
        accent:     tcell.ColorGreen,
        selected:   tcell.ColorYellow,
    }
}

// Add to nextTheme() method
themes := []Theme{
    DefaultTheme(),
    DarkTheme(),
    LightTheme(),
    SolarizedTheme(),
    DraculaTheme(),
    NordTheme(),
    GruvboxTheme(),
    MonokaiTheme(),
    CyberpunkTheme(),
    NewAwesomeTheme(), // Add new theme
}
```

### Theme Structure

```go
type Theme struct {
    background tcell.Color  // Main background color
    foreground tcell.Color  // Main text color
    header     tcell.Color  // Header/title color
    accent     tcell.Color  // Accent/border color
    selected   tcell.Color  // Selected item color
}
```

### Kubernetes Deployment
}
```
```

## Deployment

### Docker Build

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o kgo ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/kgo .
CMD ["./kgo"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kgo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kgo
  template:
    metadata:
      labels:
        app: kgo
    spec:
      serviceAccountName: kgo-service-account
      containers:
      - name: kgo
        image: your-registry/kgo:latest
        ports:
        - containerPort: 8080
        env:
        - name: KUBECONFIG
          value: "/etc/kubernetes/kubeconfig"
        volumeMounts:
        - name: kubeconfig
          mountPath: /etc/kubernetes/
      volumes:
      - name: kubeconfig
        secret:
          secretName: kubeconfig-secret
```

## Contributing Guidelines
```

### Code Style

- Follow Go conventions and formatting
- Use `gofmt` and `goimports`
- Add comments for exported functions
- Use meaningful variable names

### Commit Messages

```
feat: add new resource type support
fix: resolve theme switching bug
docs: update developer guide
refactor: improve async loading architecture
```

### Pull Request Process

1. Create a feature branch
2. Implement changes with tests
3. Update documentation
4. Ensure all tests pass
5. Submit PR with detailed description

### Code Review Checklist

- [ ] Code compiles without errors
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Documentation updated
- [ ] No breaking changes
- [ ] Follows existing patterns
- [ ] Proper error handling

## Troubleshooting

### Common Issues

**TUI not starting:**
- Check kubeconfig path
- Verify cluster connectivity with `kubectl get pods`
- Check Go version (1.24+ required)

**API errors:**
- Check namespace permissions
- Verify resource exists
- Check request format

**Theme switching not working:**
- Ensure TUI is in focus
- Check keyboard layout
- Verify 't' key binding

### Debug Mode

Enable verbose logging:
```bash
export GLOG_v=3
./kgo -tui
```

### Performance Profiling

```bash
go build -o kgo ./cmd/server
./kgo -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Future Enhancements

### Planned Features

- **Additional Resource Types**: StatefulSets, Jobs, CronJobs, Ingress
- **Advanced Filtering**: Multi-column sorting, custom filters
- **Metrics Dashboard**: Resource usage graphs and alerts
- **RBAC Integration**: Role-based access control
- **Plugin System**: Extensible architecture for custom resources
- **Multi-cluster Support**: Manage multiple clusters simultaneously

### Architecture Improvements

- **Caching Layer**: Redis integration for performance
- **Database Backend**: Persistent storage for configurations
- **Event Streaming**: Kafka integration for cluster events
- **Service Mesh**: Istio integration for advanced networking
- **CI/CD Pipeline**: Automated testing and deployment

## Support

For questions, issues, or contributions:

- **GitHub Issues**: Bug reports and feature requests
- **Pull Requests**: Code contributions
- **Discussions**: General questions and architecture discussions

## License

This project is licensed under the MIT License - see the LICENSE file for details.</content>
<parameter name="filePath">/DATA/LLM_Projs/Network/kgo/DEVELOPER_GUIDE.md