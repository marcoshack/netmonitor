# T035: Frontend API Integration

## Overview
Complete the integration between frontend components and backend APIs, implementing WebSocket connections for real-time updates, error handling, and state management.

## Context
The frontend components need seamless integration with the Go backend through Wails context. This includes real-time data updates, proper error handling, loading states, and consistent state management across all components.

## Task Description
Implement comprehensive frontend-backend integration with WebSocket support, centralized state management, error handling, and real-time update coordination across all dashboard components.

## Acceptance Criteria
- [ ] WebSocket integration for real-time data updates
- [ ] Centralized API client with error handling
- [ ] State management system for frontend data
- [ ] Loading states and skeleton screens
- [ ] Error boundaries and graceful error handling
- [ ] Real-time update coordination across components
- [ ] Optimistic updates for better user experience
- [ ] Connection status monitoring and reconnection
- [ ] Performance monitoring and optimization

## API Client Architecture
```javascript
class NetMonitorAPI {
  constructor() {
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.eventListeners = new Map();
    this.requestQueue = [];
    this.pendingRequests = new Map();

    this.init();
  }

  async init() {
    try {
      // Initialize Wails context
      await this.initializeWailsContext();

      // Setup WebSocket for real-time updates
      await this.setupWebSocket();

      // Start connection monitoring
      this.startConnectionMonitoring();

      this.isConnected = true;
      this.emit('connected');
    } catch (error) {
      console.error('Failed to initialize API client:', error);
      this.handleConnectionError(error);
    }
  }

  async initializeWailsContext() {
    if (!window.go) {
      throw new Error('Wails context not available');
    }

    // Test connection with a simple call
    try {
      await window.go.App.GetSystemInfo();
    } catch (error) {
      throw new Error('Backend not responding');
    }
  }

  async setupWebSocket() {
    // Note: WebSocket simulation using polling for Wails
    // In a real WebSocket implementation, this would establish WS connection
    this.startRealTimeUpdates();
  }

  startRealTimeUpdates() {
    this.updateInterval = setInterval(async () => {
      if (!this.isConnected) return;

      try {
        // Get latest status updates
        const updates = await this.getStatusUpdates();
        if (updates) {
          this.emit('statusUpdate', updates);
        }
      } catch (error) {
        console.warn('Failed to get status updates:', error);
      }
    }, 5000); // Update every 5 seconds
  }

  // API Methods with error handling and caching
  async getConfiguration() {
    return this.makeRequest('getConfiguration', () =>
      window.go.App.GetConfiguration()
    );
  }

  async getMonitoringStatus() {
    return this.makeRequest('getMonitoringStatus', () =>
      window.go.App.GetMonitoringStatus()
    );
  }

  async getEndpointStatus(endpointId) {
    return this.makeRequest(`getEndpointStatus-${endpointId}`, () =>
      window.go.App.GetEndpointStatus(endpointId)
    );
  }

  async getRegionStatus(regionName) {
    return this.makeRequest(`getRegionStatus-${regionName}`, () =>
      window.go.App.GetRegionStatus(regionName)
    );
  }

  async queryTimeSeries(request) {
    return this.makeRequest(`queryTimeSeries-${JSON.stringify(request)}`, () =>
      window.go.App.QueryTimeSeries(request)
    );
  }

  async runManualTest(endpointId) {
    return this.makeRequest(`runManualTest-${endpointId}`, () =>
      window.go.App.RunManualTest(endpointId)
    );
  }

  async addEndpoint(regionName, endpoint) {
    const result = await this.makeRequest('addEndpoint', () =>
      window.go.App.AddEndpoint(regionName, endpoint)
    );

    // Emit event for optimistic updates
    this.emit('endpointAdded', { regionName, endpoint });
    return result;
  }

  async updateEndpoint(endpointId, endpoint) {
    const result = await this.makeRequest('updateEndpoint', () =>
      window.go.App.UpdateEndpoint(endpointId, endpoint)
    );

    this.emit('endpointUpdated', { endpointId, endpoint });
    return result;
  }

  // Generic request handler with caching and error handling
  async makeRequest(cacheKey, requestFn, options = {}) {
    const {
      cache = true,
      cacheTTL = 30000, // 30 seconds
      retry = true
    } = options;

    // Check cache first
    if (cache && this.cache.has(cacheKey)) {
      const cached = this.cache.get(cacheKey);
      if (Date.now() - cached.timestamp < cacheTTL) {
        return cached.data;
      }
    }

    try {
      const result = await requestFn();

      // Cache successful results
      if (cache) {
        this.cache.set(cacheKey, {
          data: result,
          timestamp: Date.now()
        });
      }

      return result;
    } catch (error) {
      if (retry && this.shouldRetry(error)) {
        await this.delay(1000);
        return this.makeRequest(cacheKey, requestFn, { ...options, retry: false });
      }

      throw this.handleAPIError(error);
    }
  }

  shouldRetry(error) {
    // Retry on network errors, timeouts, etc.
    return error.message.includes('network') ||
           error.message.includes('timeout') ||
           error.code === 'CONNECTION_ERROR';
  }

  handleAPIError(error) {
    const apiError = {
      message: error.message || 'Unknown error occurred',
      code: error.code || 'UNKNOWN_ERROR',
      timestamp: new Date().toISOString(),
      context: error.context || {}
    };

    this.emit('error', apiError);
    return apiError;
  }

  // Event system for component communication
  on(event, callback) {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, []);
    }
    this.eventListeners.get(event).push(callback);
  }

  off(event, callback) {
    if (this.eventListeners.has(event)) {
      const listeners = this.eventListeners.get(event);
      const index = listeners.indexOf(callback);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  }

  emit(event, data) {
    if (this.eventListeners.has(event)) {
      this.eventListeners.get(event).forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in event listener for ${event}:`, error);
        }
      });
    }
  }

  // Connection monitoring
  startConnectionMonitoring() {
    setInterval(async () => {
      try {
        await window.go.App.GetSystemInfo();
        if (!this.isConnected) {
          this.isConnected = true;
          this.reconnectAttempts = 0;
          this.emit('connected');
        }
      } catch (error) {
        if (this.isConnected) {
          this.isConnected = false;
          this.emit('disconnected');
          this.attemptReconnection();
        }
      }
    }, 10000); // Check every 10 seconds
  }

  async attemptReconnection() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.emit('reconnectionFailed');
      return;
    }

    this.reconnectAttempts++;
    this.emit('reconnecting', { attempt: this.reconnectAttempts });

    await this.delay(this.reconnectDelay * this.reconnectAttempts);

    try {
      await this.init();
    } catch (error) {
      this.attemptReconnection();
    }
  }

  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // Cleanup
  destroy() {
    if (this.updateInterval) {
      clearInterval(this.updateInterval);
    }
    this.eventListeners.clear();
    this.cache.clear();
  }
}
```

## State Management System
```javascript
class StateManager {
  constructor() {
    this.state = {
      configuration: null,
      monitoringStatus: null,
      endpoints: new Map(),
      regions: new Map(),
      timeSeriesData: new Map(),
      ui: {
        selectedRegions: ['all'],
        timeRange: { key: '24h', startTime: null, endTime: null },
        theme: 'auto',
        loading: new Set(),
        errors: []
      }
    };

    this.subscribers = new Map();
    this.middleware = [];
  }

