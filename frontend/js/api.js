// API Client for Backend Communication

class APIClient {
    constructor() {
        this.isConnected = false;
        this.cache = new Map();
        this.cacheTTL = 30000; // 30 seconds
    }

    async init() {
        try {
            // Test connection to backend
            await this.testConnection();
            this.isConnected = true;
            console.log('API client connected to backend');
        } catch (error) {
            console.error('Failed to connect to backend:', error);
            this.isConnected = false;
            throw error;
        }
    }

    async testConnection() {
        // Test basic connectivity with the backend
        try {
            const result = await window.go.main.App.GetSystemInfo();
            return result;
        } catch (error) {
            throw new Error('Backend not responding: ' + error.message);
        }
    }

    // Generic request wrapper with caching and error handling
    async makeRequest(method, args = [], options = {}) {
        const { 
            cache = true, 
            cacheTTL = this.cacheTTL,
            retry = true 
        } = options;

        const cacheKey = `${method}_${JSON.stringify(args)}`;

        // Check cache first
        if (cache && this.cache.has(cacheKey)) {
            const cached = this.cache.get(cacheKey);
            if (Date.now() - cached.timestamp < cacheTTL) {
                return cached.data;
            }
        }

        try {
            // Get the method from the Wails go bindings
            const methodPath = method.split('.');
            let target = window.go.main.App;
            
            for (let i = 0; i < methodPath.length; i++) {
                target = target[methodPath[i]];
                if (!target) {
                    throw new Error(`Method ${method} not found`);
                }
            }

            // Call the method
            const result = await target(...args);

            // Cache successful results
            if (cache) {
                this.cache.set(cacheKey, {
                    data: result,
                    timestamp: Date.now()
                });
            }

            return result;

        } catch (error) {
            console.error(`API call failed: ${method}`, error);
            
            if (retry && this.shouldRetry(error)) {
                console.log(`Retrying API call: ${method}`);
                await this.delay(1000);
                return this.makeRequest(method, args, { ...options, retry: false });
            }

            throw error;
        }
    }

    shouldRetry(error) {
        // Retry on network errors, timeouts, etc.
        const retryableErrors = [
            'network error',
            'timeout',
            'connection refused',
            'backend not responding'
        ];

        const errorMessage = error.message.toLowerCase();
        return retryableErrors.some(pattern => errorMessage.includes(pattern));
    }

    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    // Clear cache
    clearCache() {
        this.cache.clear();
    }

    // System Info API
    async getSystemInfo() {
        return this.makeRequest('GetSystemInfo');
    }

    // Configuration API  
    async getConfiguration() {
        return this.makeRequest('GetConfiguration', [], { cache: true, cacheTTL: 60000 });
    }

    async updateConfiguration(config) {
        const result = await this.makeRequest('UpdateConfiguration', [config], { cache: false });
        this.clearCache(); // Clear cache after configuration update
        return result;
    }

    // Theme API
    async setTheme(theme) {
        return this.makeRequest('SetTheme', [theme], { cache: false });
    }

    // Monitoring API
    async getMonitoringStatus() {
        return this.makeRequest('GetMonitoringStatus', [], { cache: true, cacheTTL: 5000 });
    }

    async startMonitoring() {
        return this.makeRequest('StartMonitoring', [], { cache: false });
    }

    async stopMonitoring() {
        return this.makeRequest('StopMonitoring', [], { cache: false });
    }

    // Test API
    async runManualTest(endpointId) {
        return this.makeRequest('RunManualTest', [endpointId], { cache: false });
    }

    // Connection status
    isConnected() {
        return this.isConnected;
    }

    // Event handling for real-time updates
    onConnectionChange(callback) {
        // TODO: Implement WebSocket or polling for real-time updates
        // For now, this is a placeholder
        this._connectionCallback = callback;
    }

    onDataUpdate(callback) {
        // TODO: Implement real-time data update notifications
        this._dataCallback = callback;
    }

    // Utility methods
    formatError(error) {
        if (typeof error === 'string') {
            return error;
        } else if (error.message) {
            return error.message;
        } else {
            return 'Unknown error occurred';
        }
    }
}

export { APIClient };