# T032: Time Range Selector Component

## Overview
Create a comprehensive time range selector component that allows users to choose time periods for viewing historical data in graphs and reports.

## Context
NetMonitor displays historical network monitoring data over various time periods. Users need intuitive controls to select time ranges (last 24 hours, week, month, custom) for analyzing trends and performance over time.

## Task Description
Implement a flexible time range selector with preset options, custom date/time picking, relative time ranges, and integration with dashboard components.

## Acceptance Criteria
- [ ] Preset time range buttons (24H, 7D, 30D, etc.)
- [ ] Custom date/time range picker
- [ ] Relative time range options (last X hours/days)
- [ ] Time zone handling and display
- [ ] Real-time range updating for "live" views
- [ ] Integration with graphs and data views
- [ ] Keyboard shortcuts for common ranges
- [ ] Mobile-friendly date/time selection
- [ ] Range validation and error handling

## Time Range Selector Component
```html
<div class="time-range-selector">
  <div class="range-presets">
    <button class="preset-btn active" data-range="24h" data-label="Last 24 Hours">
      24H
    </button>
    <button class="preset-btn" data-range="7d" data-label="Last 7 Days">
      7D
    </button>
    <button class="preset-btn" data-range="30d" data-label="Last 30 Days">
      30D
    </button>
    <button class="preset-btn" data-range="90d" data-label="Last 90 Days">
      90D
    </button>
    <button class="preset-btn" data-range="custom" data-label="Custom Range">
      Custom
    </button>
  </div>

  <div class="range-display">
    <span class="range-text">Last 24 Hours</span>
    <span class="range-dates">
      Sep 26, 2025 7:30 PM - Sep 27, 2025 7:30 PM
    </span>
  </div>

  <div class="range-controls">
    <button class="range-control-btn" data-action="refresh" title="Refresh">
      <span class="btn-icon">↻</span>
    </button>
    <button class="range-control-btn" data-action="live" title="Live Mode">
      <span class="btn-icon">⚡</span>
    </button>
    <button class="range-control-btn" data-action="settings" title="Settings">
      <span class="btn-icon">⚙</span>
    </button>
  </div>

  <!-- Custom Range Picker (hidden by default) -->
  <div class="custom-range-picker" style="display: none;">
    <div class="picker-header">
      <h3>Select Custom Time Range</h3>
      <button class="close-picker-btn">×</button>
    </div>

    <div class="picker-content">
      <div class="date-time-inputs">
        <div class="input-group">
          <label for="start-date">Start Date:</label>
          <input type="date" id="start-date" class="date-input">
          <input type="time" id="start-time" class="time-input">
        </div>
        <div class="input-group">
          <label for="end-date">End Date:</label>
          <input type="date" id="end-date" class="date-input">
          <input type="time" id="end-time" class="time-input">
        </div>
      </div>

      <div class="quick-ranges">
        <h4>Quick Ranges</h4>
        <div class="quick-range-buttons">
          <button class="quick-range-btn" data-hours="1">Last Hour</button>
          <button class="quick-range-btn" data-hours="6">Last 6 Hours</button>
          <button class="quick-range-btn" data-hours="12">Last 12 Hours</button>
          <button class="quick-range-btn" data-days="3">Last 3 Days</button>
          <button class="quick-range-btn" data-days="7">Last Week</button>
          <button class="quick-range-btn" data-days="14">Last 2 Weeks</button>
        </div>
      </div>

      <div class="relative-ranges">
        <h4>Relative Ranges</h4>
        <div class="relative-inputs">
          <label>
            Last
            <input type="number" class="relative-value" value="24" min="1" max="365">
            <select class="relative-unit">
              <option value="hours">Hours</option>
              <option value="days">Days</option>
              <option value="weeks">Weeks</option>
              <option value="months">Months</option>
            </select>
          </label>
        </div>
      </div>

      <div class="timezone-selector">
        <label for="timezone-select">Timezone:</label>
        <select id="timezone-select" class="timezone-select">
          <option value="local">Local Time</option>
          <option value="UTC">UTC</option>
          <option value="America/New_York">Eastern Time</option>
          <option value="America/Los_Angeles">Pacific Time</option>
          <option value="Europe/London">GMT</option>
        </select>
      </div>
    </div>

    <div class="picker-footer">
      <button class="cancel-btn">Cancel</button>
      <button class="apply-btn">Apply Range</button>
    </div>
  </div>

  <!-- Live Mode Indicator -->
  <div class="live-mode-indicator" style="display: none;">
    <span class="live-dot">●</span>
    <span class="live-text">Live</span>
    <span class="live-update">Updated 30s ago</span>
  </div>
</div>
```

