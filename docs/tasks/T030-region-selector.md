# T030: Region Selector Component

## Overview
Implement a region selector component that allows users to filter dashboard views by geographical regions and provides regional health information.

## Context
NetMonitor organizes endpoints by geographical regions (e.g., NA-East, EU-West, Asia-Pacific). Users need to view data for specific regions and compare regional performance easily.

## Task Description
Create a comprehensive region selector component with dropdown interface, regional health indicators, multi-selection capability, and integration with dashboard filtering.

## Acceptance Criteria
- [ ] Dropdown region selector with health indicators
- [ ] Multi-region selection capability
- [ ] Regional health status visualization
- [ ] Integration with dashboard graphs and grids
- [ ] "All Regions" option for complete view
- [ ] Recent regional performance metrics
- [ ] Responsive design for mobile devices
- [ ] Keyboard navigation support
- [ ] Regional endpoint count display
- [ ] Quick region health comparison

## Region Selector Component
```html
<div class="region-selector-container">
  <div class="region-selector">
    <button class="region-selector-toggle" aria-expanded="false" aria-haspopup="listbox">
      <div class="selected-regions">
        <span class="selected-text">All Regions</span>
        <span class="selected-count">(14 endpoints)</span>
      </div>
      <span class="dropdown-arrow">▼</span>
    </button>

    <div class="region-dropdown" role="listbox" aria-label="Select regions">
      <div class="dropdown-header">
        <div class="dropdown-title">Select Regions</div>
        <div class="dropdown-actions">
          <button class="select-all-btn">All</button>
          <button class="select-none-btn">None</button>
        </div>
      </div>

      <div class="region-list">
        <div class="region-option" role="option" data-region="all" tabindex="0">
          <label class="region-checkbox">
            <input type="checkbox" value="all" checked>
            <span class="checkmark">✓</span>
          </label>
          <div class="region-info">
            <div class="region-name">All Regions</div>
            <div class="region-stats">14 endpoints</div>
          </div>
        </div>

        <div class="region-option" role="option" data-region="na-east" tabindex="0">
          <label class="region-checkbox">
            <input type="checkbox" value="na-east">
            <span class="checkmark">✓</span>
          </label>
          <div class="region-info">
            <div class="region-name">NA-East</div>
            <div class="region-stats">6 endpoints • Avg: 23ms</div>
          </div>
          <div class="region-health">
            <div class="health-indicator healthy" title="Region Healthy">●</div>
            <div class="uptime-badge">99.8%</div>
          </div>
        </div>

        <div class="region-option" role="option" data-region="eu-west" tabindex="0">
          <label class="region-checkbox">
            <input type="checkbox" value="eu-west">
            <span class="checkmark">✓</span>
          </label>
          <div class="region-info">
            <div class="region-name">EU-West</div>
            <div class="region-stats">4 endpoints • Avg: 156ms</div>
          </div>
          <div class="region-health">
            <div class="health-indicator warning" title="Region Warning">●</div>
            <div class="uptime-badge">98.2%</div>
          </div>
        </div>

        <div class="region-option" role="option" data-region="asia-pacific" tabindex="0">
          <label class="region-checkbox">
            <input type="checkbox" value="asia-pacific">
            <span class="checkmark">✓</span>
          </label>
          <div class="region-info">
            <div class="region-name">Asia-Pacific</div>
            <div class="region-stats">4 endpoints • Avg: 234ms</div>
          </div>
          <div class="region-health">
            <div class="health-indicator healthy" title="Region Healthy">●</div>
            <div class="uptime-badge">99.1%</div>
          </div>
        </div>
      </div>

      <div class="dropdown-footer">
        <button class="apply-selection-btn">Apply Selection</button>
      </div>
    </div>
  </div>

  <div class="region-quick-stats">
    <div class="quick-stat">
      <span class="stat-label">Total Endpoints:</span>
      <span class="stat-value">14</span>
    </div>
    <div class="quick-stat">
      <span class="stat-label">Healthy Regions:</span>
      <span class="stat-value healthy">2</span>
    </div>
    <div class="quick-stat">
      <span class="stat-label">Warning Regions:</span>
      <span class="stat-value warning">1</span>
    </div>
  </div>
</div>
```

