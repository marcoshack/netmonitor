# T031: Manual Test Interface

## Overview
Create a user interface for triggering manual network tests with real-time progress reporting, detailed results display, and test management capabilities.

## Context
Users need the ability to manually trigger network tests for troubleshooting, verification, or immediate feedback. The interface should provide detailed results beyond automated monitoring and support various test scenarios.

## Task Description
Implement a comprehensive manual testing interface with test selection, execution controls, real-time progress monitoring, and detailed results presentation.

## Acceptance Criteria
- [ ] Test selection interface (individual endpoints, regions, all)
- [ ] Real-time test execution progress reporting
- [ ] Detailed test results with timing breakdown
- [ ] Test cancellation capability
- [ ] Test history and results comparison
- [ ] Export test results functionality
- [ ] Batch test execution for multiple endpoints
- [ ] Test scheduling for future execution
- [ ] Visual indicators for test status and progress

## Manual Test Interface
```html
<div class="manual-test-container">
  <div class="test-header">
    <h2 class="test-title">Manual Network Tests</h2>
    <div class="test-actions">
      <button class="test-history-btn">📊 Test History</button>
      <button class="export-results-btn">📤 Export Results</button>
    </div>
  </div>

  <div class="test-selector">
    <div class="selector-tabs">
      <button class="tab-btn active" data-tab="single">Single Endpoint</button>
      <button class="tab-btn" data-tab="region">Region</button>
      <button class="tab-btn" data-tab="all">All Endpoints</button>
      <button class="tab-btn" data-tab="custom">Custom Selection</button>
    </div>

    <div class="tab-content">
      <!-- Single Endpoint Tab -->
      <div class="tab-panel active" data-tab="single">
        <div class="endpoint-selector">
          <select class="endpoint-select">
            <option value="">Select an endpoint...</option>
            <option value="na-east-google-dns">Google DNS (NA-East)</option>
            <option value="eu-west-cloudflare">Cloudflare (EU-West)</option>
          </select>
        </div>
      </div>

      <!-- Region Tab -->
      <div class="tab-panel" data-tab="region">
        <div class="region-selector">
          <select class="region-select">
            <option value="">Select a region...</option>
            <option value="na-east">NA-East (6 endpoints)</option>
            <option value="eu-west">EU-West (4 endpoints)</option>
          </select>
        </div>
      </div>

      <!-- All Endpoints Tab -->
      <div class="tab-panel" data-tab="all">
        <div class="all-endpoints-info">
          <div class="info-item">
            <span class="info-label">Total Endpoints:</span>
            <span class="info-value">14</span>
          </div>
          <div class="info-item">
            <span class="info-label">Estimated Duration:</span>
            <span class="info-value">~45 seconds</span>
          </div>
        </div>
      </div>

      <!-- Custom Selection Tab -->
      <div class="tab-panel" data-tab="custom">
        <div class="endpoint-checklist">
          <div class="checklist-header">
            <label class="select-all">
              <input type="checkbox" id="select-all-endpoints">
              Select All
            </label>
            <span class="selected-count">0 selected</span>
          </div>
          <div class="endpoint-list">
            <label class="endpoint-checkbox">
              <input type="checkbox" value="na-east-google-dns">
              <span class="endpoint-name">Google DNS (NA-East)</span>
              <span class="endpoint-type">ICMP</span>
            </label>
            <!-- More endpoints... -->
          </div>
        </div>
      </div>
    </div>
  </div>

  <div class="test-options">
    <div class="option-group">
      <label for="test-timeout">Timeout (seconds):</label>
      <input type="number" id="test-timeout" value="10" min="1" max="60">
    </div>
    <div class="option-group">
      <label for="test-repeat">Repeat Count:</label>
      <input type="number" id="test-repeat" value="1" min="1" max="10">
    </div>
    <div class="option-group">
      <label>
        <input type="checkbox" id="detailed-timing">
        Include detailed timing breakdown
      </label>
    </div>
  </div>

  <div class="test-controls">
    <button class="start-test-btn" disabled>
      <span class="btn-icon">▶</span>
      Start Test
    </button>
    <button class="cancel-test-btn" style="display: none;">
      <span class="btn-icon">⏹</span>
      Cancel Test
    </button>
  </div>

  <div class="test-progress" style="display: none;">
    <div class="progress-header">
      <h3 class="progress-title">Running Tests...</h3>
      <span class="progress-status">2 of 6 completed</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" style="width: 33%"></div>
    </div>
    <div class="current-test">
      <span class="current-test-label">Testing:</span>
      <span class="current-test-name">Cloudflare DNS (EU-West)</span>
      <span class="current-test-status">Connecting...</span>
    </div>
  </div>

  <div class="test-results" style="display: none;">
    <div class="results-header">
      <h3 class="results-title">Test Results</h3>
      <div class="results-summary">
        <span class="summary-item success">4 Successful</span>
        <span class="summary-item warning">1 Warning</span>
        <span class="summary-item failed">1 Failed</span>
      </div>
    </div>
    <div class="results-grid">
      <!-- Dynamic test results -->
    </div>
  </div>
</div>
```

