# T029: Endpoint Status Grid

## Overview
Create a comprehensive endpoint status grid that displays real-time status, latency, uptime, and detailed information for all monitored endpoints with sorting, filtering, and management capabilities.

## Context
Users need a detailed view of all monitored endpoints with their current status, recent performance metrics, and quick access to endpoint management functions. The grid should provide both overview and drill-down capabilities.

## Task Description
Implement a dynamic, sortable, and filterable grid displaying endpoint status information with real-time updates, interactive features, and management integration.

## Acceptance Criteria
- [ ] Grid displaying all endpoints with status indicators
- [ ] Real-time status updates without page refresh
- [ ] Sortable columns (name, status, latency, uptime)
- [ ] Filterable by region, status, and endpoint type
- [ ] Search functionality for endpoint names
- [ ] Individual endpoint detail views
- [ ] Quick action buttons (test now, edit, delete)
- [ ] Responsive grid layout for mobile devices
- [ ] Pagination for large numbers of endpoints
- [ ] Export grid data functionality

## Grid Structure
```html
<div class="endpoint-grid-container">
  <div class="grid-controls">
    <div class="search-filter">
      <input type="text" class="search-input" placeholder="Search endpoints...">
      <select class="region-filter">
        <option value="">All Regions</option>
        <option value="NA-East">NA-East</option>
        <option value="EU-West">EU-West</option>
      </select>
      <select class="status-filter">
        <option value="">All Status</option>
        <option value="healthy">Healthy</option>
        <option value="warning">Warning</option>
        <option value="down">Down</option>
      </select>
    </div>
    <div class="grid-actions">
      <button class="add-endpoint-btn">+ Add Endpoint</button>
      <button class="refresh-btn">↻ Refresh</button>
      <button class="export-btn">📊 Export</button>
    </div>
  </div>

  <div class="grid-wrapper">
    <table class="endpoint-grid" role="grid">
      <thead>
        <tr role="row">
          <th class="sortable" data-sort="name">
            <span>Name</span>
            <span class="sort-indicator">↕</span>
          </th>
          <th class="sortable" data-sort="status">
            <span>Status</span>
            <span class="sort-indicator">↕</span>
          </th>
          <th class="sortable" data-sort="region">
            <span>Region</span>
            <span class="sort-indicator">↕</span>
          </th>
          <th class="sortable" data-sort="latency">
            <span>Latency</span>
            <span class="sort-indicator">↕</span>
          </th>
          <th class="sortable" data-sort="uptime">
            <span>Uptime</span>
            <span class="sort-indicator">↕</span>
          </th>
          <th class="sortable" data-sort="lastTest">
            <span>Last Test</span>
            <span class="sort-indicator">↕</span>
          </th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody id="endpoint-grid-body">
        <!-- Dynamic content -->
      </tbody>
    </table>
  </div>

  <div class="grid-pagination">
    <div class="pagination-info">
      Showing 1-20 of 45 endpoints
    </div>
    <div class="pagination-controls">
      <button class="page-btn" data-page="prev">‹ Previous</button>
      <button class="page-btn active" data-page="1">1</button>
      <button class="page-btn" data-page="2">2</button>
      <button class="page-btn" data-page="3">3</button>
      <button class="page-btn" data-page="next">Next ›</button>
    </div>
  </div>
</div>
```

## Grid Row Component
```html
<tr class="endpoint-row" data-endpoint-id="na-east-google-dns" data-status="healthy">
  <td class="endpoint-name">
    <div class="name-cell">
      <div class="endpoint-icon">
        <span class="protocol-badge icmp">ICMP</span>
      </div>
      <div class="endpoint-info">
        <div class="primary-name">Google DNS</div>
        <div class="endpoint-address">8.8.8.8</div>
      </div>
    </div>
  </td>
  <td class="endpoint-status">
    <div class="status-indicator healthy" title="Healthy">
      <span class="status-dot">●</span>
      <span class="status-text">Healthy</span>
    </div>
  </td>
  <td class="endpoint-region">
    <span class="region-badge">NA-East</span>
  </td>
  <td class="endpoint-latency">
    <div class="latency-cell">
      <span class="current-latency">23ms</span>
      <div class="latency-trend">
        <span class="trend-indicator up">↗</span>
        <span class="trend-value">+2ms</span>
      </div>
    </div>
  </td>
  <td class="endpoint-uptime">
    <div class="uptime-cell">
      <span class="uptime-value">99.8%</span>
      <div class="uptime-bar">
        <div class="uptime-fill" style="width: 99.8%"></div>
      </div>
    </div>
  </td>
  <td class="endpoint-last-test">
    <span class="timestamp" data-timestamp="2025-09-27T19:30:00Z">
      2 min ago
    </span>
  </td>
  <td class="endpoint-actions">
    <div class="action-buttons">
      <button class="action-btn test-btn" title="Test Now">▶</button>
      <button class="action-btn edit-btn" title="Edit">✎</button>
      <button class="action-btn delete-btn" title="Delete">🗑</button>
      <button class="action-btn more-btn" title="More">⋯</button>
    </div>
  </td>
</tr>
```