  // Subscribe to state changes
  subscribe(path, callback) {
    if (!this.subscribers.has(path)) {
      this.subscribers.set(path, []);
    }
    this.subscribers.get(path).push(callback);

    // Return unsubscribe function
    return () => {
      const callbacks = this.subscribers.get(path);
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    };
  }

  // Update state with change notification
  setState(path, value) {
    const oldValue = this.getState(path);
    this.setNestedProperty(this.state, path, value);

    // Notify subscribers
    this.notifySubscribers(path, value, oldValue);
  }

  // Get state value
  getState(path) {
    return this.getNestedProperty(this.state, path);
  }

  // Set loading state
  setLoading(key, isLoading) {
    if (isLoading) {
      this.state.ui.loading.add(key);
    } else {
      this.state.ui.loading.delete(key);
    }
    this.notifySubscribers('ui.loading', this.state.ui.loading);
  }

  // Add error
  addError(error) {
    this.state.ui.errors.push({
      ...error,
      id: Date.now(),
      timestamp: new Date().toISOString()
    });
    this.notifySubscribers('ui.errors', this.state.ui.errors);
  }

  // Remove error
  removeError(errorId) {
    this.state.ui.errors = this.state.ui.errors.filter(e => e.id !== errorId);
    this.notifySubscribers('ui.errors', this.state.ui.errors);
  }

