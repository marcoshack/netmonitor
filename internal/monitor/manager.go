package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marcoshack/netmonitor/internal/config"
	"github.com/marcoshack/netmonitor/internal/network"
	"github.com/marcoshack/netmonitor/internal/scheduler"
	"github.com/marcoshack/netmonitor/internal/storage"
)

// Manager handles network monitoring operations
type Manager struct {
	config    *config.Manager
	storage   *storage.Manager
	scheduler *scheduler.TestScheduler
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
	// Create scheduler with max 10 concurrent tests
	testScheduler, err := scheduler.NewScheduler(ctx, configMgr, storageMgr, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to create test scheduler: %w", err)
	}

	return &Manager{
		config:    configMgr,
		storage:   storageMgr,
		scheduler: testScheduler,
		ctx:       ctx,
		running:   false,
		stopChan:  make(chan struct{}),
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

	// Start the scheduler
	if err := m.scheduler.Start(m.ctx); err != nil {
		m.running = false
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

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

	// Stop the scheduler
	if err := m.scheduler.Stop(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

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

	// Find the endpoint in configuration
	cfg := m.config.GetConfig()
	var endpoint *config.Endpoint
	var regionName string

	for rName, region := range cfg.Regions {
		for _, ep := range region.Endpoints {
			if fmt.Sprintf("%s-%s", rName, ep.Name) == endpointID {
				endpoint = ep
				regionName = rName
				break
			}
		}
		if endpoint != nil {
			break
		}
	}

	if endpoint == nil {
		return nil, fmt.Errorf("endpoint not found: %s", endpointID)
	}

	// Execute the actual network test
	result, err := m.executeTest(endpoint, endpointID)
	if err != nil {
		log.Ctx(ctx).Error().
			Str("endpoint_id", endpointID).
			Str("region", regionName).
			Err(err).
			Msg("Manual test execution failed")

		// Still store failed results
		if result != nil {
			if storeErr := m.storage.StoreTestResult(result); storeErr != nil {
				log.Ctx(ctx).Error().Err(storeErr).Msg("Failed to store test result")
			}
		}
		return result, err
	}

	// Store the result
	if err := m.storage.StoreTestResult(result); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to store test result")
		return result, fmt.Errorf("failed to store test result: %w", err)
	}

	log.Ctx(ctx).Info().
		Str("endpoint_id", endpointID).
		Str("region", regionName).
		Dur("latency", result.Latency).
		Str("status", result.Status).
		Msg("Manual test completed successfully")

	return result, nil
}

// GetSchedulerStatus returns the scheduler status
func (m *Manager) GetSchedulerStatus() scheduler.SchedulerStatus {
	return m.scheduler.GetStatus()
}

// executeTest executes a network test for an endpoint
func (m *Manager) executeTest(endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(m.ctx, time.Duration(endpoint.Timeout)*time.Millisecond)
	defer cancel()

	var networkTest network.NetworkTest
	var testConfig network.TestConfig

	// Create the appropriate test based on protocol type
	switch endpoint.Type {
	case "ICMP":
		networkTest = &network.ICMPTest{}
		testConfig = network.TestConfig{
			Name:     endpoint.Name,
			Address:  endpoint.Address,
			Timeout:  time.Duration(endpoint.Timeout) * time.Millisecond,
			Protocol: "ICMP",
			Config: &network.ICMPConfig{
				Count:      1,
				PacketSize: 64,
				TTL:        64,
				Privileged: false,
			},
		}
	default:
		return nil, fmt.Errorf("unsupported protocol type: %s", endpoint.Type)
	}

	// Validate configuration
	if err := networkTest.Validate(testConfig); err != nil {
		return nil, fmt.Errorf("invalid test configuration: %w", err)
	}

	// Execute the network test
	networkResult, err := networkTest.Execute(ctx, testConfig)
	if err != nil {
		// Return a failed test result even on error
		return &storage.TestResult{
			Timestamp:  time.Now(),
			EndpointID: endpointID,
			Protocol:   endpoint.Type,
			Latency:    0,
			Status:     string(network.TestStatusFailed),
			Error:      err.Error(),
		}, err
	}

	// Convert network.TestResult to storage.TestResult
	storageResult := &storage.TestResult{
		Timestamp:  networkResult.Timestamp,
		EndpointID: endpointID,
		Protocol:   networkResult.Protocol,
		Latency:    networkResult.Latency,
		Status:     string(networkResult.Status),
		Error:      networkResult.Error,
	}

	log.Ctx(m.ctx).Debug().
		Str("endpoint_id", endpointID).
		Float64("latencyInMs", float64(storageResult.Latency.Nanoseconds())/1_000_000.0).
		Str("status", storageResult.Status).
		Msg("Test executed")

	return storageResult, nil
}