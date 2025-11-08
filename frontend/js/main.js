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
            <a href="#mainContent" class="skip-to-main">Skip to main content</a>
            <div class="app" role="application">
                <header class="app-header" role="banner">
                    <h1 class="app-title">NetMonitor</h1>
                    <div class="header-controls">
                        <button class="theme-toggle" id="themeToggle" aria-label="Toggle theme" aria-pressed="false" title="Toggle dark/light theme">
                            <span aria-hidden="true">Theme</span>
                        </button>
                        <button class="btn-icon" id="settingsBtn" aria-label="Open settings" title="Settings">
                            <span aria-hidden="true">Settings</span>
                        </button>
                    </div>
                </header>

                <div class="app-main">
                    <nav class="sidebar" role="navigation" aria-label="Main navigation">
                        <ul class="nav-menu" id="navMenu" role="menubar">
                            <li class="nav-item" role="none">
                                <a href="#overview" class="nav-link active" data-view="overview" role="menuitem" aria-current="page">Overview</a>
                            </li>
                            <li class="nav-item" role="none">
                                <a href="#regions" class="nav-link" data-view="regions" role="menuitem">Regions</a>
                            </li>
                            <li class="nav-item" role="none">
                                <a href="#endpoints" class="nav-link" data-view="endpoints" role="menuitem">Endpoints</a>
                            </li>
                            <li class="nav-item" role="none">
                                <a href="#manual" class="nav-link" data-view="manual" role="menuitem">Manual Tests</a>
                            </li>
                            <li class="nav-item" role="none">
                                <a href="#settings" class="nav-link" data-view="settings" role="menuitem">Settings</a>
                            </li>
                        </ul>
                    </nav>

                    <main class="content" id="mainContent" role="main" aria-live="polite">
                        <!-- Dynamic content will be loaded here -->
                    </main>
                </div>

                <footer class="app-footer" role="contentinfo">
                    <div class="status-bar">
                        <span class="connection-status" id="connectionStatus" role="status" aria-live="polite">
                            <span class="status-indicator status-healthy">Connected</span>
                        </span>
                        <span class="last-update" id="lastUpdate" role="status" aria-live="polite">Last update: Just now</span>
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
            link.removeAttribute('aria-current');
            if (link.dataset.view === viewName) {
                link.classList.add('active');
                link.setAttribute('aria-current', 'page');
            }
        });

        // Update content
        const content = document.getElementById('mainContent');
        this.currentView = viewName;

        // Set page title for screen readers
        document.title = `NetMonitor - ${viewName.charAt(0).toUpperCase() + viewName.slice(1)}`;

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
                <!-- First row: System Status and Endpoint Summary -->
                <div class="grid grid-cols-2">
                    <div class="card" id="systemStatusWidget" role="region" aria-label="System Status">
                        <div class="card-header">
                            <h3 class="card-title">System Status</h3>
                            <div class="status-indicator status-healthy" id="systemStatusIndicator" aria-label="System health indicator"></div>
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
                                <div class="status-item">
                                    <span class="status-label">Last Test:</span>
                                    <span class="status-value" id="lastTestTime">Never</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="card" id="endpointSummaryWidget" role="region" aria-label="Endpoint Summary">
                        <div class="card-header">
                            <h3 class="card-title">Endpoints</h3>
                            <button class="refresh-btn" id="refreshEndpoints" aria-label="Refresh endpoint data" title="Refresh">↻</button>
                        </div>
                        <div class="card-content">
                            <div class="summary-stats">
                                <div class="summary-stat">
                                    <div class="stat-number" id="totalEndpoints">0</div>
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
                </div>

                <!-- Second row: Regional Health and Latest Results -->
                <div class="grid grid-cols-2">
                    <div class="card" id="regionalHealthWidget" role="region" aria-label="Regional Health">
                        <div class="card-header">
                            <h3 class="card-title">Regional Health</h3>
                        </div>
                        <div class="card-content">
                            <div class="region-list" id="regionalHealthList">
                                <div class="skeleton skeleton-text"></div>
                                <div class="skeleton skeleton-text"></div>
                                <div class="skeleton skeleton-text"></div>
                            </div>
                        </div>
                    </div>

                    <div class="card" id="latestResultsWidget" role="region" aria-label="Latest Test Results">
                        <div class="card-header">
                            <h3 class="card-title">Latest Results</h3>
                        </div>
                        <div class="card-content">
                            <div class="latest-results" id="latestResultsList">
                                <div class="skeleton skeleton-text"></div>
                                <div class="skeleton skeleton-text"></div>
                                <div class="skeleton skeleton-text"></div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Third row: Performance Metrics and Quick Actions -->
                <div class="grid grid-cols-2">
                    <div class="card" id="performanceMetricsWidget" role="region" aria-label="Performance Metrics">
                        <div class="card-header">
                            <h3 class="card-title">Performance Metrics</h3>
                        </div>
                        <div class="card-content">
                            <div class="status-grid">
                                <div class="status-item">
                                    <span class="status-label">CPU Usage:</span>
                                    <span class="status-value metric-value" id="cpuUsage">--</span>
                                </div>
                                <div class="status-item">
                                    <span class="status-label">Memory Usage:</span>
                                    <span class="status-value metric-value" id="memoryUsage">--</span>
                                </div>
                                <div class="status-item">
                                    <span class="status-label">Disk Usage:</span>
                                    <span class="status-value metric-value" id="diskUsage">--</span>
                                </div>
                                <div class="status-item">
                                    <span class="status-label">Network Usage:</span>
                                    <span class="status-value metric-value" id="networkUsage">--</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="card" role="region" aria-label="Quick Actions">
                        <div class="card-header">
                            <h3 class="card-title">Quick Actions</h3>
                        </div>
                        <div class="card-content">
                            <div class="action-buttons">
                                <button class="btn btn-primary" id="startMonitoringBtn" disabled aria-label="Start monitoring">
                                    Start Monitoring
                                </button>
                                <button class="btn btn-secondary" id="stopMonitoringBtn" disabled aria-label="Stop monitoring">
                                    Stop Monitoring
                                </button>
                                <button class="btn btn-secondary" id="runTestBtn" disabled aria-label="Run quick test">
                                    Run Test
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        // Initialize widgets and bind events
        this.initializeOverviewWidgets();
        this.bindOverviewEvents();
    }

    initializeOverviewWidgets() {
        // Import and initialize widgets dynamically
        // In a production app, we would use proper ES6 imports
        // For now, we'll manually update the widgets using the existing methods

        // Initialize System Status Widget
        this.updateSystemStatusWidget();

        // Initialize Endpoint Summary Widget
        this.updateEndpointSummaryWidget();

        // Initialize Regional Health Widget
        this.updateRegionalHealthWidget();

        // Initialize Latest Results Widget
        this.updateLatestResultsWidget();

        // Initialize Performance Metrics Widget
        this.updatePerformanceMetricsWidget();

        // Set up auto-update interval (30 seconds)
        if (this.widgetUpdateInterval) {
            clearInterval(this.widgetUpdateInterval);
        }
        this.widgetUpdateInterval = setInterval(() => {
            if (this.currentView === 'overview') {
                this.updateSystemStatusWidget();
                this.updateEndpointSummaryWidget();
                this.updateRegionalHealthWidget();
                this.updateLatestResultsWidget();
                this.updatePerformanceMetricsWidget();
            }
        }, 30000);
    }

    async updateSystemStatusWidget() {
        try {
            const systemInfo = this.monitoringData.systemInfo || await this.api.getSystemInfo();
            const status = this.monitoringData.status || await this.api.getMonitoringStatus();

            // Update status indicator
            const indicator = document.getElementById('systemStatusIndicator');
            if (indicator) {
                indicator.className = 'status-indicator';
                if (status.running && systemInfo.running) {
                    indicator.classList.add('status-healthy');
                    indicator.setAttribute('aria-label', 'System status: healthy');
                } else {
                    indicator.classList.add('status-warning');
                    indicator.setAttribute('aria-label', 'System status: warning');
                }
            }

            // Update last test time
            if (status.lastTestTime) {
                const lastTestEl = document.getElementById('lastTestTime');
                if (lastTestEl) {
                    lastTestEl.textContent = this.formatTimeAgo(new Date(status.lastTestTime));
                }
            }
        } catch (error) {
            console.error('Failed to update system status widget:', error);
        }
    }

    async updateEndpointSummaryWidget() {
        try {
            const config = this.monitoringData.config || await this.api.getConfiguration();

            // Calculate endpoint stats (simplified)
            let total = 0;
            let healthy = 0;
            let warning = 0;
            let down = 0;

            if (config && config.regions) {
                Object.values(config.regions).forEach(region => {
                    if (region.endpoints) {
                        total += region.endpoints.length;
                        // Mock health calculation
                        healthy += Math.floor(region.endpoints.length * 0.8);
                        warning += Math.floor(region.endpoints.length * 0.15);
                        down += region.endpoints.length - Math.floor(region.endpoints.length * 0.95);
                    }
                });
            }

            this.updateElement('#totalEndpoints', total);
            this.updateElement('#healthyEndpoints', healthy);
            this.updateElement('#warningEndpoints', warning);
            this.updateElement('#downEndpoints', down);
        } catch (error) {
            console.error('Failed to update endpoint summary widget:', error);
        }
    }

    async updateRegionalHealthWidget() {
        try {
            const config = this.monitoringData.config || await this.api.getConfiguration();
            const container = document.getElementById('regionalHealthList');
            if (!container || !config || !config.regions) return;

            const regions = Object.entries(config.regions).map(([name, region]) => {
                // Calculate mock health status
                const avgLatency = Math.floor(Math.random() * 200) + 20;
                const uptime = (99 + Math.random()).toFixed(1);
                const status = avgLatency < 100 ? 'healthy' : avgLatency < 150 ? 'warning' : 'critical';

                return { name, avgLatency, uptime, status };
            });

            container.innerHTML = regions.map(r => `
                <div class="region-item" data-status="${r.status}" data-region="${r.name}" role="button" tabindex="0" aria-label="View ${r.name} region details">
                    <div class="region-name">${r.name}</div>
                    <div class="region-stats">
                        <span class="avg-latency" title="Average latency">${r.avgLatency}ms</span>
                        <span class="uptime" title="Uptime">${r.uptime}%</span>
                    </div>
                    <div class="region-indicator ${r.status}" aria-label="Region status: ${r.status}"></div>
                </div>
            `).join('');

            // Bind region click events
            container.querySelectorAll('.region-item').forEach(item => {
                item.addEventListener('click', () => {
                    this.showView('regions');
                });
                item.addEventListener('keydown', (e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        item.click();
                    }
                });
            });
        } catch (error) {
            console.error('Failed to update regional health widget:', error);
        }
    }

    async updateLatestResultsWidget() {
        try {
            const config = this.monitoringData.config || await this.api.getConfiguration();
            const container = document.getElementById('latestResultsList');
            if (!container) return;

            const mockResults = [];
            if (config && config.regions) {
                let count = 0;
                for (const [regionName, region] of Object.entries(config.regions)) {
                    if (region.endpoints && count < 5) {
                        for (const endpoint of region.endpoints.slice(0, 2)) {
                            if (count >= 5) break;
                            const latency = Math.floor(Math.random() * 200) + 10;
                            const status = latency < 100 ? 'success' : latency < 150 ? 'warning' : 'error';
                            mockResults.push({
                                endpoint: endpoint.name,
                                latency,
                                status
                            });
                            count++;
                        }
                    }
                }
            }

            if (mockResults.length === 0) {
                container.innerHTML = '<p class="no-data" style="color: var(--color-text-secondary); text-align: center; padding: 1rem;">No recent test results</p>';
                return;
            }

            container.innerHTML = mockResults.map(result => `
                <div class="result-item">
                    <span class="result-endpoint">${result.endpoint}</span>
                    <span class="result-latency">${result.latency}ms</span>
                    <span class="result-status ${result.status}">${result.status.toUpperCase()}</span>
                </div>
            `).join('');
        } catch (error) {
            console.error('Failed to update latest results widget:', error);
        }
    }

    async updatePerformanceMetricsWidget() {
        try {
            // Mock performance metrics
            const metrics = {
                cpu: Math.floor(Math.random() * 60) + 10,
                memory: Math.floor(Math.random() * 70) + 20,
                disk: Math.floor(Math.random() * 50) + 30,
                network: Math.floor(Math.random() * 40) + 10
            };

            this.updateMetric('#cpuUsage', metrics.cpu);
            this.updateMetric('#memoryUsage', metrics.memory);
            this.updateMetric('#diskUsage', metrics.disk);
            this.updateMetric('#networkUsage', metrics.network);
        } catch (error) {
            console.error('Failed to update performance metrics widget:', error);
        }
    }

    updateMetric(selector, value) {
        const el = document.querySelector(selector);
        if (el) {
            el.textContent = `${value}%`;

            // Update color based on value
            el.className = 'status-value metric-value';
            if (value < 60) {
                el.classList.add('status-healthy');
            } else if (value < 80) {
                el.classList.add('status-warning');
            } else {
                el.classList.add('status-danger');
            }
        }
    }

    updateElement(selector, value) {
        const el = document.querySelector(selector);
        if (el) {
            el.textContent = value;
        }
    }

    formatTimeAgo(date) {
        const seconds = Math.floor((new Date() - date) / 1000);

        if (seconds < 60) return 'Just now';
        if (seconds < 3600) return `${Math.floor(seconds / 60)} minutes ago`;
        if (seconds < 86400) return `${Math.floor(seconds / 3600)} hours ago`;
        return `${Math.floor(seconds / 86400)} days ago`;
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
                                <div class="theme-options" role="radiogroup" aria-label="Theme selection">
                                    <button class="btn btn-secondary theme-btn" data-theme="light" role="radio" aria-checked="false">Light</button>
                                    <button class="btn btn-secondary theme-btn" data-theme="dark" role="radio" aria-checked="false">Dark</button>
                                    <button class="btn btn-secondary theme-btn active" data-theme="auto" role="radio" aria-checked="true">Auto</button>
                                </div>
                            </div>
                            <div class="setting-group">
                                <h4>System Information</h4>
                                <div class="system-info" id="systemInfo" role="region" aria-live="polite">
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
                // Refresh all widgets
                await this.updateEndpointSummaryWidget();
                await this.updateRegionalHealthWidget();
                await this.updateLatestResultsWidget();
            });
        }

        // Make endpoint summary widget clickable
        const endpointWidget = document.getElementById('endpointSummaryWidget');
        if (endpointWidget) {
            endpointWidget.style.cursor = 'pointer';
            endpointWidget.addEventListener('click', (e) => {
                if (!e.target.closest('.refresh-btn')) {
                    this.showView('endpoints');
                }
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

                // Update active button and ARIA attributes
                themeButtons.forEach(b => {
                    b.classList.remove('active');
                    b.setAttribute('aria-checked', 'false');
                });
                btn.classList.add('active');
                btn.setAttribute('aria-checked', 'true');
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

            this.showSuccess(`Test completed: ${result.status} (${result.latencyInMs}ms)`);
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
                                <span class="result-value">${result.latencyInMs}ms</span>
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