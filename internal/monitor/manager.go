package monitor

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"NetMonitor/internal/config"
	"NetMonitor/internal/storage"
)

// Manager handles network monitoring operations
type Manager struct {
	config    *config.Manager
	storage   *storage.Manager
	logger    *slog.Logger
	running   bool
	stopChan  chan struct{}
	mutex     sync.RWMutex
}

// TestStatus represents the status of a network test
type TestStatus string

const (
	TestStatusSuccess TestStatus = "success"
	TestStatusFailed  TestStatus = "failed"
	TestStatusTimeout TestStatus = "timeout"
)

// NewManager creates a new monitoring manager
func NewManager(configMgr *config.Manager, storageMgr *storage.Manager, logger *slog.Logger) (*Manager, error) {
	return &Manager{
		config:   configMgr,
		storage:  storageMgr,
		logger:   logger,
		running:  false,
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins the monitoring process
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("monitoring is already running")
	}

	m.logger.Info("Starting network monitoring")
	m.running = true

	// Start monitoring loop in goroutine
	go m.monitoringLoop()

	return nil
}

// Stop stops the monitoring process
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return nil
	}

	m.logger.Info("Stopping network monitoring")
	m.running = false
	close(m.stopChan)

	return nil
}

// IsRunning returns whether monitoring is currently active
func (m *Manager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// RunManualTest executes a single test for the specified endpoint
func (m *Manager) RunManualTest(endpointID string) (*storage.TestResult, error) {
	m.logger.Info("Running manual test", "endpoint_id", endpointID)

	// For now, return a mock result
	// This will be implemented with actual network testing later
	result := &storage.TestResult{
		Timestamp:  time.Now(),
		EndpointID: endpointID,
		Protocol:   "ICMP",
		Latency:    25 * time.Millisecond,
		Status:     string(TestStatusSuccess),
	}

	// Store the result
	if err := m.storage.StoreTestResult(result); err != nil {
		m.logger.Error("Failed to store test result", "error", err)
		return result, fmt.Errorf("failed to store test result: %w", err)
	}

	return result, nil
}

// monitoringLoop runs the main monitoring loop
func (m *Manager) monitoringLoop() {
	config := m.config.GetConfig()
	interval := time.Duration(config.Settings.TestIntervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.logger.Info("Monitoring loop started", "interval", interval)

	for {
		select {
		case <-ticker.C:
			m.runScheduledTests()
		case <-m.stopChan:
			m.logger.Info("Monitoring loop stopped")
			return
		}
	}
}

// runScheduledTests executes tests for all configured endpoints
func (m *Manager) runScheduledTests() {
	config := m.config.GetConfig()
	
	for regionName, region := range config.Regions {
		for _, endpoint := range region.Endpoints {
			endpointID := fmt.Sprintf("%s-%s", regionName, endpoint.Name)
			
			// Run test for this endpoint
			result, err := m.executeTest(endpoint, endpointID)
			if err != nil {
				m.logger.Error("Failed to execute test", 
					"endpoint_id", endpointID, 
					"error", err)
				continue
			}

			// Store result
			if err := m.storage.StoreTestResult(result); err != nil {
				m.logger.Error("Failed to store test result", 
					"endpoint_id", endpointID, 
					"error", err)
			}
		}
	}
}

// executeTest executes a network test for an endpoint
func (m *Manager) executeTest(endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error) {
	// Mock implementation - will be replaced with actual network testing
	result := &storage.TestResult{
		Timestamp:  time.Now(),
		EndpointID: endpointID,
		Protocol:   endpoint.Type,
		Latency:    time.Duration(25+len(endpoint.Name)) * time.Millisecond, // Vary by name for demo
		Status:     string(TestStatusSuccess),
	}

	m.logger.Debug("Test executed",
		"endpoint_id", endpointID,
		"latency", result.Latency,
		"status", result.Status)

	return result, nil
}