## Region Selector Class
```javascript
class RegionSelector {
  constructor(container, options = {}) {
    this.container = container;
    this.options = {
      multiSelect: true,
      showHealth: true,
      autoApply: false,
      ...options
    };

    this.selectedRegions = new Set(['all']);
    this.regions = [];
    this.isOpen = false;
    this.callbacks = {
      onSelectionChange: options.onSelectionChange || (() => {})
    };

    this.init();
  }

  async init() {
    await this.loadRegions();
    this.render();
    this.bindEvents();
    this.updateDisplay();
  }

  async loadRegions() {
    try {
      const response = await window.go.App.GetRegionStatus();
      this.regions = [
        {
          id: 'all',
          name: 'All Regions',
          endpointCount: response.totalEndpoints,
          health: 'mixed'
        },
        ...Object.entries(response.regions).map(([id, region]) => ({
          id,
          name: region.name,
          endpointCount: region.endpointCount,
          averageLatency: region.averageLatency,
          uptime: region.uptime,
          health: region.health
        }))
      ];
    } catch (error) {
      console.error('Failed to load regions:', error);
    }
  }

  render() {
    const dropdown = this.container.querySelector('.region-dropdown');
    const regionList = dropdown.querySelector('.region-list');

    regionList.innerHTML = this.regions.map(region => this.renderRegionOption(region)).join('');
  }

  renderRegionOption(region) {
    const isSelected = this.selectedRegions.has(region.id);
    const isAll = region.id === 'all';

    return `
      <div class="region-option" role="option" data-region="${region.id}" tabindex="0">
        <label class="region-checkbox">
          <input type="checkbox" value="${region.id}" ${isSelected ? 'checked' : ''}>
          <span class="checkmark">✓</span>
        </label>
        <div class="region-info">
          <div class="region-name">${region.name}</div>
          <div class="region-stats">
            ${region.endpointCount} endpoints
            ${!isAll ? `• Avg: ${Math.round(region.averageLatency)}ms` : ''}
          </div>
        </div>
        ${!isAll ? `
          <div class="region-health">
            <div class="health-indicator ${region.health}" title="Region ${region.health}">●</div>
            <div class="uptime-badge">${region.uptime.toFixed(1)}%</div>
          </div>
        ` : ''}
      </div>
    `;
  }

  bindEvents() {
    const toggle = this.container.querySelector('.region-selector-toggle');
    const dropdown = this.container.querySelector('.region-dropdown');

    // Toggle dropdown
    toggle.addEventListener('click', () => this.toggleDropdown());

    // Region selection
    dropdown.addEventListener('click', (e) => {
      const checkbox = e.target.closest('input[type="checkbox"]');
      if (checkbox) {
        this.handleRegionToggle(checkbox.value, checkbox.checked);
      }
    });

    // Keyboard navigation
    dropdown.addEventListener('keydown', (e) => this.handleKeyDown(e));

    // Close on outside click
    document.addEventListener('click', (e) => {
      if (!this.container.contains(e.target)) {
        this.closeDropdown();
      }
    });

    // Quick actions
    const selectAllBtn = dropdown.querySelector('.select-all-btn');
    const selectNoneBtn = dropdown.querySelector('.select-none-btn');

    selectAllBtn.addEventListener('click', () => this.selectAll());
    selectNoneBtn.addEventListener('click', () => this.selectNone());

    // Apply selection
    const applyBtn = dropdown.querySelector('.apply-selection-btn');
    applyBtn.addEventListener('click', () => this.applySelection());
  }

  handleRegionToggle(regionId, checked) {
    if (regionId === 'all') {
      if (checked) {
        this.selectedRegions.clear();
        this.selectedRegions.add('all');
      } else {
        this.selectedRegions.delete('all');
      }
    } else {
      if (checked) {
        this.selectedRegions.delete('all');
        this.selectedRegions.add(regionId);
      } else {
        this.selectedRegions.delete(regionId);
        if (this.selectedRegions.size === 0) {
          this.selectedRegions.add('all');
        }
      }
    }

    this.updateCheckboxes();
    this.updateDisplay();

    if (this.options.autoApply) {
      this.applySelection();
    }
  }

  updateDisplay() {
    const selectedText = this.container.querySelector('.selected-text');
    const selectedCount = this.container.querySelector('.selected-count');

    if (this.selectedRegions.has('all')) {
      selectedText.textContent = 'All Regions';
      const totalEndpoints = this.regions[0]?.endpointCount || 0;
      selectedCount.textContent = `(${totalEndpoints} endpoints)`;
    } else {
      const selectedNames = Array.from(this.selectedRegions)
        .map(id => this.regions.find(r => r.id === id)?.name)
        .filter(Boolean);

      if (selectedNames.length === 1) {
        selectedText.textContent = selectedNames[0];
      } else {
        selectedText.textContent = `${selectedNames.length} Regions`;
      }

      const totalEndpoints = Array.from(this.selectedRegions)
        .reduce((total, id) => {
          const region = this.regions.find(r => r.id === id);
          return total + (region?.endpointCount || 0);
        }, 0);

      selectedCount.textContent = `(${totalEndpoints} endpoints)`;
    }
  }

  applySelection() {
    const selection = Array.from(this.selectedRegions);
    this.callbacks.onSelectionChange(selection);
    this.closeDropdown();
  }

  getSelectedRegions() {
    return Array.from(this.selectedRegions);
  }
}
```

## Integration Features

### Dashboard Integration
```javascript
// Initialize region selector with dashboard callback
const regionSelector = new RegionSelector(
  document.querySelector('.region-selector-container'),
  {
    onSelectionChange: (selectedRegions) => {
      // Update graphs
      latencyGraph.filterByRegions(selectedRegions);

      // Update endpoint grid
      endpointGrid.filterByRegions(selectedRegions);

      // Update status widgets
      statusWidgets.updateRegionFilter(selectedRegions);
    }
  }
);
```

### Regional Health Calculation
- **Healthy**: All endpoints in region are healthy
- **Warning**: Some endpoints above threshold or showing warnings
- **Critical**: One or more endpoints are down
- **Mixed**: Used for "All Regions" when regions have different health states

## Verification Steps
1. Display region selector - should show all configured regions
2. Test region selection - should update selected regions correctly
3. Verify health indicators - should show accurate regional health status
4. Test multi-selection - should allow selecting multiple regions
5. Verify dashboard integration - should filter graphs and grids correctly
6. Test keyboard navigation - should support arrow keys and enter
7. Verify responsive behavior - should work on mobile devices
8. Test accessibility - should support screen readers

## Dependencies
- T026: Dashboard Layout and Structure
- T015: Monitoring Status API
- T028: Interactive Latency Graphs
- T029: Endpoint Status Grid

## Notes
- Consider implementing region grouping for large numbers of regions
- Provide visual feedback for region health changes
- Implement smooth transitions for dropdown animations
- Consider adding region management capabilities (add/edit/delete regions)
- Plan for region-specific thresholds and settings
- Optimize for performance with many regions