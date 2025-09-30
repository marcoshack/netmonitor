// Main Application JavaScript Entry Point

import { APIClient } from './api.js';
import { ThemeManager } from './theme.js';

class NetMonitorApp {
    constructor() {
        this.api = new APIClient();
        this.themeManager = new ThemeManager();
        this.currentView = 'overview';
        this.isInitialized = false;
        this.monitoringData = {
            systemInfo: null,
            config: null,
            status: null
        };
    }

    async init() {
        try {
            console.log('Initializing NetMonitor application...');
            
            // Initialize theme first
            await this.themeManager.init();
            
            // Initialize API client
            await this.api.init();
            
            // Setup UI
            this.setupUI();
            this.bindEvents();
            
            // Load initial data
            await this.loadInitialData();
            
            this.isInitialized = true;
            console.log('NetMonitor application initialized successfully');
            
        } catch (error) {
            console.error('Failed to initialize NetMonitor application:', error);
            this.showError('Failed to initialize application: ' + error.message);
        }
    }

    setupUI() {
        this.renderMainLayout();
        this.renderNavigation();
        this.renderContent();
    }

    renderMainLayout() {
        const app = document.getElementById('app');
        app.innerHTML = `
            <div class="app">
                <header class="app-header">
                    <h1 class="app-title">NetMonitor</h1>
                    <div class="header-controls">
                        <button class="theme-toggle" id="themeToggle" aria-label="Toggle theme" title="Toggle theme">
                            🌙
                        </button>
                        <button class="btn-icon" id="settingsBtn" aria-label="Settings" title="Settings">
                            ⚙️
                        </button>
                    </div>
                </header>

                <div class="app-main">
                    <nav class="sidebar">
                        <ul class="nav-menu" id="navMenu">
                            <li class="nav-item">
                                <a href="#overview" class="nav-link active" data-view="overview">📊 Overview</a>
                            </li>
                            <li class="nav-item">
                                <a href="#regions" class="nav-link" data-view="regions">🌍 Regions</a>
                            </li>
                            <li class="nav-item">
                                <a href="#endpoints" class="nav-link" data-view="endpoints">🎯 Endpoints</a>
                            </li>
                            <li class="nav-item">
                                <a href="#manual" class="nav-link" data-view="manual">⚡ Manual Tests</a>
                            </li>
                            <li class="nav-item">
                                <a href="#settings" class="nav-link" data-view="settings">⚙️ Settings</a>
                            </li>
                        </ul>
                    </nav>

                    <main class="content" id="mainContent">
                        <!-- Dynamic content will be loaded here -->
                    </main>
                </div>

                <footer class="app-footer">
                    <div class="status-bar">
                        <span class="connection-status" id="connectionStatus">
                            <span class="status-indicator status-healthy">Connected</span>
                        </span>
                        <span class="last-update" id="lastUpdate">Last update: Just now</span>
                    </div>
                </footer>
            </div>
        `;
    }

    renderNavigation() {
        // Navigation is already rendered in the main layout
        // This method can be used for dynamic navigation updates
    }

    renderContent() {
        this.showView(this.currentView);
    }

    bindEvents() {
        // Theme toggle
        const themeToggle = document.getElementById('themeToggle');
        themeToggle.addEventListener('click', () => {
            this.themeManager.toggleTheme();
        });

        // Settings button
        const settingsBtn = document.getElementById('settingsBtn');
        settingsBtn.addEventListener('click', () => {
            this.showView('settings');
        });

        // Navigation
        const navMenu = document.getElementById('navMenu');
        navMenu.addEventListener('click', (e) => {
            if (e.target.classList.contains('nav-link')) {
                e.preventDefault();
                const view = e.target.dataset.view;
                this.showView(view);
            }
        });

        // Handle system events
        this.bindSystemEvents();
    }

    bindSystemEvents() {
        // Handle window focus/blur for performance optimization
        window.addEventListener('focus', () => {
            this.onWindowFocus();
        });

        window.addEventListener('blur', () => {
            this.onWindowBlur();
        });

        // Handle online/offline status
        window.addEventListener('online', () => {
            this.updateConnectionStatus(true);
        });

        window.addEventListener('offline', () => {
            this.updateConnectionStatus(false);
        });
    }

