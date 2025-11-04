// Kubernetes Dashboard Web UI JavaScript

class KubernetesDashboard {
    constructor() {
        this.apiBase = window.location.origin;
        this.currentNamespace = 'default';
        this.currentResource = 'pods';
        this.websocket = null;
        this.selectedResource = null;
        this.allResources = [];
        this.filteredResources = [];
        this.searchTerm = '';
        this.autoRefreshInterval = null;
        this.logsFollowInterval = null;
        this.theme = localStorage.getItem('theme') || 'light';

        this.init();
    }

    init() {
        this.bindEvents();
        this.loadResources();
        this.connectWebSocket();
        this.setupKeyboardShortcuts();
        this.startAutoRefresh();
        this.applyTheme();
    }

    bindEvents() {
        // Resource tabs
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.switchResource(e.target.dataset.resource);
            });
        });

        // Namespace selector
        document.getElementById('namespace').addEventListener('change', (e) => {
            this.currentNamespace = e.target.value;
            this.loadResources();
            this.reconnectWebSocket();
        });

        // Search functionality
        const searchInput = document.getElementById('search-input');
        searchInput.addEventListener('input', (e) => {
            this.searchTerm = e.target.value.toLowerCase();
            this.applyFilters();
        });

        document.getElementById('clear-search').addEventListener('click', () => {
            this.clearSearch();
        });

        // Buttons
        document.getElementById('refresh-btn').addEventListener('click', () => this.loadResources());
        document.getElementById('create-btn').addEventListener('click', () => this.showCreateModal());
        document.getElementById('logs-btn').addEventListener('click', () => this.showLogsModal());
        document.getElementById('metrics-btn').addEventListener('click', () => this.showMetricsPanel());
        document.getElementById('theme-toggle').addEventListener('click', () => this.toggleTheme());

        // Metrics panel
        document.getElementById('close-metrics').addEventListener('click', () => this.hideMetricsPanel());

        // Modal events
        this.bindModalEvents();
    }

    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Don't trigger shortcuts when typing in inputs
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
                return;
            }

            switch (e.key) {
                case 'F5':
                    e.preventDefault();
                    this.loadResources();
                    break;
                case 'r':
                case 'R':
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        this.loadResources();
                    }
                    break;
                case 'n':
                case 'N':
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        this.showCreateModal();
                    }
                    break;
                case '/':
                    e.preventDefault();
                    document.getElementById('search-input').focus();
                    break;
                case 'Escape':
                    // Close any open modals
                    document.querySelectorAll('.modal.show').forEach(modal => {
                        this.hideModal(modal.id);
                    });
                    // Clear search if active
                    if (this.searchTerm) {
                        this.clearSearch();
                    }
                    break;
            }
        });
    }

    switchResource(resourceType) {
        this.currentResource = resourceType;

        // Update active tab
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.resource === resourceType);
        });

        // Update UI elements
        this.updateUIForResource();
        this.loadResources();
        this.reconnectWebSocket();

        // Update search placeholder
        const searchInput = document.getElementById('search-input');
        searchInput.placeholder = `🔍 Search ${resourceType}...`;

        // Show/hide logs button (only for pods)
        const logsBtn = document.getElementById('logs-btn');
        logsBtn.style.display = resourceType === 'pods' ? 'inline-block' : 'none';
    }

    updateUIForResource() {
        const headers = {
            pods: ['Name', 'Status', 'Ready', 'Restarts', 'Age', 'Node', 'Actions'],
            deployments: ['Name', 'Ready', 'Up-to-date', 'Available', 'Age', 'Actions'],
            services: ['Name', 'Type', 'Cluster-IP', 'External-IP', 'Ports', 'Age', 'Actions'],
            configmaps: ['Name', 'Data', 'Age', 'Actions']
        };

        const tableHeader = document.getElementById('table-header');
        const headerRow = headers[this.currentResource].map(header => `<th>${header}</th>`).join('');
        tableHeader.innerHTML = `<tr>${headerRow}</tr>`;

        // Update create button text
        const createBtn = document.getElementById('create-btn');
        createBtn.textContent = `➕ Create ${this.currentResource.charAt(0).toUpperCase() + this.currentResource.slice(0, -1)}`;
    }

    startAutoRefresh() {
        // Auto-refresh every 30 seconds
        this.autoRefreshInterval = setInterval(() => {
            if (!this.websocket || this.websocket.readyState !== WebSocket.OPEN) {
                this.loadResources();
            }
        }, 30000);
    }

    clearSearch() {
        this.searchTerm = '';
        document.getElementById('search-input').value = '';
        this.applyFilters();
        document.getElementById('clear-search').style.display = 'none';
    }

    applyFilters() {
        if (!this.searchTerm) {
            this.filteredResources = [...this.allResources];
            document.getElementById('clear-search').style.display = 'none';
            document.getElementById('filtered-count').style.display = 'none';
        } else {
            this.filteredResources = this.allResources.filter(resource =>
                resource.metadata.name.toLowerCase().includes(this.searchTerm) ||
                (resource.spec && resource.spec.nodeName && resource.spec.nodeName.toLowerCase().includes(this.searchTerm)) ||
                (resource.status && resource.status.phase && resource.status.phase.toLowerCase().includes(this.searchTerm)) ||
                (resource.spec && resource.spec.type && resource.spec.type.toLowerCase().includes(this.searchTerm))
            );
            document.getElementById('clear-search').style.display = 'inline-block';
            document.getElementById('filtered-count').textContent = `(${this.filteredResources.length} filtered)`;
            document.getElementById('filtered-count').style.display = 'inline';
        }

        this.renderResources(this.filteredResources);
    }

    bindModalEvents() {
        // Pod modal
        const podModal = document.getElementById('pod-modal');
        const podForm = document.getElementById('pod-form');

        document.getElementById('modal-cancel').addEventListener('click', () => this.hideModal('pod-modal'));
        document.getElementById('modal-save').addEventListener('click', () => this.savePod());

        podForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.savePod();
        });

        // Details modal
        document.getElementById('details-close').addEventListener('click', () => this.hideModal('details-modal'));

        // Delete modal
        document.getElementById('delete-cancel').addEventListener('click', () => this.hideModal('delete-modal'));
        document.getElementById('delete-confirm').addEventListener('click', () => this.confirmDelete());

        // Logs modal
        document.getElementById('logs-close').addEventListener('click', () => {
            this.stopLogsFollow();
            this.hideModal('logs-modal');
        });

        document.getElementById('logs-refresh').addEventListener('click', () => this.loadPodLogs());
        document.getElementById('logs-container').addEventListener('change', () => this.loadPodLogs());
        document.getElementById('logs-follow').addEventListener('change', (e) => {
            if (e.target.checked) {
                this.startLogsFollow(`${this.apiBase}/api/v1/pods/${this.currentNamespace}/${this.selectedPod}/logs`);
            } else {
                this.stopLogsFollow();
            }
        });

        // Close modals on outside click
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    if (modal.id === 'logs-modal') {
                        this.stopLogsFollow();
                    }
                    this.hideModal(modal.id);
                }
            });
        });

        // Close buttons
        document.querySelectorAll('.close').forEach(closeBtn => {
            closeBtn.addEventListener('click', (e) => {
                const modal = e.target.closest('.modal');
                if (modal.id === 'logs-modal') {
                    this.stopLogsFollow();
                }
                this.hideModal(modal.id);
            });
        });
    }

    async loadResources() {
        const refreshBtn = document.getElementById('refresh-btn');
        const spinner = refreshBtn.querySelector('.spinner');

        refreshBtn.disabled = true;
        spinner.style.display = 'inline-block';

        this.setStatus(`Loading ${this.currentResource}...`, 'loading');

        try {
            const response = await fetch(`${this.apiBase}/api/v1/${this.currentResource}?namespace=${this.currentNamespace}`);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            this.allResources = data[this.currentResource] || [];
            this.applyFilters();
            this.setStatus(`${this.currentResource} loaded successfully`, 'success');

        } catch (error) {
            console.error(`Failed to load ${this.currentResource}:`, error);
            this.setStatus(`Error loading ${this.currentResource}: ${error.message}`, 'error');
            this.allResources = [];
            this.applyFilters();
        } finally {
            refreshBtn.disabled = false;
            spinner.style.display = 'none';
        }
    }

    renderResources(resources) {
        const tbody = document.getElementById('resources-body');
        const resourceCount = document.getElementById('pod-count'); // Reusing pod-count element
        const logsBtn = document.getElementById('logs-btn');

        resourceCount.textContent = `${this.allResources.length} ${this.currentResource}`;

        // Show/hide logs button based on resource type and selection
        logsBtn.style.display = (this.currentResource === 'pods' && resources.length > 0) ? 'inline-block' : 'none';

        if (resources.length === 0) {
            const headers = {
                pods: 7, deployments: 6, services: 7, configmaps: 4
            };
            const colspan = headers[this.currentResource] || 5;
            const message = this.searchTerm ? `No ${this.currentResource} match your search` : `No ${this.currentResource} found`;
            tbody.innerHTML = `<tr><td colspan="${colspan}" class="loading">${message}</td></tr>`;
            return;
        }

        tbody.innerHTML = resources.map(resource => this.renderResourceRow(resource)).join('');
    }

    renderResourceRow(resource) {
        switch (this.currentResource) {
            case 'pods':
                return this.renderPodRow(resource);
            case 'deployments':
                return this.renderDeploymentRow(resource);
            case 'services':
                return this.renderServiceRow(resource);
            case 'configmaps':
                return this.renderConfigMapRow(resource);
            default:
                return `<tr><td colspan="5">Unknown resource type</td></tr>`;
        }
    }

    renderPodRow(pod) {
        return `
            <tr data-resource-name="${pod.metadata.name}">
                <td>
                    <div class="pod-name">${pod.metadata.name}</div>
                </td>
                <td><span class="status-${pod.status.phase.toLowerCase()}">${pod.status.phase}</span></td>
                <td>${this.getReadyCount(pod)}</td>
                <td>${this.getRestartCount(pod)}</td>
                <td>${this.formatAge(pod.metadata.creationTimestamp)}</td>
                <td>${pod.spec.nodeName || '<none>'}</td>
                <td>
                    <button class="btn action-btn view" onclick="dashboard.showResourceDetails('${pod.metadata.name}')" title="View details">👁️</button>
                    <button class="btn action-btn edit" onclick="dashboard.editResource('${pod.metadata.name}')" title="Edit pod">✏️</button>
                    <button class="btn action-btn logs" onclick="dashboard.showLogsModal('${pod.metadata.name}')" title="View logs">📋</button>
                    <button class="btn action-btn exec" onclick="dashboard.execPod('${pod.metadata.name}')" title="Execute command">💻</button>
                    <button class="btn action-btn delete" onclick="dashboard.deleteResource('${pod.metadata.name}')" title="Delete pod">🗑️</button>
                </td>
            </tr>
        `;
    }

    renderDeploymentRow(deployment) {
        const ready = deployment.status.readyReplicas || 0;
        const replicas = deployment.spec.replicas || 0;
        const updated = deployment.status.updatedReplicas || 0;
        const available = deployment.status.availableReplicas || 0;

        return `
            <tr data-resource-name="${deployment.metadata.name}">
                <td><div class="pod-name">${deployment.metadata.name}</div></td>
                <td>${ready}/${replicas}</td>
                <td>${updated}</td>
                <td>${available}</td>
                <td>${this.formatAge(deployment.metadata.creationTimestamp)}</td>
                <td>
                    <button class="btn action-btn view" onclick="dashboard.showResourceDetails('${deployment.metadata.name}')" title="View details">👁️</button>
                    <button class="btn action-btn edit" onclick="dashboard.editResource('${deployment.metadata.name}')" title="Edit deployment">✏️</button>
                    <button class="btn action-btn delete" onclick="dashboard.deleteResource('${deployment.metadata.name}')" title="Delete deployment">🗑️</button>
                </td>
            </tr>
        `;
    }

    renderServiceRow(service) {
        const ports = service.spec.ports ? service.spec.ports.map(p => `${p.port}/${p.protocol}`).join(', ') : '';
        const externalIP = service.status.loadBalancer && service.status.loadBalancer.ingress ?
            service.status.loadBalancer.ingress[0].ip || service.status.loadBalancer.ingress[0].hostname : '<none>';

        return `
            <tr data-resource-name="${service.metadata.name}">
                <td><div class="pod-name">${service.metadata.name}</div></td>
                <td>${service.spec.type}</td>
                <td>${service.spec.clusterIP || '<none>'}</td>
                <td>${externalIP}</td>
                <td>${ports}</td>
                <td>${this.formatAge(service.metadata.creationTimestamp)}</td>
                <td>
                    <button class="btn action-btn view" onclick="dashboard.showResourceDetails('${service.metadata.name}')" title="View details">�️</button>
                    <button class="btn action-btn edit" onclick="dashboard.editResource('${service.metadata.name}')" title="Edit service">✏️</button>
                    <button class="btn action-btn delete" onclick="dashboard.deleteResource('${service.metadata.name}')" title="Delete service">�🗑️</button>
                </td>
            </tr>
        `;
    }

    renderConfigMapRow(configmap) {
        const dataCount = configmap.data ? Object.keys(configmap.data).length : 0;

        return `
            <tr data-resource-name="${configmap.metadata.name}">
                <td><div class="pod-name">${configmap.metadata.name}</div></td>
                <td>${dataCount} keys</td>
                <td>${this.formatAge(configmap.metadata.creationTimestamp)}</td>
                <td>
                    <button class="btn action-btn view" onclick="dashboard.showResourceDetails('${configmap.metadata.name}')" title="View details">👁️</button>
                    <button class="btn action-btn edit" onclick="dashboard.editResource('${configmap.metadata.name}')" title="Edit configmap">✏️</button>
                    <button class="btn action-btn delete" onclick="dashboard.deleteResource('${configmap.metadata.name}')" title="Delete configmap">🗑️</button>
                </td>
            </tr>
        `;
    }

    getRestartCount(pod) {
        const containers = pod.status.containerStatuses || [];
        let totalRestarts = 0;
        containers.forEach(container => {
            totalRestarts += container.restartCount || 0;
        });
        return totalRestarts;
    }

    showLogsModal(podName = null) {
        if (podName) {
            this.selectedPod = podName;
        }

        if (!this.selectedPod && this.filteredPods.length > 0) {
            this.selectedPod = this.filteredPods[0].metadata.name;
        }

        if (!this.selectedPod) {
            this.setStatus('No pod selected for logs', 'warning');
            return;
        }

        document.getElementById('logs-title').textContent = `Logs: ${this.selectedPod}`;
        this.showModal('logs-modal');
        this.loadPodLogs();
    }

    async loadPodLogs() {
        const logsContent = document.getElementById('logs-content');
        const containerSelect = document.getElementById('logs-container');
        const followCheckbox = document.getElementById('logs-follow');

        logsContent.textContent = 'Loading logs...';

        try {
            // Get pod details to populate container selector
            const podResponse = await fetch(`${this.apiBase}/api/v1/pods?namespace=${this.currentNamespace}`);
            if (podResponse.ok) {
                const podData = await podResponse.json();
                const pod = podData.pods.find(p => p.metadata.name === this.selectedPod);

                if (pod) {
                    containerSelect.innerHTML = '<option value="">All containers</option>';
                    pod.spec.containers.forEach(container => {
                        containerSelect.innerHTML += `<option value="${container.name}">${container.name}</option>`;
                    });
                }
            }

            const container = containerSelect.value;
            const follow = followCheckbox.checked;

            const logsUrl = `${this.apiBase}/api/v1/pods/${this.currentNamespace}/${this.selectedPod}/logs${container ? `?container=${container}` : ''}`;

            if (follow) {
                // For follow mode, we'll poll periodically
                this.startLogsFollow(logsUrl);
            } else {
                const response = await fetch(logsUrl);
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }

                const logs = await response.text();
                logsContent.textContent = logs || 'No logs available';
            }

        } catch (error) {
            console.error('Failed to load logs:', error);
            logsContent.textContent = `Error loading logs: ${error.message}`;
        }
    }

    startLogsFollow(logsUrl) {
        const logsContent = document.getElementById('logs-content');

        const followLogs = async () => {
            try {
                const response = await fetch(logsUrl);
                if (response.ok) {
                    const logs = await response.text();
                    logsContent.textContent = logs || 'No logs available';
                }
            } catch (error) {
                console.error('Failed to follow logs:', error);
            }
        };

        // Initial load
        followLogs();

        // Follow every 2 seconds
        this.logsFollowInterval = setInterval(followLogs, 2000);
    }

    stopLogsFollow() {
        if (this.logsFollowInterval) {
            clearInterval(this.logsFollowInterval);
            this.logsFollowInterval = null;
        }
    }

    formatAge(timestamp) {
        if (!timestamp) return 'Unknown';

        const now = new Date();
        const created = new Date(timestamp);
        const diff = now - created;

        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);

        if (days > 0) return `${days}d`;
        if (hours > 0) return `${hours}h`;
        if (minutes > 0) return `${minutes}m`;
        return '<1m';
    }

    showCreateModal() {
        document.getElementById('modal-title').textContent = 'Create Pod';
        document.getElementById('pod-form').reset();
        document.getElementById('pod-namespace').value = this.currentNamespace;
        this.showModal('pod-modal');
        document.getElementById('pod-name').focus();
    }

    editPod(podName) {
        // For now, just show create modal - in a full implementation,
        // you'd fetch the pod details and populate the form
        this.showCreateModal();
        document.getElementById('modal-title').textContent = `Edit Pod: ${podName}`;
        // TODO: Load existing pod data
    }

    async savePod() {
        const formData = new FormData(document.getElementById('pod-form'));
        const podData = {
            apiVersion: 'v1',
            kind: 'Pod',
            metadata: {
                name: document.getElementById('pod-name').value,
                namespace: document.getElementById('pod-namespace').value
            },
            spec: {
                containers: [{
                    name: 'app',
                    image: document.getElementById('pod-image').value,
                    ports: [{
                        containerPort: parseInt(document.getElementById('container-port').value) || 80
                    }]
                }]
            }
        };

        try {
            const response = await fetch(`${this.apiBase}/api/v1/pods/${podData.metadata.namespace}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(podData)
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            this.hideModal('pod-modal');
            this.loadResources();
            this.setStatus('Resource created successfully', 'success');

        } catch (error) {
            console.error('Failed to create pod:', error);
            alert(`Failed to create pod: ${error.message}`);
        }
    }

    deletePod(podName) {
        document.getElementById('delete-pod-name').textContent = podName;
        this.selectedPod = podName;
        this.showModal('delete-modal');
    }

    async confirmDelete() {
        if (!this.selectedPod) return;

        try {
            const response = await fetch(`${this.apiBase}/api/v1/pods/${this.currentNamespace}/${this.selectedPod}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            this.hideModal('delete-modal');
            this.loadResources();
            this.setStatus('Resource deleted successfully', 'success');
            this.selectedResource = null;

        } catch (error) {
            console.error('Failed to delete pod:', error);
            alert(`Failed to delete pod: ${error.message}`);
        }
    }

    async showResourceDetails(resourceName) {
        document.getElementById('details-title').textContent = `${this.currentResource.charAt(0).toUpperCase() + this.currentResource.slice(1, -1)}: ${resourceName}`;
        document.getElementById('pod-details').textContent = 'Loading...';
        this.showModal('details-modal');

        try {
            const response = await fetch(`${this.apiBase}/api/v1/${this.currentResource}?namespace=${this.currentNamespace}`);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            const resource = data[this.currentResource].find(r => r.metadata.name === resourceName);

            if (resource) {
                document.getElementById('pod-details').textContent = JSON.stringify(resource, null, 2);
            } else {
                document.getElementById('pod-details').textContent = 'Resource not found';
            }

        } catch (error) {
            console.error('Failed to load resource details:', error);
            document.getElementById('pod-details').textContent = `Error loading details: ${error.message}`;
        }
    }

    connectWebSocket() {
        if (this.websocket) {
            this.websocket.close();
        }

        // Only connect WebSocket for pods (other resources don't have watch endpoints yet)
        if (this.currentResource !== 'pods') {
            this.updateConnectionStatus('disconnected', 'Real-time updates not available');
            return;
        }

        this.updateConnectionStatus('connecting', 'Connecting...');

        const wsUrl = `ws://${window.location.host}/api/v1/${this.currentResource}/watch?namespace=${this.currentNamespace}`;
        this.websocket = new WebSocket(wsUrl);

        this.websocket.onopen = () => {
            console.log('WebSocket connected');
            this.updateConnectionStatus('connected', 'Real-time updates active');
        };

        this.websocket.onmessage = (event) => {
            try {
                const change = JSON.parse(event.data);
                console.log('Resource change:', change);
                this.handleResourceChange(change);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        this.websocket.onclose = () => {
            console.log('WebSocket disconnected');
            this.updateConnectionStatus('disconnected', 'Real-time updates disconnected');
            // Auto-reconnect after 5 seconds
            setTimeout(() => this.connectWebSocket(), 5000);
        };

        this.websocket.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus('disconnected', 'Connection error');
        };
    }

    updateConnectionStatus(status, text) {
        const dot = document.getElementById('connection-dot');
        const textEl = document.getElementById('connection-text');

        dot.className = `connection-dot ${status}`;
        textEl.textContent = text;
    }

    reconnectWebSocket() {
        if (this.websocket) {
            this.websocket.close();
        }
        this.connectWebSocket();
    }

    handleResourceChange(change) {
        // Simple refresh for now - in a more sophisticated implementation,
        // you could update the table incrementally
        console.log('Resource change detected, refreshing...');
        this.loadResources();
    }

    showModal(modalId) {
        document.getElementById(modalId).classList.add('show');
    }

    hideModal(modalId) {
        document.getElementById(modalId).classList.remove('show');
    }

    setStatus(message, type = 'info') {
        const statusEl = document.getElementById('status');
        statusEl.textContent = message;
        statusEl.className = `status-${type}`;

        // Auto-clear status after 5 seconds for success messages
        if (type === 'success') {
            setTimeout(() => {
                statusEl.textContent = 'Ready';
                statusEl.className = '';
            }, 5000);
        }
    }

    toggleTheme() {
        this.theme = this.theme === 'light' ? 'dark' : 'light';
        localStorage.setItem('theme', this.theme);
        this.applyTheme();
    }

    applyTheme() {
        document.documentElement.setAttribute('data-theme', this.theme);
        const themeBtn = document.getElementById('theme-toggle');
        themeBtn.textContent = this.theme === 'light' ? '🌙' : '☀️';
        themeBtn.title = `Switch to ${this.theme === 'light' ? 'dark' : 'light'} theme`;
    }

    showMetricsPanel() {
        document.getElementById('metrics-panel').style.display = 'block';
        this.loadMetrics();
    }

    hideMetricsPanel() {
        document.getElementById('metrics-panel').style.display = 'none';
    }

    async loadMetrics() {
        try {
            // Load cluster metrics
            const clusterResponse = await fetch(`${this.apiBase}/api/v1/metrics/cluster`);
            if (clusterResponse.ok) {
                const clusterData = await clusterResponse.json();
                this.updateClusterMetrics(clusterData);
            }

            // Load namespace metrics
            const nsResponse = await fetch(`${this.apiBase}/api/v1/metrics/namespace/${this.currentNamespace}`);
            if (nsResponse.ok) {
                const nsData = await nsResponse.json();
                this.updateNamespaceMetrics(nsData);
            }
        } catch (error) {
            console.error('Failed to load metrics:', error);
        }
    }

    updateClusterMetrics(data) {
        document.getElementById('nodes-count').textContent = data.cluster.nodes;
        document.getElementById('total-pods').textContent = data.cluster.pods;
        document.getElementById('namespaces-count').textContent = data.cluster.namespaces;
        document.getElementById('running-pods').textContent = data.pod_status.running;
        document.getElementById('pending-pods').textContent = data.pod_status.pending;
        document.getElementById('failed-pods').textContent = data.pod_status.failed;
    }

    updateNamespaceMetrics(data) {
        document.getElementById('current-namespace-metrics').textContent = data.namespace;
        document.getElementById('ns-pods').textContent = data.pods.total;
        document.getElementById('ns-deployments').textContent = data.deployments.total;
        document.getElementById('ns-services').textContent = data.services.total;
    }

    // Cleanup on page unload
    destroy() {
        if (this.websocket) {
            this.websocket.close();
        }
        if (this.autoRefreshInterval) {
            clearInterval(this.autoRefreshInterval);
        }
        this.stopLogsFollow();
    }

    execPod(podName) {
        const command = prompt('Enter command to execute (default: /bin/sh):', '/bin/sh');
        if (!command) return;

        // For now, just show an alert - full exec implementation would require WebSocket
        alert(`Exec functionality for pod '${podName}' with command '${command}' would be implemented here.\n\nThis requires WebSocket connection to stream terminal I/O.`);
    }
}

// Initialize the dashboard when the page loads
const dashboard = new KubernetesDashboard();

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    dashboard.destroy();
});