## Test Result Card
```html
<div class="test-result-card" data-status="success">
  <div class="result-header">
    <div class="endpoint-info">
      <span class="endpoint-name">Google DNS</span>
      <span class="endpoint-address">8.8.8.8</span>
      <span class="endpoint-region">NA-East</span>
    </div>
    <div class="result-status success">
      <span class="status-icon">✓</span>
      <span class="status-text">Success</span>
    </div>
  </div>

  <div class="result-metrics">
    <div class="metric-item">
      <span class="metric-label">Latency:</span>
      <span class="metric-value">23ms</span>
    </div>
    <div class="metric-item">
      <span class="metric-label">Status:</span>
      <span class="metric-value">200 OK</span>
    </div>
    <div class="metric-item">
      <span class="metric-label">Response Size:</span>
      <span class="metric-value">512 bytes</span>
    </div>
  </div>

  <div class="result-timing" style="display: none;">
    <div class="timing-breakdown">
      <div class="timing-item">
        <span class="timing-phase">DNS Lookup:</span>
        <span class="timing-value">2ms</span>
      </div>
      <div class="timing-item">
        <span class="timing-phase">Connection:</span>
        <span class="timing-value">8ms</span>
      </div>
      <div class="timing-item">
        <span class="timing-phase">TLS Handshake:</span>
        <span class="timing-value">12ms</span>
      </div>
      <div class="timing-item">
        <span class="timing-phase">Response:</span>
        <span class="timing-value">1ms</span>
      </div>
    </div>
  </div>

  <div class="result-actions">
    <button class="toggle-details-btn">Show Details</button>
    <button class="retest-btn">Test Again</button>
    <button class="copy-result-btn">Copy</button>
  </div>
</div>
```