  // Helper methods
  getNestedProperty(obj, path) {
    return path.split('.').reduce((current, key) => current?.[key], obj);
  }

  setNestedProperty(obj, path, value) {
    const keys = path.split('.');
    const lastKey = keys.pop();
    const target = keys.reduce((current, key) => {
      if (!current[key]) current[key] = {};
      return current[key];
    }, obj);
    target[lastKey] = value;
  }

  notifySubscribers(path, newValue, oldValue) {
    if (this.subscribers.has(path)) {
      this.subscribers.get(path).forEach(callback => {
        try {
          callback(newValue, oldValue);
        } catch (error) {
          console.error(`Error in state subscriber for ${path}:`, error);
        }
      });
    }
  }
}
```

## Component Integration System
```javascript
class ComponentManager {
  constructor(api, stateManager) {
    this.api = api;
    this.state = stateManager;
    this.components = new Map();
    this.updateQueue = [];

    this.init();
  }

  init() {
    // Setup API event listeners
    this.api.on('statusUpdate', (data) => this.handleStatusUpdate(data));
    this.api.on('connected', () => this.handleConnectionChange(true));
    this.api.on('disconnected', () => this.handleConnectionChange(false));
    this.api.on('error', (error) => this.handleAPIError(error));

    // Setup state change handlers
    this.state.subscribe('ui.selectedRegions', (regions) => {
      this.updateComponentsForRegionChange(regions);
    });

    this.state.subscribe('ui.timeRange', (timeRange) => {
      this.updateComponentsForTimeRangeChange(timeRange);
    });
  }

  // Register component for management
  registerComponent(id, component) {
    this.components.set(id, component);

    // Setup component error handling
    if (component.onError) {
      component.onError = (error) => {
        this.state.addError({
          source: id,
          message: error.message,
          code: error.code
        });
      };
    }
  }

  // Coordinate updates across components
  async refreshAllData() {
    this.state.setLoading('refreshAll', true);

    try {
      // Load all data in parallel
      const [config, status, regions] = await Promise.all([
        this.api.getConfiguration(),
        this.api.getMonitoringStatus(),
        this.api.getRegionStatus()
      ]);

      // Update state
      this.state.setState('configuration', config);
      this.state.setState('monitoringStatus', status);
      this.state.setState('regions', regions);

      // Notify all components
      this.components.forEach(component => {
        if (component.onDataRefresh) {
          component.onDataRefresh();
        }
      });

    } catch (error) {
      this.state.addError({
        source: 'global',
        message: 'Failed to refresh data',
        details: error.message
      });
    } finally {
      this.state.setLoading('refreshAll', false);
    }
  }

  handleStatusUpdate(data) {
    // Update relevant state
    if (data.endpoints) {
      Object.entries(data.endpoints).forEach(([id, endpoint]) => {
        this.state.setState(`endpoints.${id}`, endpoint);
      });
    }

    if (data.regions) {
      Object.entries(data.regions).forEach(([name, region]) => {
        this.state.setState(`regions.${name}`, region);
      });
    }

    // Batch component updates
    this.queueComponentUpdates(['statusWidgets', 'endpointGrid', 'graphs']);
  }

  queueComponentUpdates(componentIds) {
    // Add to update queue and debounce
    this.updateQueue.push(...componentIds);

    if (this.updateTimeout) {
      clearTimeout(this.updateTimeout);
    }

    this.updateTimeout = setTimeout(() => {
      this.processUpdateQueue();
    }, 100); // 100ms debounce
  }

  processUpdateQueue() {
    const uniqueComponents = [...new Set(this.updateQueue)];
    this.updateQueue = [];

    uniqueComponents.forEach(id => {
      const component = this.components.get(id);
      if (component && component.onDataUpdate) {
        component.onDataUpdate();
      }
    });
  }

  handleConnectionChange(isConnected) {
    this.state.setState('ui.connected', isConnected);

    this.components.forEach(component => {
      if (component.onConnectionChange) {
        component.onConnectionChange(isConnected);
      }
    });

    if (isConnected) {
      // Refresh data when reconnected
      this.refreshAllData();
    }
  }

