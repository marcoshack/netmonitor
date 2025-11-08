// Widget Base Class and Implementations for Status Overview

/**
 * Base class for status widgets
 */
class StatusWidget {
    constructor(element, app) {
        this.element = element;
        this.app = app;
        this.updateInterval = 30000; // 30 seconds
        this.intervalId = null;
        this.isUpdating = false;
    }

    async init() {
        await this.updateData();
        this.startAutoUpdate();
        this.bindEvents();
    }

    async updateData() {
        if (this.isUpdating) return;

        this.isUpdating = true;
        try {
            const data = await this.fetchData();
            this.render(data);
            this.updateTimestamp();
        } catch (error) {
            console.error(`Widget update failed:`, error);
            this.showError(error);
        } finally {
            this.isUpdating = false;
        }
    }

    async fetchData() {
        // Override in subclass
        throw new Error('fetchData must be implemented by subclass');
    }

    render(data) {
        // Override in subclass
        throw new Error('render must be implemented by subclass');
    }

    bindEvents() {
        // Override in subclass if needed
    }

    startAutoUpdate() {
        if (this.intervalId) {
            clearInterval(this.intervalId);
        }
        this.intervalId = setInterval(() => this.updateData(), this.updateInterval);
    }

    stopAutoUpdate() {
        if (this.intervalId) {
            clearInterval(this.intervalId);
            this.intervalId = null;
        }
    }

    updateTimestamp() {
        const timestampEl = this.element.querySelector('[data-timestamp]');
        if (timestampEl) {
            timestampEl.textContent = `Updated: ${new Date().toLocaleTimeString()}`;
            timestampEl.setAttribute('data-time', Date.now());
        }
    }

    showError(error) {
        const errorEl = this.element.querySelector('[data-error]');
        if (errorEl) {
            errorEl.textContent = `Error: ${error.message}`;
            errorEl.style.display = 'block';
        }
    }

    hideError() {
        const errorEl = this.element.querySelector('[data-error]');
        if (errorEl) {
            errorEl.style.display = 'none';
        }
    }

    destroy() {
        this.stopAutoUpdate();
    }
}

/**
 * System Status Widget
 */
class SystemStatusWidget extends StatusWidget {
    async fetchData() {
        const systemInfo = await this.app.api.getSystemInfo();
        const monitoringStatus = await this.app.api.getMonitoringStatus();
        return { systemInfo, monitoringStatus };
    }

    render(data) {
        const { systemInfo, monitoringStatus } = data;

        // Update status indicator
        const indicator = this.element.querySelector('#systemStatusIndicator');
        if (indicator) {
            indicator.className = 'status-indicator';
            if (monitoringStatus.running && systemInfo.running) {
                indicator.classList.add('status-healthy');
            } else {
                indicator.classList.add('status-warning');
            }
        }

        // Update status values
        this.updateElement('#applicationStatus', systemInfo.running ? 'Running' : 'Stopped');
        this.updateElement('#monitoringStatus', monitoringStatus.running ? 'Active' : 'Inactive');
        this.updateElement('#appVersion', systemInfo.version);

        // Update last test time
        if (monitoringStatus.lastTestTime) {
            const lastTest = this.formatTimeAgo(new Date(monitoringStatus.lastTestTime));
            const lastTestEl = this.element.querySelector('#lastTestTime');
            if (lastTestEl) {
                lastTestEl.textContent = lastTest;
            }
        }
    }

    updateElement(selector, value) {
        const el = this.element.querySelector(selector);
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
}

/**
 * Endpoint Summary Widget
 */
class EndpointSummaryWidget extends StatusWidget {
    async fetchData() {
        const config = await this.app.api.getConfiguration();
        const status = await this.app.api.getMonitoringStatus();

        // Calculate endpoint statistics
        const stats = this.calculateEndpointStats(config, status);
        return stats;
    }

    calculateEndpointStats(config, status) {
        let total = 0;
        let healthy = 0;
        let warning = 0;
        let down = 0;

        if (config && config.regions) {
            Object.values(config.regions).forEach(region => {
                if (region.endpoints) {
                    total += region.endpoints.length;
                    // In a real scenario, we would check actual endpoint health
                    // For now, we'll use placeholder logic
                    healthy += Math.floor(region.endpoints.length * 0.8);
                    warning += Math.floor(region.endpoints.length * 0.15);
                    down += region.endpoints.length - Math.floor(region.endpoints.length * 0.95);
                }
            });
        }

        return { total, healthy, warning, down };
    }