## Grid Management Class
```javascript
class EndpointGrid {
  constructor(container, dataSource) {
    this.container = container;
    this.dataSource = dataSource;
    this.currentPage = 1;
    this.pageSize = 20;
    this.sortColumn = 'name';
    this.sortDirection = 'asc';
    this.filters = {
      search: '',
      region: '',
      status: '',
      type: ''
    };

    this.init();
  }

  async init() {
    this.bindEvents();
    await this.loadData();
    this.startAutoUpdate();
  }

  async loadData() {
    try {
      const response = await this.dataSource.getEndpoints({
        page: this.currentPage,
        pageSize: this.pageSize,
        sort: this.sortColumn,
        direction: this.sortDirection,
        filters: this.filters
      });

      this.renderGrid(response.endpoints);
      this.updatePagination(response.pagination);
    } catch (error) {
      this.showError('Failed to load endpoint data');
    }
  }

  renderGrid(endpoints) {
    const tbody = this.container.querySelector('#endpoint-grid-body');
    tbody.innerHTML = endpoints.map(endpoint => this.renderRow(endpoint)).join('');
  }

  renderRow(endpoint) {
    return `
      <tr class="endpoint-row" data-endpoint-id="${endpoint.id}" data-status="${endpoint.status}">
        <td class="endpoint-name">
          <div class="name-cell">
            <div class="endpoint-icon">
              <span class="protocol-badge ${endpoint.type.toLowerCase()}">${endpoint.type}</span>
            </div>
            <div class="endpoint-info">
              <div class="primary-name">${endpoint.name}</div>
              <div class="endpoint-address">${endpoint.address}</div>
            </div>
          </div>
        </td>
        <td class="endpoint-status">
          <div class="status-indicator ${endpoint.status}" title="${endpoint.status}">
            <span class="status-dot">●</span>
            <span class="status-text">${endpoint.status}</span>
          </div>
        </td>
        <td class="endpoint-region">
          <span class="region-badge">${endpoint.region}</span>
        </td>
        <td class="endpoint-latency">
          <div class="latency-cell">
            <span class="current-latency">${endpoint.latency}ms</span>
            ${this.renderLatencyTrend(endpoint.latencyTrend)}
          </div>
        </td>
        <td class="endpoint-uptime">
          <div class="uptime-cell">
            <span class="uptime-value">${endpoint.uptime}%</span>
            <div class="uptime-bar">
              <div class="uptime-fill" style="width: ${endpoint.uptime}%"></div>
            </div>
          </div>
        </td>
        <td class="endpoint-last-test">
          <span class="timestamp" data-timestamp="${endpoint.lastTest}">
            ${this.formatRelativeTime(endpoint.lastTest)}
          </span>
        </td>
        <td class="endpoint-actions">
          ${this.renderActionButtons(endpoint)}
        </td>
      </tr>
    `;
  }

  bindEvents() {
    // Sorting
    this.container.addEventListener('click', (e) => {
      if (e.target.closest('.sortable')) {
        this.handleSort(e.target.closest('.sortable').dataset.sort);
      }
    });

    // Filtering
    const searchInput = this.container.querySelector('.search-input');
    searchInput.addEventListener('input', (e) => {
      this.filters.search = e.target.value;
      this.debounceFilter();
    });

    // Actions
    this.container.addEventListener('click', (e) => {
      if (e.target.closest('.test-btn')) {
        this.handleTestEndpoint(e.target.closest('.endpoint-row').dataset.endpointId);
      }
      if (e.target.closest('.edit-btn')) {
        this.handleEditEndpoint(e.target.closest('.endpoint-row').dataset.endpointId);
      }
      if (e.target.closest('.delete-btn')) {
        this.handleDeleteEndpoint(e.target.closest('.endpoint-row').dataset.endpointId);
      }
    });
  }

  startAutoUpdate() {
    setInterval(() => {
      this.updateGridData();
    }, 30000); // Update every 30 seconds
  }

  async updateGridData() {
    // Update only the data that might have changed (status, latency, etc.)
    try {
      const updates = await this.dataSource.getEndpointUpdates();
      this.applyUpdates(updates);
    } catch (error) {
      console.error('Failed to update grid data:', error);
    }
  }
}
```

## Features

### Status Indicators
- **Healthy**: Green dot, normal latency within thresholds
- **Warning**: Yellow dot, latency above warning threshold
- **Down**: Red dot, endpoint not responding
- **Unknown**: Gray dot, no recent test data

### Interactive Features
- Click row to view detailed endpoint information
- Sort by any column (name, status, latency, uptime)
- Filter by region, status, protocol type
- Search by endpoint name or address
- Quick actions: test now, edit, delete

### Real-time Updates
- Automatically update status indicators
- Refresh latency and uptime values
- Maintain user's current sort/filter state
- Smooth animations for status changes

## Verification Steps
1. Display endpoint grid - should show all configured endpoints
2. Test sorting - should sort by clicked column
3. Test filtering - should filter by region, status, and search term
4. Verify real-time updates - should update status and metrics automatically
5. Test quick actions - should trigger test, edit, and delete operations
6. Verify responsive behavior - should adapt to mobile screens
7. Test pagination - should handle large numbers of endpoints
8. Verify accessibility - should support keyboard navigation and screen readers

## Dependencies
- T026: Dashboard Layout and Structure
- T015: Monitoring Status API
- T014: Endpoint Management
- T012: Manual Test Execution

## Notes
- Implement virtual scrolling for very large endpoint lists
- Use efficient DOM updates to maintain performance
- Consider implementing bulk operations for multiple endpoints
- Plan for endpoint grouping and categorization features
- Ensure consistent behavior across different browsers
- Implement proper loading states for all operations