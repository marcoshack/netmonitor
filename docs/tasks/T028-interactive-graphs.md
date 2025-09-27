# T028: Interactive Latency Graphs

## Overview
Implement interactive time series graphs for displaying network latency trends with Chart.js, including zoom, pan, time range selection, and multi-endpoint comparison.

## Context
NetMonitor needs to display latency trends over time through interactive graphs. Users should be able to analyze historical data, compare endpoints, and examine different time periods with intuitive graph controls.

## Task Description
Create comprehensive interactive graphs using Chart.js that display latency trends with full interactivity including zooming, panning, time range selection, and endpoint comparison features.

## Acceptance Criteria
- [ ] Line charts for latency trends over time
- [ ] Interactive zoom and pan functionality
- [ ] Time range selector (24h, week, month, custom)
- [ ] Multi-endpoint comparison on single graph
- [ ] Real-time data updates
- [ ] Responsive graph sizing
- [ ] Customizable graph appearance
- [ ] Export graph as image
- [ ] Tooltip with detailed information
- [ ] Performance optimization for large datasets

## Chart.js Integration
```javascript
class LatencyGraph {
  constructor(canvasElement, options = {}) {
    this.canvas = canvasElement;
    this.options = {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 0 // Disabled for real-time updates
      },
      scales: {
        x: {
          type: 'time',
          time: {
            unit: 'minute',
            displayFormats: {
              minute: 'HH:mm',
              hour: 'MMM DD HH:mm'
            }
          },
          title: {
            display: true,
            text: 'Time'
          }
        },
        y: {
          title: {
            display: true,
            text: 'Latency (ms)'
          },
          beginAtZero: true
        }
      },
      plugins: {
        zoom: {
          pan: {
            enabled: true,
            mode: 'x'
          },
          zoom: {
            wheel: {
              enabled: true
            },
            pinch: {
              enabled: true
            },
            mode: 'x'
          }
        },
        tooltip: {
          mode: 'index',
          intersect: false,
          callbacks: {
            title: function(context) {
              return new Date(context[0].parsed.x).toLocaleString();
            },
            label: function(context) {
              return `${context.dataset.label}: ${context.parsed.y}ms`;
            }
          }
        }
      },
      ...options
    };

    this.init();
  }

  init() {
    this.chart = new Chart(this.canvas, {
      type: 'line',
      data: {
        datasets: []
      },
      options: this.options
    });
  }

  updateData(timeSeriesData) {
    this.chart.data.datasets = this.formatDatasets(timeSeriesData);
    this.chart.update('none'); // No animation for real-time updates
  }

  formatDatasets(data) {
    return data.series.map((series, index) => ({
      label: series.endpointName,
      data: series.points.map(point => ({
        x: point.timestamp,
        y: point.value
      })),
      borderColor: this.getColorForIndex(index),
      backgroundColor: this.getColorForIndex(index, 0.1),
      borderWidth: 2,
      fill: false,
      pointRadius: 0, // Hide points for performance
      pointHoverRadius: 5
    }));
  }
}
```

## Graph Components

### Time Range Selector
```html
<div class="graph-controls">
  <div class="time-range-selector">
    <button class="range-btn active" data-range="24h">24H</button>
    <button class="range-btn" data-range="7d">7D</button>
    <button class="range-btn" data-range="30d">30D</button>
    <button class="range-btn" data-range="custom">Custom</button>
  </div>
  <div class="endpoint-selector">
    <select multiple class="endpoint-multiselect">
      <option value="all" selected>All Endpoints</option>
      <option value="na-east-google">Google DNS (NA-East)</option>
      <option value="eu-west-cloudflare">Cloudflare (EU-West)</option>
    </select>
  </div>
</div>
```

### Graph Container
```html
<div class="graph-container">
  <div class="graph-header">
    <h3 class="graph-title">Latency Trends</h3>
    <div class="graph-actions">
      <button class="reset-zoom-btn">Reset Zoom</button>
      <button class="export-btn">Export</button>
      <button class="fullscreen-btn">⛶</button>
    </div>
  </div>
  <div class="graph-content">
    <canvas id="latency-chart"></canvas>
  </div>
  <div class="graph-legend">
    <div class="legend-item">
      <div class="legend-color" style="background: #007bff;"></div>
      <span class="legend-label">Google DNS (NA-East)</span>
      <span class="legend-stats">Avg: 23ms</span>
    </div>
  </div>
</div>
```

## Graph Features

### Real-time Updates
```javascript
class RealTimeGraphUpdater {
  constructor(graph, dataSource) {
    this.graph = graph;
    this.dataSource = dataSource;
    this.updateInterval = 30000; // 30 seconds
    this.isUpdating = false;
  }

  start() {
    this.isUpdating = true;
    this.update();
    this.intervalId = setInterval(() => this.update(), this.updateInterval);
  }

  stop() {
    this.isUpdating = false;
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
  }

  async update() {
    if (!this.isUpdating) return;

    try {
      const newData = await this.dataSource.getLatestData();
      this.graph.appendData(newData);
    } catch (error) {
      console.error('Failed to update graph data:', error);
    }
  }
}
```

### Data Sampling for Performance
```javascript
class DataSampler {
  static sampleData(data, maxPoints) {
    if (data.length <= maxPoints) return data;

    const step = Math.ceil(data.length / maxPoints);
    const sampled = [];

    for (let i = 0; i < data.length; i += step) {
      // Use average of points in the step range
      const stepData = data.slice(i, i + step);
      const avgValue = stepData.reduce((sum, point) => sum + point.value, 0) / stepData.length;

      sampled.push({
        timestamp: stepData[Math.floor(stepData.length / 2)].timestamp,
        value: Math.round(avgValue * 100) / 100 // Round to 2 decimal places
      });
    }

    return sampled;
  }
}
```

## Graph Types
1. **Latency Trends**: Line chart showing latency over time
2. **Availability Chart**: Bar chart showing uptime percentages
3. **Comparison View**: Multiple endpoints on single chart
4. **Regional Overview**: Separate charts for each region
5. **Performance Distribution**: Histogram of latency values

## Interactive Features
- **Zoom**: Mouse wheel and pinch gestures
- **Pan**: Drag to move through time
- **Crosshair**: Show values at cursor position
- **Range Selection**: Select time ranges with mouse
- **Legend Interaction**: Click to show/hide series
- **Tooltip**: Detailed information on hover

## Performance Optimizations
- Data point sampling for large datasets
- Disabled animations for real-time updates
- Canvas optimization for smooth rendering
- Efficient data structure updates
- Memory management for long-running graphs

## Verification Steps
1. Display latency data - should render line chart with time series data
2. Test zoom functionality - should zoom in/out on mouse wheel
3. Test pan functionality - should allow dragging to move through time
4. Verify time range selection - should update data for selected ranges
5. Test multi-endpoint comparison - should display multiple series
6. Verify real-time updates - should append new data automatically
7. Test responsive behavior - should resize appropriately
8. Verify export functionality - should generate downloadable images

## Dependencies
- T026: Dashboard Layout and Structure
- T021: Historical Data Queries
- T015: Monitoring Status API
- Chart.js library

## Notes
- Use Chart.js v4+ for best performance and features
- Consider using Web Workers for heavy data processing
- Implement proper error handling for data loading failures
- Plan for accessibility features (keyboard navigation, screen readers)
- Consider implementing graph presets for common views
- Optimize for both desktop and mobile interactions