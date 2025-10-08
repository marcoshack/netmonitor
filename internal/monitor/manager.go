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

// RunManualTestDetailed executes a single test with detailed timing information
func (m *Manager) RunManualTestDetailed(ctx context.Context, endpointID string) (*storage.DetailedTestResult, error) {
	log.Ctx(ctx).Info().Str("endpoint_id", endpointID).Msg("Running detailed manual test")

	startTime := time.Now()

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
	executionTime := time.Since(startTime)

	// Create detailed result
	detailedResult := &storage.DetailedTestResult{
		TestResult:    *result,
		ExecutionTime: executionTime,
		IntermediateSteps: []string{
			fmt.Sprintf("Test started at %s", startTime.Format(time.RFC3339)),
			fmt.Sprintf("Protocol: %s", endpoint.Type),
			fmt.Sprintf("Target: %s", endpoint.Address),
		},
	}

	if err != nil {
		detailedResult.IntermediateSteps = append(detailedResult.IntermediateSteps,
			fmt.Sprintf("Test failed: %v", err))
		log.Ctx(ctx).Error().
			Str("endpoint_id", endpointID).
			Str("region", regionName).
			Err(err).
			Msg("Detailed manual test execution failed")

		// Still store failed results
		if storeErr := m.storage.StoreTestResult(result); storeErr != nil {
			log.Ctx(ctx).Error().Err(storeErr).Msg("Failed to store test result")
		}
		return detailedResult, err
	}

	detailedResult.IntermediateSteps = append(detailedResult.IntermediateSteps,
		fmt.Sprintf("Test completed successfully in %v", executionTime),
		fmt.Sprintf("Latency: %v", result.Latency),
		fmt.Sprintf("Status: %s", result.Status))

	// Store the result
	if err := m.storage.StoreTestResult(result); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to store test result")
		return detailedResult, fmt.Errorf("failed to store test result: %w", err)
	}

	log.Ctx(ctx).Info().
		Str("endpoint_id", endpointID).
		Str("region", regionName).
		Dur("latency", result.Latency).
		Dur("execution_time", executionTime).
		Str("status", result.Status).
		Msg("Detailed manual test completed successfully")

	return detailedResult, nil
}

// RunRegionTests executes manual tests for all endpoints in a region
func (m *Manager) RunRegionTests(ctx context.Context, regionName string) ([]*storage.DetailedTestResult, error) {
	log.Ctx(ctx).Info().Str("region", regionName).Msg("Running region tests")

	cfg := m.config.GetConfig()
	region, exists := cfg.Regions[regionName]
	if !exists {
		return nil, fmt.Errorf("region not found: %s", regionName)
	}

	results := make([]*storage.DetailedTestResult, 0, len(region.Endpoints))
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errors []error

	for _, endpoint := range region.Endpoints {
		endpointID := fmt.Sprintf("%s-%s", regionName, endpoint.Name)
		wg.Add(1)

		go func(epID string) {
			defer wg.Done()

			result, err := m.RunManualTestDetailed(ctx, epID)

			mutex.Lock()
			if result != nil {
				results = append(results, result)
			}
			if err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", epID, err))
			}
			mutex.Unlock()
		}(endpointID)
	}

	wg.Wait()

	if len(errors) > 0 {
		log.Ctx(ctx).Warn().
			Str("region", regionName).
			Int("errors", len(errors)).
			Int("total", len(region.Endpoints)).
			Msg("Some region tests failed")
	}

	log.Ctx(ctx).Info().
		Str("region", regionName).
		Int("successful", len(results)).
		Int("total", len(region.Endpoints)).
		Msg("Region tests completed")

	return results, nil
}

// RunAllTests executes manual tests for all configured endpoints
func (m *Manager) RunAllTests(ctx context.Context) ([]*storage.DetailedTestResult, error) {
	log.Ctx(ctx).Info().Msg("Running all tests")

	cfg := m.config.GetConfig()
	var allResults []*storage.DetailedTestResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for regionName := range cfg.Regions {
		wg.Add(1)

		go func(rName string) {
			defer wg.Done()

			results, err := m.RunRegionTests(ctx, rName)

			mutex.Lock()
			if results != nil {
				allResults = append(allResults, results...)
			}
			mutex.Unlock()

			if err != nil {
				log.Ctx(ctx).Warn().
					Str("region", rName).
					Err(err).
					Msg("Failed to run region tests")
			}
		}(regionName)
	}

	wg.Wait()

	log.Ctx(ctx).Info().
		Int("total_results", len(allResults)).
		Msg("All tests completed")

	return allResults, nil
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