## Time Range Manager Class
```javascript
class TimeRangeSelector {
  constructor(container, options = {}) {
    this.container = container;
    this.options = {
      defaultRange: '24h',
      allowLiveMode: true,
      timezone: 'local',
      updateInterval: 30000, // 30 seconds for live mode
      maxRange: 365, // Maximum days for custom range
      ...options
    };

    this.currentRange = null;
    this.isLiveMode = false;
    this.liveUpdateTimer = null;
    this.callbacks = {
      onRangeChange: options.onRangeChange || (() => {}),
      onLiveModeToggle: options.onLiveModeToggle || (() => {})
    };

    this.presetRanges = {
      '1h': { hours: 1, label: 'Last Hour' },
      '6h': { hours: 6, label: 'Last 6 Hours' },
      '24h': { hours: 24, label: 'Last 24 Hours' },
      '7d': { days: 7, label: 'Last 7 Days' },
      '30d': { days: 30, label: 'Last 30 Days' },
      '90d': { days: 90, label: 'Last 90 Days' }
    };

    this.init();
  }

  init() {
    this.bindEvents();
    this.setRange(this.options.defaultRange);
    this.updateDisplay();
  }

  bindEvents() {
    // Preset buttons
    this.container.addEventListener('click', (e) => {
      if (e.target.matches('.preset-btn')) {
        const range = e.target.dataset.range;
        if (range === 'custom') {
          this.showCustomPicker();
        } else {
          this.setRange(range);
        }
      }
    });

    // Control buttons
    this.container.addEventListener('click', (e) => {
      if (e.target.closest('[data-action="refresh"]')) {
        this.refreshData();
      } else if (e.target.closest('[data-action="live"]')) {
        this.toggleLiveMode();
      } else if (e.target.closest('[data-action="settings"]')) {
        this.showSettings();
      }
    });

    // Custom picker
    const customPicker = this.container.querySelector('.custom-range-picker');

    customPicker.addEventListener('click', (e) => {
      if (e.target.matches('.close-picker-btn, .cancel-btn')) {
        this.hideCustomPicker();
      } else if (e.target.matches('.apply-btn')) {
        this.applyCustomRange();
      } else if (e.target.matches('.quick-range-btn')) {
        this.applyQuickRange(e.target);
      }
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
          case '1': this.setRange('24h'); e.preventDefault(); break;
          case '2': this.setRange('7d'); e.preventDefault(); break;
          case '3': this.setRange('30d'); e.preventDefault(); break;
          case 'r': this.refreshData(); e.preventDefault(); break;
          case 'l': this.toggleLiveMode(); e.preventDefault(); break;
        }
      }
    });
  }

  setRange(rangeKey, customRange = null) {
    // Stop live mode when changing range
    if (this.isLiveMode && rangeKey !== 'live') {
      this.stopLiveMode();
    }

    if (customRange) {
      this.currentRange = {
        key: 'custom',
        startTime: customRange.startTime,
        endTime: customRange.endTime,
        label: this.formatCustomRangeLabel(customRange)
      };
    } else if (this.presetRanges[rangeKey]) {
      const preset = this.presetRanges[rangeKey];
      const endTime = new Date();
      const startTime = new Date();

      if (preset.hours) {
        startTime.setHours(endTime.getHours() - preset.hours);
      } else if (preset.days) {
        startTime.setDate(endTime.getDate() - preset.days);
      }

      this.currentRange = {
        key: rangeKey,
        startTime,
        endTime,
        label: preset.label
      };
    }

    this.updateDisplay();
    this.updateActiveButton();
    this.notifyRangeChange();
  }

  showCustomPicker() {
    const picker = this.container.querySelector('.custom-range-picker');

    // Set default values
    const now = new Date();
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);

    const startDateInput = picker.querySelector('#start-date');
    const startTimeInput = picker.querySelector('#start-time');
    const endDateInput = picker.querySelector('#end-date');
    const endTimeInput = picker.querySelector('#end-time');

    startDateInput.value = this.formatDateForInput(yesterday);
    startTimeInput.value = this.formatTimeForInput(yesterday);
    endDateInput.value = this.formatDateForInput(now);
    endTimeInput.value = this.formatTimeForInput(now);

    picker.style.display = 'block';
  }

  hideCustomPicker() {
    this.container.querySelector('.custom-range-picker').style.display = 'none';
  }

  applyCustomRange() {
    const picker = this.container.querySelector('.custom-range-picker');
    const startDate = picker.querySelector('#start-date').value;
    const startTime = picker.querySelector('#start-time').value;
    const endDate = picker.querySelector('#end-date').value;
    const endTime = picker.querySelector('#end-time').value;

    if (!startDate || !endDate) {
      this.showError('Please select both start and end dates');
      return;
    }

    const startTime_dt = new Date(`${startDate}T${startTime || '00:00'}`);
    const endTime_dt = new Date(`${endDate}T${endTime || '23:59'}`);

    if (startTime_dt >= endTime_dt) {
      this.showError('End time must be after start time');
      return;
    }

    const maxRange = this.options.maxRange * 24 * 60 * 60 * 1000; // Convert days to ms
    if (endTime_dt - startTime_dt > maxRange) {
      this.showError(`Range cannot exceed ${this.options.maxRange} days`);
      return;
    }

    this.setRange('custom', {
      startTime: startTime_dt,
      endTime: endTime_dt
    });

    this.hideCustomPicker();
  }

  toggleLiveMode() {
    if (this.isLiveMode) {
      this.stopLiveMode();
    } else {
      this.startLiveMode();
    }
  }

  startLiveMode() {
    if (!this.options.allowLiveMode) return;

    this.isLiveMode = true;
    this.setRange('24h'); // Default to last 24 hours for live mode

    // Start updating the end time
    this.liveUpdateTimer = setInterval(() => {
      if (this.currentRange && this.currentRange.key !== 'custom') {
        this.updateLiveRange();
      }
    }, this.options.updateInterval);

    this.showLiveIndicator();
    this.callbacks.onLiveModeToggle(true);
  }

  stopLiveMode() {
    this.isLiveMode = false;

    if (this.liveUpdateTimer) {
      clearInterval(this.liveUpdateTimer);
      this.liveUpdateTimer = null;
    }

    this.hideLiveIndicator();
    this.callbacks.onLiveModeToggle(false);
  }

  updateLiveRange() {
    if (!this.currentRange || this.currentRange.key === 'custom') return;

    const preset = this.presetRanges[this.currentRange.key];
    const endTime = new Date();
    const startTime = new Date();

    if (preset.hours) {
      startTime.setHours(endTime.getHours() - preset.hours);
    } else if (preset.days) {
      startTime.setDate(endTime.getDate() - preset.days);
    }

    this.currentRange.startTime = startTime;
    this.currentRange.endTime = endTime;

    this.updateDisplay();
    this.notifyRangeChange();
    this.updateLiveIndicator();
  }

  updateDisplay() {
    if (!this.currentRange) return;

    const rangeText = this.container.querySelector('.range-text');
    const rangeDates = this.container.querySelector('.range-dates');

    rangeText.textContent = this.currentRange.label;
    rangeDates.textContent = this.formatDateRange(
      this.currentRange.startTime,
      this.currentRange.endTime
    );
  }

  formatDateRange(startTime, endTime) {
    const options = {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit'
    };

    const start = startTime.toLocaleDateString('en-US', options);
    const end = endTime.toLocaleDateString('en-US', options);

    return `${start} - ${end}`;
  }

  getCurrentRange() {
    return this.currentRange ? {
      startTime: this.currentRange.startTime,
      endTime: this.currentRange.endTime,
      isLive: this.isLiveMode
    } : null;
  }

  notifyRangeChange() {
    const range = this.getCurrentRange();
    if (range) {
      this.callbacks.onRangeChange(range);
    }
  }

  showLiveIndicator() {
    const indicator = this.container.querySelector('.live-mode-indicator');
    indicator.style.display = 'flex';
    this.updateLiveIndicator();
  }

  hideLiveIndicator() {
    this.container.querySelector('.live-mode-indicator').style.display = 'none';
  }

  updateLiveIndicator() {
    const updateText = this.container.querySelector('.live-update');
    updateText.textContent = 'Updated just now';
  }
}
```