## Manual Test Manager
```javascript
class ManualTestManager {
  constructor(container) {
    this.container = container;
    this.currentTest = null;
    this.testResults = [];
    this.isTestRunning = false;

    this.init();
  }

  init() {
    this.bindEvents();
    this.updateUI();
  }

  bindEvents() {
    // Tab switching
    this.container.addEventListener('click', (e) => {
      if (e.target.matches('.tab-btn')) {
        this.switchTab(e.target.dataset.tab);
      }
    });

    // Test selection changes
    this.container.addEventListener('change', (e) => {
      if (e.target.matches('select, input[type="checkbox"]')) {
        this.updateStartButton();
      }
    });

    // Start test
    const startBtn = this.container.querySelector('.start-test-btn');
    startBtn.addEventListener('click', () => this.startTest());

    // Cancel test
    const cancelBtn = this.container.querySelector('.cancel-test-btn');
    cancelBtn.addEventListener('click', () => this.cancelTest());

    // Result actions
    this.container.addEventListener('click', (e) => {
      if (e.target.matches('.toggle-details-btn')) {
        this.toggleResultDetails(e.target.closest('.test-result-card'));
      }
      if (e.target.matches('.retest-btn')) {
        this.retestEndpoint(e.target.closest('.test-result-card'));
      }
    });
  }

  async startTest() {
    if (this.isTestRunning) return;

    const testConfig = this.buildTestConfig();
    if (!testConfig.targets.length) {
      this.showError('No endpoints selected for testing');
      return;
    }

    this.isTestRunning = true;
    this.showProgress();
    this.updateUI();

    try {
      if (testConfig.targets.length === 1) {
        await this.runSingleTest(testConfig);
      } else {
        await this.runBatchTest(testConfig);
      }
    } catch (error) {
      this.showError(`Test failed: ${error.message}`);
    } finally {
      this.isTestRunning = false;
      this.hideProgress();
      this.updateUI();
    }
  }

  async runSingleTest(config) {
    this.updateProgress(0, 1, 'Starting test...');

    try {
      const result = await window.go.App.RunManualTest(config.targets[0]);
      this.updateProgress(1, 1, 'Test completed');
      this.displayResults([result]);
    } catch (error) {
      throw new Error(`Single test failed: ${error.message}`);
    }
  }

  async runBatchTest(config) {
    const totalTests = config.targets.length;
    const results = [];

    for (let i = 0; i < totalTests; i++) {
      if (!this.isTestRunning) break; // Check for cancellation

      const target = config.targets[i];
      this.updateProgress(i, totalTests, `Testing ${target.name}...`);

      try {
        const result = await window.go.App.RunManualTest(target.id);
        results.push(result);
      } catch (error) {
        results.push({
          endpointId: target.id,
          endpointName: target.name,
          status: 'failed',
          error: error.message
        });
      }

      // Small delay between tests to prevent overwhelming
      await new Promise(resolve => setTimeout(resolve, 100));
    }

    this.displayResults(results);
  }

  buildTestConfig() {
    const activeTab = this.container.querySelector('.tab-btn.active').dataset.tab;
    const timeout = parseInt(this.container.querySelector('#test-timeout').value) * 1000;
    const repeatCount = parseInt(this.container.querySelector('#test-repeat').value);
    const detailedTiming = this.container.querySelector('#detailed-timing').checked;

    let targets = [];

    switch (activeTab) {
      case 'single':
        const endpointSelect = this.container.querySelector('.endpoint-select');
        if (endpointSelect.value) {
          targets = [{ id: endpointSelect.value, name: endpointSelect.selectedOptions[0].text }];
        }
        break;

      case 'region':
        const regionSelect = this.container.querySelector('.region-select');
        if (regionSelect.value) {
          targets = this.getEndpointsForRegion(regionSelect.value);
        }
        break;

      case 'all':
        targets = this.getAllEndpoints();
        break;

      case 'custom':
        const checkedBoxes = this.container.querySelectorAll('.endpoint-checkbox input:checked');
        targets = Array.from(checkedBoxes).map(cb => ({
          id: cb.value,
          name: cb.closest('label').querySelector('.endpoint-name').textContent
        }));
        break;
    }

    return {
      targets,
      timeout,
      repeatCount,
      detailedTiming
    };
  }

  displayResults(results) {
    const resultsContainer = this.container.querySelector('.test-results');
    const resultsGrid = resultsContainer.querySelector('.results-grid');

    // Update summary
    const summary = this.calculateResultsSummary(results);
    this.updateResultsSummary(summary);

    // Render result cards
    resultsGrid.innerHTML = results.map(result => this.renderResultCard(result)).join('');

    // Show results
    resultsContainer.style.display = 'block';
    resultsContainer.scrollIntoView({ behavior: 'smooth' });

    this.testResults = results;
  }

  renderResultCard(result) {
    const statusClass = result.status === 'success' ? 'success' :
                       result.status === 'warning' ? 'warning' : 'failed';

    return `
      <div class="test-result-card" data-status="${statusClass}">
        <div class="result-header">
          <div class="endpoint-info">
            <span class="endpoint-name">${result.endpointName}</span>
            <span class="endpoint-address">${result.address}</span>
            <span class="endpoint-region">${result.region}</span>
          </div>
          <div class="result-status ${statusClass}">
            <span class="status-icon">${this.getStatusIcon(result.status)}</span>
            <span class="status-text">${result.status}</span>
          </div>
        </div>
        ${this.renderResultMetrics(result)}
        ${result.detailedTiming ? this.renderDetailedTiming(result) : ''}
        <div class="result-actions">
          <button class="toggle-details-btn">Show Details</button>
          <button class="retest-btn">Test Again</button>
          <button class="copy-result-btn">Copy</button>
        </div>
      </div>
    `;
  }

  updateProgress(completed, total, currentAction) {
    const progressContainer = this.container.querySelector('.test-progress');
    const progressFill = progressContainer.querySelector('.progress-fill');
    const progressStatus = progressContainer.querySelector('.progress-status');
    const currentTestStatus = progressContainer.querySelector('.current-test-status');

    const percentage = total > 0 ? (completed / total) * 100 : 0;

    progressFill.style.width = `${percentage}%`;
    progressStatus.textContent = `${completed} of ${total} completed`;
    currentTestStatus.textContent = currentAction;
  }

  showProgress() {
    this.container.querySelector('.test-progress').style.display = 'block';
  }

  hideProgress() {
    this.container.querySelector('.test-progress').style.display = 'none';
  }

  async cancelTest() {
    if (!this.isTestRunning) return;

    try {
      await window.go.App.CancelManualTests();
      this.isTestRunning = false;
      this.hideProgress();
      this.updateUI();
      this.showMessage('Test cancelled by user');
    } catch (error) {
      console.error('Failed to cancel test:', error);
    }
  }
}
```

## Verification Steps
1. Select single endpoint and run test - should execute and show detailed results
2. Select region and run batch test - should test all endpoints in region
3. Test with custom selection - should test only selected endpoints
4. Verify progress reporting - should show real-time progress during execution
5. Test cancellation - should stop running tests gracefully
6. Verify detailed timing - should show timing breakdown when enabled
7. Test result export - should export results in various formats
8. Verify error handling - should handle test failures gracefully

## Dependencies
- T026: Dashboard Layout and Structure
- T012: Manual Test Execution
- T015: Monitoring Status API
- T014: Endpoint Management

## Notes
- Implement proper error handling for network failures
- Consider implementing test presets for common scenarios
- Provide clear feedback for long-running batch tests
- Plan for test result archiving and comparison features
- Optimize UI for both desktop and mobile use
- Consider implementing test scheduling for future execution