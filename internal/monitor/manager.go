package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"NetMonitor/internal/config"
	"NetMonitor/internal/storage"
)

// Manager handles network monitoring operations
type Manager struct {
	config    *config.Manager
	storage   *storage.Manager
	ctx       context.Context
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
func NewManager(ctx context.Context, configMgr *config.Manager, storageMgr *storage.Manager) (*Manager, error) {
	return &Manager{
		config:   configMgr,
		storage:  storageMgr,
		ctx:      ctx,
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

	log.Ctx(m.ctx).Info().Msg("Starting network monitoring")
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

	log.Ctx(m.ctx).Info().Msg("Stopping network monitoring")
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
func (m *Manager) RunManualTest(ctx context.Context, endpointID string) (*storage.TestResult, error) {
	log.Ctx(ctx).Info().Str("endpoint_id", endpointID).Msg("Running manual test")

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
		log.Ctx(ctx).Error().Err(err).Msg("Failed to store test result")
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

	log.Ctx(m.ctx).Info().Dur("interval", interval).Msg("Monitoring loop started")

	for {
		select {
		case <-ticker.C:
			m.runScheduledTests()
		case <-m.stopChan:
			log.Ctx(m.ctx).Info().Msg("Monitoring loop stopped")
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
				log.Ctx(m.ctx).Error().
					Str("endpoint_id", endpointID).
					Err(err).
					Msg("Failed to execute test")
				continue
			}

			// Store result
			if err := m.storage.StoreTestResult(result); err != nil {
				log.Ctx(m.ctx).Error().
					Str("endpoint_id", endpointID).
					Err(err).
					Msg("Failed to store test result")
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

	log.Ctx(m.ctx).Debug().
		Str("endpoint_id", endpointID).
		Dur("latency", result.Latency).
		Str("status", result.Status).
		Msg("Test executed")

	return result, nil
}