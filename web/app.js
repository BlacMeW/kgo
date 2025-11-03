// Kubernetes Dashboard Web UI JavaScript

class KubernetesDashboard {
    constructor() {
        this.apiBase = window.location.origin;
        this.currentNamespace = 'default';
        this.websocket = null;
        this.selectedPod = null;
        this.allPods = [];
        this.filteredPods = [];
        this.searchTerm = '';
        this.autoRefreshInterval = null;
        this.logsFollowInterval = null;

        this.init();
    }

    init() {
        this.bindEvents();
        this.loadPods();
        this.connectWebSocket();
        this.setupKeyboardShortcuts();
        this.startAutoRefresh();
    }

    bindEvents() {
        // Namespace selector
        document.getElementById('namespace').addEventListener('change', (e) => {
            this.currentNamespace = e.target.value;
            this.loadPods();
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
        document.getElementById('refresh-btn').addEventListener('click', () => this.loadPods());
        document.getElementById('create-btn').addEventListener('click', () => this.showCreateModal());
        document.getElementById('logs-btn').addEventListener('click', () => this.showLogsModal());

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
                    this.loadPods();
                    break;
                case 'r':
                case 'R':
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        this.loadPods();
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

    startAutoRefresh() {
        // Auto-refresh every 30 seconds
        this.autoRefreshInterval = setInterval(() => {
            if (!this.websocket || this.websocket.readyState !== WebSocket.OPEN) {
                this.loadPods();
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
            this.filteredPods = [...this.allPods];
            document.getElementById('clear-search').style.display = 'none';
            document.getElementById('filtered-count').style.display = 'none';
        } else {
            this.filteredPods = this.allPods.filter(pod =>
                pod.metadata.name.toLowerCase().includes(this.searchTerm) ||
                (pod.spec.nodeName && pod.spec.nodeName.toLowerCase().includes(this.searchTerm)) ||
                pod.status.phase.toLowerCase().includes(this.searchTerm)
            );
            document.getElementById('clear-search').style.display = 'inline-block';
            document.getElementById('filtered-count').textContent = `(${this.filteredPods.length} filtered)`;
            document.getElementById('filtered-count').style.display = 'inline';
        }

        this.renderPods(this.filteredPods);
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

    async loadPods() {
        const refreshBtn = document.getElementById('refresh-btn');
        const spinner = refreshBtn.querySelector('.spinner');

        refreshBtn.disabled = true;
        spinner.style.display = 'inline-block';

        this.setStatus('Loading pods...', 'loading');

        try {
            const response = await fetch(`${this.apiBase}/api/v1/pods?namespace=${this.currentNamespace}`);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            this.allPods = data.pods || [];
            this.applyFilters();
            this.setStatus('Pods loaded successfully', 'success');

        } catch (error) {
            console.error('Failed to load pods:', error);
            this.setStatus(`Error loading pods: ${error.message}`, 'error');
            this.allPods = [];
            this.applyFilters();
        } finally {
            refreshBtn.disabled = false;
            spinner.style.display = 'none';
        }
    }

    renderPods(pods) {
        const tbody = document.getElementById('pods-body');
        const podCount = document.getElementById('pod-count');
        const logsBtn = document.getElementById('logs-btn');

        podCount.textContent = `${this.allPods.length} pod${this.allPods.length !== 1 ? 's' : ''}`;

        // Show/hide logs button based on pod selection
        logsBtn.style.display = pods.length > 0 ? 'inline-block' : 'none';

        if (pods.length === 0) {
            const colspan = this.searchTerm ? 7 : 7;
            const message = this.searchTerm ? 'No pods match your search' : 'No pods found';
            tbody.innerHTML = `<tr><td colspan="${colspan}" class="loading">${message}</td></tr>`;
            return;
        }

        tbody.innerHTML = pods.map(pod => `
            <tr data-pod-name="${pod.metadata.name}">
                <td>
                    <div class="pod-name">${pod.metadata.name}</div>
                </td>
                <td><span class="status-${pod.status.phase.toLowerCase()}">${pod.status.phase}</span></td>
                <td>${this.getReadyCount(pod)}</td>
                <td>${this.getRestartCount(pod)}</td>
                <td>${this.formatAge(pod.metadata.creationTimestamp)}</td>
                <td>${pod.spec.nodeName || '<none>'}</td>
                <td>
                    <button class="btn action-btn view" onclick="dashboard.showPodDetails('${pod.metadata.name}')" title="View details">👁️</button>
                    <button class="btn action-btn edit" onclick="dashboard.editPod('${pod.metadata.name}')" title="Edit pod">✏️</button>
                    <button class="btn action-btn logs" onclick="dashboard.showLogsModal('${pod.metadata.name}')" title="View logs">📋</button>
                    <button class="btn action-btn delete" onclick="dashboard.deletePod('${pod.metadata.name}')" title="Delete pod">🗑️</button>
                </td>
            </tr>
        `).join('');
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
            this.loadPods();
            this.setStatus('Pod created successfully', 'success');

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
            this.loadPods();
            this.setStatus('Pod deleted successfully', 'success');
            this.selectedPod = null;

        } catch (error) {
            console.error('Failed to delete pod:', error);
            alert(`Failed to delete pod: ${error.message}`);
        }
    }

    async showPodDetails(podName) {
        document.getElementById('details-title').textContent = `Pod: ${podName}`;
        document.getElementById('pod-details').textContent = 'Loading...';
        this.showModal('details-modal');

        try {
            const response = await fetch(`${this.apiBase}/api/v1/pods?namespace=${this.currentNamespace}`);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            const pod = data.pods.find(p => p.metadata.name === podName);

            if (pod) {
                document.getElementById('pod-details').textContent = JSON.stringify(pod, null, 2);
            } else {
                document.getElementById('pod-details').textContent = 'Pod not found';
            }

        } catch (error) {
            console.error('Failed to load pod details:', error);
            document.getElementById('pod-details').textContent = `Error: ${error.message}`;
        }
    }

    connectWebSocket() {
        if (this.websocket) {
            this.websocket.close();
        }

        this.updateConnectionStatus('connecting', 'Connecting...');

        const wsUrl = `ws://${window.location.host}/api/v1/pods/watch?namespace=${this.currentNamespace}`;
        this.websocket = new WebSocket(wsUrl);

        this.websocket.onopen = () => {
            console.log('WebSocket connected');
            this.updateConnectionStatus('connected', 'Real-time updates active');
        };

        this.websocket.onmessage = (event) => {
            try {
                const change = JSON.parse(event.data);
                console.log('Pod change:', change);
                this.handlePodChange(change);
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

    handlePodChange(change) {
        // Simple refresh for now - in a more sophisticated implementation,
        // you could update the table incrementally
        console.log('Pod change detected, refreshing...');
        this.loadPods();
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
}

// Initialize the dashboard when the page loads
const dashboard = new KubernetesDashboard();

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    dashboard.destroy();
});