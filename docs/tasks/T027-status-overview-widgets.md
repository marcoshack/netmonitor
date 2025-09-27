# T027: Status Overview Widgets

## Overview
Create status overview widgets that provide at-a-glance information about network monitoring health, including real-time indicators, summary statistics, and alert status.

## Context
The dashboard needs prominent status widgets that immediately communicate the overall health of monitored networks. Users should be able to quickly assess the monitoring status without diving into detailed graphs.

## Task Description
Implement a set of status overview widgets that display real-time monitoring status, summary statistics, regional health indicators, and system status information.

## Acceptance Criteria
- [ ] Overall system status indicator
- [ ] Regional health summary widgets
- [ ] Real-time endpoint status counters
- [ ] Latest test results summary
- [ ] Alert and warning indicators
- [ ] Performance metrics display
- [ ] Auto-updating real-time data
- [ ] Interactive widgets with drill-down capability
- [ ] Visual status indicators (colors, icons)

## Widget Components

### System Status Widget
```html
<div class="widget system-status">
  <div class="widget-header">
    <h3 class="widget-title">System Status</h3>
    <div class="status-indicator" data-status="healthy">●</div>
  </div>
  <div class="widget-content">
    <div class="status-grid">
      <div class="status-item">
        <span class="status-label">Monitoring</span>
        <span class="status-value running">Running</span>
      </div>
      <div class="status-item">
        <span class="status-label">Last Test</span>
        <span class="status-value">2 minutes ago</span>
      </div>
      <div class="status-item">
        <span class="status-label">Next Test</span>
        <span class="status-value">3 minutes</span>
      </div>
    </div>
  </div>
</div>
```

### Endpoint Summary Widget
```html
<div class="widget endpoint-summary">
  <div class="widget-header">
    <h3 class="widget-title">Endpoints</h3>
    <button class="refresh-btn" aria-label="Refresh">↻</button>
  </div>
  <div class="widget-content">
    <div class="summary-stats">
      <div class="stat-item stat-healthy">
        <div class="stat-number">12</div>
        <div class="stat-label">Healthy</div>
      </div>
      <div class="stat-item stat-warning">
        <div class="stat-number">2</div>
        <div class="stat-label">Warning</div>
      </div>
      <div class="stat-item stat-down">
        <div class="stat-number">0</div>
        <div class="stat-label">Down</div>
      </div>
    </div>
  </div>
</div>
```

### Regional Health Widget
```html
<div class="widget regional-health">
  <div class="widget-header">
    <h3 class="widget-title">Regional Health</h3>
  </div>
  <div class="widget-content">
    <div class="region-list">
      <div class="region-item" data-status="healthy">
        <div class="region-name">NA-East</div>
        <div class="region-stats">
          <span class="avg-latency">23ms</span>
          <span class="uptime">99.8%</span>
        </div>
        <div class="region-indicator healthy">●</div>
      </div>
      <div class="region-item" data-status="warning">
        <div class="region-name">EU-West</div>
        <div class="region-stats">
          <span class="avg-latency">156ms</span>
          <span class="uptime">98.2%</span>
        </div>
        <div class="region-indicator warning">●</div>
      </div>
    </div>
  </div>
</div>
```

## Widget Data Integration
```javascript
class StatusWidget {
  constructor(element, dataSource) {
    this.element = element;
    this.dataSource = dataSource;
    this.updateInterval = 30000; // 30 seconds
    this.init();
  }

  async init() {
    await this.updateData();
    this.startAutoUpdate();
    this.bindEvents();
  }

  async updateData() {
    try {
      const data = await this.dataSource.getStatusData();
      this.render(data);
    } catch (error) {
      this.showError(error);
    }
  }

  render(data) {
    // Update widget content with new data
    this.updateStatusIndicators(data.status);
    this.updateCounters(data.counts);
    this.updateTimestamps(data.timestamps);
  }

  startAutoUpdate() {
    setInterval(() => this.updateData(), this.updateInterval);
  }
}
```

## Status Indicators
- **Healthy**: Green (●) - All endpoints responding normally
- **Warning**: Yellow (●) - Some endpoints above threshold
- **Critical**: Red (●) - Endpoints down or major issues
- **Unknown**: Gray (●) - No recent data available

## Widget Types
1. **System Status**: Overall monitoring system health
2. **Endpoint Summary**: Count of healthy/warning/down endpoints
3. **Regional Health**: Health overview by geographic region
4. **Latest Results**: Most recent test results summary
5. **Performance Metrics**: System resource usage
6. **Alert Summary**: Active alerts and warnings

## Real-time Updates
- WebSocket integration for live data updates
- Smooth animations for status changes
- Timestamp tracking for data freshness
- Connection status indicators
- Auto-retry for failed updates

## Interactive Features
- Click widgets to navigate to detailed views
- Hover tooltips with additional information
- Refresh buttons for manual data updates
- Expandable widgets for more details
- Context menus for quick actions

## Verification Steps
1. Display system status - should show current monitoring state
2. Update endpoint counters - should reflect real endpoint status
3. Test real-time updates - should update automatically without page refresh
4. Verify status indicators - should use correct colors and states
5. Test widget interactions - should navigate to appropriate detailed views
6. Verify responsive behavior - should adapt to different screen sizes
7. Test error handling - should display errors gracefully
8. Verify accessibility - should support screen readers and keyboard navigation

## Dependencies
- T026: Dashboard Layout and Structure
- T015: Monitoring Status API
- T005: Wails Frontend-Backend Integration

## Notes
- Use consistent visual design language across all widgets
- Implement smooth transitions for status changes
- Consider using skeleton loading states
- Optimize for performance with many widgets
- Plan for future widget customization options
- Ensure widgets work well on touch devices