    render(data) {
        this.updateElement('#totalEndpoints', data.total);
        this.updateElement('#healthyEndpoints', data.healthy);
        this.updateElement('#warningEndpoints', data.warning);
        this.updateElement('#downEndpoints', data.down);
    }

    updateElement(selector, value) {
        const el = this.element.querySelector(selector);
        if (el) {
            el.textContent = value;
        }
    }

    bindEvents() {
        const refreshBtn = this.element.querySelector('#refreshEndpoints');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.updateData();
            });
        }

        // Make widget clickable to navigate to endpoints view
        this.element.style.cursor = 'pointer';
        this.element.addEventListener('click', (e) => {
            if (e.target.id !== 'refreshEndpoints') {
                this.app.showView('endpoints');
            }
        });
    }
}

/**
 * Regional Health Widget
 */
class RegionalHealthWidget extends StatusWidget {
    async fetchData() {
        const config = await this.app.api.getConfiguration();
        return config.regions || {};
    }

    render(data) {
        const container = this.element.querySelector('#regionalHealthList');
        if (!container) return;

        const regions = Object.entries(data).map(([name, region]) => {
            // Calculate mock health status
            const avgLatency = Math.floor(Math.random() * 200) + 20;
            const uptime = (99 + Math.random()).toFixed(1);
            const status = avgLatency < 100 ? 'healthy' : avgLatency < 150 ? 'warning' : 'critical';

            return { name, avgLatency, uptime, status, region };
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

        this.bindRegionEvents();
    }

    bindRegionEvents() {
        const regionItems = this.element.querySelectorAll('.region-item');
        regionItems.forEach(item => {
            item.addEventListener('click', () => {
                const regionName = item.dataset.region;
                this.app.showView('regions');
                // Could scroll to specific region or filter
            });

            item.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    item.click();
                }
            });
        });
    }
}

/**
 * Performance Metrics Widget
 */
class PerformanceMetricsWidget extends StatusWidget {
    async fetchData() {
        // In a real implementation, this would fetch actual system metrics
        return {
            cpu: Math.floor(Math.random() * 60) + 10,
            memory: Math.floor(Math.random() * 70) + 20,
            disk: Math.floor(Math.random() * 50) + 30,
            network: Math.floor(Math.random() * 40) + 10
        };
    }

    render(data) {
        this.updateMetric('#cpuUsage', data.cpu);
        this.updateMetric('#memoryUsage', data.memory);
        this.updateMetric('#diskUsage', data.disk);
        this.updateMetric('#networkUsage', data.network);
    }

    updateMetric(selector, value) {
        const el = this.element.querySelector(selector);
        if (el) {
            el.textContent = `${value}%`;

            // Update color based on value
            el.className = 'metric-value';
            if (value < 60) {
                el.classList.add('status-healthy');
            } else if (value < 80) {
                el.classList.add('status-warning');
            } else {
                el.classList.add('status-danger');
            }
        }
    }
}

/**
 * Latest Results Widget
 */
class LatestResultsWidget extends StatusWidget {
    async fetchData() {
        // In a real implementation, this would fetch latest test results
        const config = await this.app.api.getConfiguration();
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
                            region: regionName,
                            latency,
                            status,
                            timestamp: new Date(Date.now() - Math.random() * 300000)
                        });
                        count++;
                    }
                }
            }
        }

        return mockResults;
    }

    render(data) {
        const container = this.element.querySelector('#latestResultsList');
        if (!container) return;

        if (data.length === 0) {
            container.innerHTML = '<p class="no-data">No recent test results</p>';
            return;
        }

        container.innerHTML = data.map(result => `
            <div class="result-item">
                <span class="result-endpoint">${result.endpoint}</span>
                <span class="result-latency">${result.latency}ms</span>
                <span class="result-status ${result.status}">${result.status.toUpperCase()}</span>
            </div>
        `).join('');
    }
}

// Export widgets
export {
    StatusWidget,
    SystemStatusWidget,
    EndpointSummaryWidget,
    RegionalHealthWidget,
    PerformanceMetricsWidget,
    LatestResultsWidget
};