    showView(viewName) {
        // Update navigation
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            link.classList.remove('active');
            if (link.dataset.view === viewName) {
                link.classList.add('active');
            }
        });

        // Update content
        const content = document.getElementById('mainContent');
        this.currentView = viewName;

        switch (viewName) {
            case 'overview':
                this.renderOverview(content);
                break;
            case 'regions':
                this.renderRegions(content);
                break;
            case 'endpoints':
                this.renderEndpoints(content);
                break;
            case 'manual':
                this.renderManualTests(content);
                break;
            case 'settings':
                this.renderSettings(content);
                break;
            default:
                this.renderOverview(content);
        }
    }

    renderOverview(container) {
        container.innerHTML = `
            <div class="overview">
                <div class="grid grid-cols-4">
                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">System Status</h3>
                            <div class="status-indicator status-healthy" id="systemStatusIndicator">●</div>
                        </div>
                        <div class="card-content">
                            <div class="status-grid">
                                <div class="status-item">
                                    <span class="status-label">Application:</span>
                                    <span class="status-value" id="applicationStatus">Running</span>
                                </div>
                                <div class="status-item">
                                    <span class="status-label">Monitoring:</span>
                                    <span class="status-value" id="monitoringStatus">Loading...</span>
                                </div>
                                <div class="status-item">
                                    <span class="status-label">Version:</span>
                                    <span class="status-value" id="appVersion">1.0.0</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">Endpoints</h3>
                            <button class="btn-icon" id="refreshEndpoints" title="Refresh">↻</button>
                        </div>
                        <div class="card-content">
                            <div class="endpoint-summary" id="endpointSummary">
                                <div class="summary-stat">
                                    <div class="stat-number status-healthy" id="totalEndpoints">0</div>
                                    <div class="stat-label">Total</div>
                                </div>
                                <div class="summary-stat">
                                    <div class="stat-number status-healthy" id="healthyEndpoints">0</div>
                                    <div class="stat-label">Healthy</div>
                                </div>
                                <div class="summary-stat">
                                    <div class="stat-number status-warning" id="warningEndpoints">0</div>
                                    <div class="stat-label">Warning</div>
                                </div>
                                <div class="summary-stat">
                                    <div class="stat-number status-danger" id="downEndpoints">0</div>
                                    <div class="stat-label">Down</div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">Quick Actions</h3>
                        </div>
                        <div class="card-content">
                            <div class="action-buttons">
                                <button class="btn btn-primary" id="startMonitoringBtn" disabled>
                                    ▶ Start Monitoring
                                </button>
                                <button class="btn btn-secondary" id="stopMonitoringBtn" disabled>
                                    ⏹ Stop Monitoring
                                </button>
                                <button class="btn btn-secondary" id="runTestBtn" disabled>
                                    🧪 Run Test
                                </button>
                            </div>
                        </div>
                    </div>

                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">Configuration</h3>
                        </div>
                        <div class="card-content">
                            <div class="config-summary" id="configSummary">
                                <div class="config-item">
                                    <span class="config-label">Regions:</span>
                                    <span class="config-value" id="regionCount">0</span>
                                </div>
                                <div class="config-item">
                                    <span class="config-label">Test Interval:</span>
                                    <span class="config-value" id="testInterval">5 minutes</span>
                                </div>
                                <div class="config-item">
                                    <span class="config-label">Data Retention:</span>
                                    <span class="config-value" id="dataRetention">90 days</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        // Bind overview-specific events
        this.bindOverviewEvents();
    }

    renderRegions(container) {
        container.innerHTML = `
            <div class="regions">
                <div class="card">
                    <div class="card-header">
                        <h2 class="card-title">Regional Monitoring</h2>
                    </div>
                    <div class="card-content">
                        <div class="region-list" id="regionList">
                            <!-- Regions will be populated here -->
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.loadRegions();
    }

    renderEndpoints(container) {
        container.innerHTML = `
            <div class="endpoints">
                <div class="card">
                    <div class="card-header">
                        <h2 class="card-title">Endpoint Management</h2>
                        <button class="btn btn-primary" id="addEndpointBtn">+ Add Endpoint</button>
                    </div>
                    <div class="card-content">
                        <div class="endpoint-list" id="endpointList">
                            <!-- Endpoints will be populated here -->
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.loadEndpoints();
    }

    renderManualTests(container) {
        container.innerHTML = `
            <div class="manual-tests">
                <div class="card">
                    <div class="card-header">
                        <h2 class="card-title">Manual Testing</h2>
                    </div>
                    <div class="card-content">
                        <p>Select an endpoint to test:</p>
                        <div class="test-controls" id="testControls">
                            <select id="endpointSelect" class="form-select">
                                <option value="">Select an endpoint...</option>
                            </select>
                            <button class="btn btn-primary" id="runSingleTestBtn" disabled>Run Test</button>
                        </div>
                        <div class="test-results" id="testResults" style="display: none;">
                            <!-- Test results will be displayed here -->
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.loadEndpointOptions();
        this.bindManualTestEvents();
    }

    renderSettings(container) {
        container.innerHTML = `
            <div class="settings">
                <div class="card">
                    <div class="card-header">
                        <h2 class="card-title">Application Settings</h2>
                    </div>
                    <div class="card-content">
                        <div class="settings-grid">
                            <div class="setting-group">
                                <h4>Theme</h4>
                                <div class="theme-options">
                                    <button class="btn btn-secondary theme-btn" data-theme="light">☀️ Light</button>
                                    <button class="btn btn-secondary theme-btn" data-theme="dark">🌙 Dark</button>
                                    <button class="btn btn-secondary theme-btn active" data-theme="auto">🔄 Auto</button>
                                </div>
                            </div>
                            <div class="setting-group">
                                <h4>System Information</h4>
                                <div class="system-info" id="systemInfo">
                                    Loading system information...
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.bindSettingsEvents();
        this.loadSystemInfo();
    }

    bindOverviewEvents() {
        const startBtn = document.getElementById('startMonitoringBtn');
        const stopBtn = document.getElementById('stopMonitoringBtn');
        const runTestBtn = document.getElementById('runTestBtn');
        const refreshBtn = document.getElementById('refreshEndpoints');

        if (startBtn) {
            startBtn.addEventListener('click', async () => {
                await this.startMonitoring();
            });
        }

        if (stopBtn) {
            stopBtn.addEventListener('click', async () => {
                await this.stopMonitoring();
            });
        }

        if (runTestBtn) {
            runTestBtn.addEventListener('click', async () => {
                await this.runQuickTest();
            });
        }

        if (refreshBtn) {
            refreshBtn.addEventListener('click', async () => {
                await this.refreshData();
            });
        }
    }

    bindManualTestEvents() {
        const endpointSelect = document.getElementById('endpointSelect');
        const runTestBtn = document.getElementById('runSingleTestBtn');

        if (endpointSelect) {
            endpointSelect.addEventListener('change', (e) => {
                if (runTestBtn) {
                    runTestBtn.disabled = !e.target.value;
                }
            });
        }

        if (runTestBtn) {
            runTestBtn.addEventListener('click', async () => {
                const endpointId = endpointSelect.value;
                if (endpointId) {
                    await this.runManualTest(endpointId);
                }
            });
        }
    }

    bindSettingsEvents() {
        const themeButtons = document.querySelectorAll('.theme-btn');
        themeButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                const theme = btn.dataset.theme;
                this.themeManager.setTheme(theme);
                
                // Update active button
                themeButtons.forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
            });
        });
    }

    async loadInitialData() {
        try {
            // Load system info
            this.monitoringData.systemInfo = await this.api.getSystemInfo();
            this.updateSystemStatus(this.monitoringData.systemInfo);

            // Load configuration
            this.monitoringData.config = await this.api.getConfiguration();
            this.updateConfigurationDisplay(this.monitoringData.config);

            // Load monitoring status
            this.monitoringData.status = await this.api.getMonitoringStatus();
            this.updateMonitoringStatus(this.monitoringData.status);

        } catch (error) {
            console.error('Failed to load initial data:', error);
            this.showError('Failed to load application data: ' + this.api.formatError(error));
        }
    }

    async loadRegions() {
        if (!this.monitoringData.config) return;

        const regionList = document.getElementById('regionList');
        if (!regionList) return;

        const regions = this.monitoringData.config.regions || {};
        
        regionList.innerHTML = Object.entries(regions).map(([name, region]) => `
            <div class="region-card">
                <h4>${name}</h4>
                <div class="region-stats">
                    <span>Endpoints: ${region.endpoints ? region.endpoints.length : 0}</span>
                    <span>Threshold: ${region.thresholds ? region.thresholds.latency_ms : 0}ms</span>
                </div>
            </div>
        `).join('');
    }

    async loadEndpoints() {
        if (!this.monitoringData.config) return;

        const endpointList = document.getElementById('endpointList');
        if (!endpointList) return;

        const endpoints = [];
        Object.entries(this.monitoringData.config.regions || {}).forEach(([regionName, region]) => {
            if (region.endpoints) {
                region.endpoints.forEach(endpoint => {
                    endpoints.push({
                        ...endpoint,
                        region: regionName,
                        id: `${regionName}-${endpoint.name}`
                    });
                });
            }
        });

        endpointList.innerHTML = endpoints.map(endpoint => `
            <div class="endpoint-card">
                <div class="endpoint-header">
                    <h4>${endpoint.name}</h4>
                    <span class="endpoint-type">${endpoint.type}</span>
                </div>
                <div class="endpoint-details">
                    <div>Region: ${endpoint.region}</div>
                    <div>Address: ${endpoint.address}</div>
                    <div>Timeout: ${endpoint.timeout}ms</div>
                </div>
            </div>
        `).join('');
    }

    async loadEndpointOptions() {
        if (!this.monitoringData.config) return;

        const endpointSelect = document.getElementById('endpointSelect');
        if (!endpointSelect) return;

        const endpoints = [];
        Object.entries(this.monitoringData.config.regions || {}).forEach(([regionName, region]) => {
            if (region.endpoints) {
                region.endpoints.forEach(endpoint => {
                    endpoints.push({
                        id: `${regionName}-${endpoint.name}`,
                        name: endpoint.name,
                        region: regionName
                    });
                });
            }
        });

        endpointSelect.innerHTML = '<option value="">Select an endpoint...</option>' +
            endpoints.map(endpoint => 
                `<option value="${endpoint.id}">${endpoint.name} (${endpoint.region})</option>`
            ).join('');
    }

    async loadSystemInfo() {
        const systemInfoEl = document.getElementById('systemInfo');
        if (!systemInfoEl || !this.monitoringData.systemInfo) return;

        const info = this.monitoringData.systemInfo;
        systemInfoEl.innerHTML = `
            <div class="info-grid">
                <div class="info-item">
                    <span class="info-label">Application:</span>
                    <span class="info-value">${info.applicationName}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Version:</span>
                    <span class="info-value">${info.version}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Build:</span>
                    <span class="info-value">${info.buildTime}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Status:</span>
                    <span class="info-value">${info.running ? 'Running' : 'Stopped'}</span>
                </div>
            </div>
        `;
    }

    async startMonitoring() {
        try {
            this.showLoading('Starting monitoring...');
            await this.api.startMonitoring();
            await this.refreshMonitoringStatus();
            this.hideLoading();
            this.showSuccess('Monitoring started successfully');
        } catch (error) {
            this.hideLoading();
            this.showError('Failed to start monitoring: ' + this.api.formatError(error));
        }
    }

    async stopMonitoring() {
        try {
            this.showLoading('Stopping monitoring...');
            await this.api.stopMonitoring();
            await this.refreshMonitoringStatus();
            this.hideLoading();
            this.showSuccess('Monitoring stopped successfully');
        } catch (error) {
            this.hideLoading();
            this.showError('Failed to stop monitoring: ' + this.api.formatError(error));
        }
    }

    async runQuickTest() {
        try {
            if (!this.monitoringData.config || !this.monitoringData.config.regions) {
                this.showError('No endpoints configured for testing');
                return;
            }

            // Get first endpoint for quick test
            const firstRegion = Object.values(this.monitoringData.config.regions)[0];
            if (!firstRegion || !firstRegion.endpoints || firstRegion.endpoints.length === 0) {
                this.showError('No endpoints available for testing');
                return;
            }

            const firstEndpoint = firstRegion.endpoints[0];
            const endpointId = `${Object.keys(this.monitoringData.config.regions)[0]}-${firstEndpoint.name}`;

            this.showLoading('Running quick test...');
            const result = await this.api.runManualTest(endpointId);
            this.hideLoading();
            
            this.showSuccess(`Test completed: ${result.status} (${result.latency}ms)`);
        } catch (error) {
            this.hideLoading();
            this.showError('Quick test failed: ' + this.api.formatError(error));
        }
    }

    async runManualTest(endpointId) {
        try {
            this.showLoading('Running test...');
            const result = await this.api.runManualTest(endpointId);
            this.hideLoading();

            // Display test results
            const testResults = document.getElementById('testResults');
            if (testResults) {
                testResults.style.display = 'block';
                testResults.innerHTML = `
                    <div class="test-result-card">
                        <h4>Test Result</h4>
                        <div class="result-details">
                            <div class="result-item">
                                <span class="result-label">Status:</span>
                                <span class="result-value status-${result.status}">${result.status}</span>
                            </div>
                            <div class="result-item">
                                <span class="result-label">Latency:</span>
                                <span class="result-value">${result.latency}ms</span>
                            </div>
                            <div class="result-item">
                                <span class="result-label">Timestamp:</span>
                                <span class="result-value">${new Date(result.timestamp).toLocaleString()}</span>
                            </div>
                        </div>
                    </div>
                `;
            }

            this.showSuccess('Test completed successfully');
        } catch (error) {
            this.hideLoading();
            this.showError('Test failed: ' + this.api.formatError(error));
        }
    }

    async refreshData() {
        try {
            await this.loadInitialData();
            this.updateLastRefreshTime();
            
            // Refresh current view
            if (this.currentView === 'regions') {
                this.loadRegions();
            } else if (this.currentView === 'endpoints') {
                this.loadEndpoints();
            }
        } catch (error) {
            console.error('Failed to refresh data:', error);
            this.showError('Failed to refresh data: ' + this.api.formatError(error));
        }
    }

    async refreshMonitoringStatus() {
        try {
            this.monitoringData.status = await this.api.getMonitoringStatus();
            this.updateMonitoringStatus(this.monitoringData.status);
        } catch (error) {
            console.error('Failed to refresh monitoring status:', error);
        }
    }

    updateSystemStatus(systemInfo) {
        const appVersionEl = document.getElementById('appVersion');
        const applicationStatusEl = document.getElementById('applicationStatus');

        if (appVersionEl) {
            appVersionEl.textContent = systemInfo.version;
        }

        if (applicationStatusEl) {
            applicationStatusEl.textContent = systemInfo.running ? 'Running' : 'Stopped';
        }
    }

    updateConfigurationDisplay(config) {
        if (!config) return;

        const regionCountEl = document.getElementById('regionCount');
        const testIntervalEl = document.getElementById('testInterval');
        const dataRetentionEl = document.getElementById('dataRetention');
        const totalEndpointsEl = document.getElementById('totalEndpoints');

        if (regionCountEl) {
            regionCountEl.textContent = Object.keys(config.regions || {}).length;
        }

        if (config.settings) {
            if (testIntervalEl) {
                testIntervalEl.textContent = `${config.settings.test_interval_seconds} seconds`;
            }
            if (dataRetentionEl) {
                dataRetentionEl.textContent = `${config.settings.data_retention_days} days`;
            }
        }

        // Count total endpoints
        let total = 0;
        Object.values(config.regions || {}).forEach(region => {
            if (region.endpoints) {
                total += region.endpoints.length;
            }
        });

        if (totalEndpointsEl) {
            totalEndpointsEl.textContent = total;
        }
    }

    updateMonitoringStatus(status) {
        const monitoringStatusEl = document.getElementById('monitoringStatus');
        const startBtn = document.getElementById('startMonitoringBtn');
        const stopBtn = document.getElementById('stopMonitoringBtn');
        const runTestBtn = document.getElementById('runTestBtn');

        if (monitoringStatusEl) {
            monitoringStatusEl.textContent = status.running ? 'Running' : 'Stopped';
        }

        if (startBtn) {
            startBtn.disabled = status.running;
        }

        if (stopBtn) {
            stopBtn.disabled = !status.running;
        }

        if (runTestBtn) {
            runTestBtn.disabled = false; // Manual tests can always be run
        }
    }

    updateConnectionStatus(online) {
        const connectionStatus = document.getElementById('connectionStatus');
        if (connectionStatus) {
            const indicator = connectionStatus.querySelector('.status-indicator');
            if (online) {
                indicator.className = 'status-indicator status-healthy';
                indicator.textContent = 'Connected';
            } else {
                indicator.className = 'status-indicator status-danger';
                indicator.textContent = 'Disconnected';
            }
        }
    }

    updateLastRefreshTime() {
        const lastUpdate = document.getElementById('lastUpdate');
        if (lastUpdate) {
            lastUpdate.textContent = `Last update: ${new Date().toLocaleTimeString()}`;
        }
    }

    onWindowFocus() {
        // Resume real-time updates when window gains focus
        if (this.isInitialized) {
            this.refreshData();
        }
    }

    onWindowBlur() {
        // Optionally pause some updates when window loses focus
    }

    showLoading(message = 'Loading...') {
        // Create simple loading overlay
        const existing = document.getElementById('loadingOverlay');
        if (existing) existing.remove();

        const overlay = document.createElement('div');
        overlay.id = 'loadingOverlay';
        overlay.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0,0,0,0.5);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 9999;
            color: white;
            font-size: 1.1rem;
        `;
        overlay.innerHTML = `<div><div class="loading"></div><br>${message}</div>`;
        document.body.appendChild(overlay);
    }

    hideLoading() {
        const overlay = document.getElementById('loadingOverlay');
        if (overlay) {
            overlay.remove();
        }
    }

    showError(message) {
        console.error('Error:', message);
        // Simple alert for now - will be replaced with proper UI later
        alert('Error: ' + message);
    }

    showSuccess(message) {
        console.log('Success:', message);
        // Simple alert for now - will be replaced with proper UI later
        alert('Success: ' + message);
    }
}

// Initialize application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.netMonitorApp = new NetMonitorApp();
    window.netMonitorApp.init();
});

// Export for module use
export { NetMonitorApp };