## Integration with Dashboard
```javascript
// Initialize time range selector with dashboard integration
const timeRangeSelector = new TimeRangeSelector(
  document.querySelector('.time-range-selector'),
  {
    onRangeChange: (range) => {
      // Update all dashboard components
      latencyGraph.updateTimeRange(range);
      endpointGrid.updateTimeRange(range);
      statusWidgets.updateTimeRange(range);
    },
    onLiveModeToggle: (isLive) => {
      // Enable/disable real-time updates
      if (isLive) {
        dashboardUpdater.start();
      } else {
        dashboardUpdater.stop();
      }
    }
  }
);
```

## Verification Steps
1. Select preset ranges - should update display and notify components
2. Test custom range picker - should allow selecting specific date/time ranges
3. Verify live mode - should update range automatically and show indicator
4. Test keyboard shortcuts - should respond to Ctrl+1, Ctrl+2, etc.
5. Verify range validation - should prevent invalid ranges
6. Test timezone handling - should display times in selected timezone
7. Verify mobile usability - should work well on touch devices
8. Test integration - should update graphs and other components

## Dependencies
- T026: Dashboard Layout and Structure
- T028: Interactive Latency Graphs
- T021: Historical Data Queries

## Notes
- Consider implementing range bookmarking for frequently used custom ranges
- Provide clear feedback for loading states during range changes
- Implement smooth transitions between different ranges
- Consider adding calendar view for easier date selection
- Plan for future features like range comparison
- Optimize for performance with large time ranges