  handleAPIError(error) {
    this.state.addError(error);

    // Notify error-aware components
    this.components.forEach(component => {
      if (component.onAPIError) {
        component.onAPIError(error);
      }
    });
  }
}
```

## Application Initialization
```javascript
// Initialize the complete frontend system
class NetMonitorApp {
  constructor() {
    this.api = new NetMonitorAPI();
    this.state = new StateManager();
    this.componentManager = new ComponentManager(this.api, this.state);

    this.components = {};
  }

  async init() {
    try {
      // Initialize API connection
      await this.api.init();

      // Initialize all components
      await this.initializeComponents();

      // Setup global error handling
      this.setupErrorHandling();

      // Setup keyboard shortcuts
      this.setupKeyboardShortcuts();

      // Initial data load
      await this.componentManager.refreshAllData();

      console.log('NetMonitor app initialized successfully');
    } catch (error) {
      console.error('Failed to initialize NetMonitor app:', error);
      this.showFatalError(error);
    }
  }

  async initializeComponents() {
    // Initialize all dashboard components
    const componentConfigs = [
      { id: 'statusWidgets', class: StatusWidgets, selector: '.status-widgets' },
      { id: 'latencyGraph', class: LatencyGraph, selector: '#latency-chart' },
      { id: 'endpointGrid', class: EndpointGrid, selector: '.endpoint-grid' },
      { id: 'regionSelector', class: RegionSelector, selector: '.region-selector' },
      { id: 'timeRangeSelector', class: TimeRangeSelector, selector: '.time-range-selector' },
      { id: 'manualTestInterface', class: ManualTestInterface, selector: '.manual-test-container' }
    ];

    for (const config of componentConfigs) {
      const element = document.querySelector(config.selector);
      if (element) {
        this.components[config.id] = new config.class(element, {
          api: this.api,
          state: this.state
        });

        this.componentManager.registerComponent(config.id, this.components[config.id]);
      }
    }
  }

  setupErrorHandling() {
    // Global error handler
    window.addEventListener('error', (event) => {
      this.state.addError({
        source: 'global',
        message: event.error.message,
        stack: event.error.stack
      });
    });

    // Unhandled promise rejection handler
    window.addEventListener('unhandledrejection', (event) => {
      this.state.addError({
        source: 'promise',
        message: event.reason.message || 'Unhandled promise rejection',
        details: event.reason
      });
    });
  }

  setupKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
          case 'r':
            e.preventDefault();
            this.componentManager.refreshAllData();
            break;
          case 'k':
            e.preventDefault();
            this.showKeyboardShortcuts();
            break;
        }
      }
    });
  }

  showFatalError(error) {
    document.body.innerHTML = `
      <div class="fatal-error">
        <h1>Failed to start NetMonitor</h1>
        <p>The application failed to initialize properly.</p>
        <details>
          <summary>Error details</summary>
          <pre>${error.message}\n${error.stack}</pre>
        </details>
        <button onclick="location.reload()">Retry</button>
      </div>
    `;
  }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  window.netMonitorApp = new NetMonitorApp();
  window.netMonitorApp.init();
});
```

## Verification Steps
1. Test real-time updates - should update components automatically
2. Verify error handling - should display and recover from errors gracefully
3. Test connection monitoring - should detect and handle disconnections
4. Verify state management - should maintain consistent state across components
5. Test API integration - should handle all backend calls correctly
6. Verify loading states - should show appropriate loading indicators
7. Test optimistic updates - should provide immediate feedback for user actions
8. Verify component coordination - should update related components when data changes

## Dependencies
- T026: Dashboard Layout and Structure
- T027: Status Overview Widgets
- T028: Interactive Latency Graphs
- T029: Endpoint Status Grid
- T030: Region Selector Component
- T031: Manual Test Interface
- T032: Time Range Selector Component
- T015: Monitoring Status API
- T005: Wails Frontend-Backend Integration

## Notes
- Implement proper error boundaries for React-like error handling
- Consider implementing service worker for offline functionality
- Plan for future WebSocket integration when Wails supports it
- Optimize for performance with large datasets
- Implement proper cleanup on component destruction
- Consider implementing undo/redo functionality for user actions