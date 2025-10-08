package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marcoshack/netmonitor/internal/config"
	"github.com/marcoshack/netmonitor/internal/network"
	"github.com/marcoshack/netmonitor/internal/storage"
)

// Scheduler manages periodic execution of network tests
type Scheduler interface {
	Start(ctx context.Context) error
	Stop() error
	UpdateInterval(interval time.Duration) error
	GetStatus() SchedulerStatus
}

// SchedulerStatus represents the current state of the scheduler
type SchedulerStatus struct {
	Running        bool          `json:"running"`
	Interval       time.Duration `json:"interval"`
	LastRun        time.Time     `json:"lastRun"`
	NextRun        time.Time     `json:"nextRun"`
	ActiveTests    int           `json:"activeTests"`
	CompletedTests int64         `json:"completedTests"`
}

// TestScheduler implements the Scheduler interface
type TestScheduler struct {
	config          *config.Manager
	storage         *storage.Manager
	ctx             context.Context
	cancel          context.CancelFunc
	interval        time.Duration
	ticker          *time.Ticker
	running         bool
	mutex           sync.RWMutex
	activeTests     int
	completedTests  int64
	lastRun         time.Time
	maxConcurrent   int
	semaphore       chan struct{}
	testExecutor    TestExecutor
}

// TestExecutor executes network tests
type TestExecutor interface {
	ExecuteTest(ctx context.Context, endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error)
}

// DefaultTestExecutor implements TestExecutor
type DefaultTestExecutor struct {
	storage *storage.Manager
}

// NewScheduler creates a new test scheduler
func NewScheduler(ctx context.Context, configMgr *config.Manager, storageMgr *storage.Manager, maxConcurrent int) (*TestScheduler, error) {
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // Default to 10 concurrent tests
	}

	cfg := configMgr.GetConfig()
	interval := time.Duration(cfg.Settings.TestIntervalSeconds) * time.Second

	scheduler := &TestScheduler{
		config:        configMgr,
		storage:       storageMgr,
		interval:      interval,
		running:       false,
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
		testExecutor:  &DefaultTestExecutor{storage: storageMgr},
	}

	// Add configuration change callback
	configMgr.AddCallback(func(newCfg *config.Config) {
		newInterval := time.Duration(newCfg.Settings.TestIntervalSeconds) * time.Second
		if err := scheduler.UpdateInterval(newInterval); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Failed to update scheduler interval")
		}
	})

	return scheduler, nil
}

// Start begins the test scheduling
func (s *TestScheduler) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.ticker = time.NewTicker(s.interval)

	log.Ctx(ctx).Info().Dur("interval", s.interval).Msg("Test scheduler starting")

	// Start the scheduling loop
	go s.schedulingLoop()

	return nil
}

// Stop stops the test scheduling
func (s *TestScheduler) Stop() error {
	s.mutex.Lock()

	if !s.running {
		s.mutex.Unlock()
		return nil
	}

	log.Ctx(s.ctx).Info().Msg("Test scheduler stopping")

	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.mutex.Unlock()

	// Wait for active tests to complete (with timeout)
	timeout := time.After(30 * time.Second)
	for {
		s.mutex.RLock()
		active := s.activeTests
		s.mutex.RUnlock()

		if active == 0 {
			break
		}

		select {
		case <-timeout:
			log.Ctx(s.ctx).Warn().Int("active_tests", active).Msg("Timeout waiting for active tests to complete")
			return nil
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}

	log.Ctx(s.ctx).Info().Msg("Test scheduler stopped")
	return nil
}

// UpdateInterval updates the test interval
func (s *TestScheduler) UpdateInterval(interval time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if interval < time.Minute || interval > 24*time.Hour {
		return fmt.Errorf("interval must be between 1 minute and 24 hours")
	}

	oldInterval := s.interval
	s.interval = interval

	// If running, restart ticker with new interval
	if s.running && s.ticker != nil {
		s.ticker.Stop()
		s.ticker = time.NewTicker(interval)
		log.Ctx(s.ctx).Info().
			Dur("old_interval", oldInterval).
			Dur("new_interval", interval).
			Msg("Scheduler interval updated")
	}

	return nil
}

// GetStatus returns the current scheduler status
func (s *TestScheduler) GetStatus() SchedulerStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nextRun := s.lastRun.Add(s.interval)
	if s.lastRun.IsZero() && s.running {
		nextRun = time.Now()
	}

	return SchedulerStatus{
		Running:        s.running,
		Interval:       s.interval,
		LastRun:        s.lastRun,
		NextRun:        nextRun,
		ActiveTests:    s.activeTests,
		CompletedTests: s.completedTests,
	}
}

// schedulingLoop runs the main scheduling loop
func (s *TestScheduler) schedulingLoop() {
	// Run tests immediately on start
	s.runScheduledTests()

	for {
		select {
		case <-s.ticker.C:
			s.runScheduledTests()
		case <-s.ctx.Done():
			return
		}
	}
}

// runScheduledTests executes tests for all configured endpoints
func (s *TestScheduler) runScheduledTests() {
	s.mutex.Lock()
	s.lastRun = time.Now()
	s.mutex.Unlock()

	cfg := s.config.GetConfig()
	var wg sync.WaitGroup

	log.Ctx(s.ctx).Info().Msg("Running scheduled tests")

	for regionName, region := range cfg.Regions {
		for _, endpoint := range region.Endpoints {
			endpointID := fmt.Sprintf("%s-%s", regionName, endpoint.Name)

			// Acquire semaphore
			s.semaphore <- struct{}{}
			wg.Add(1)

			// Increment active tests
			s.mutex.Lock()
			s.activeTests++
			s.mutex.Unlock()

			// Run test in goroutine
			go func(ep *config.Endpoint, epID, region string) {
				defer wg.Done()
				defer func() {
					<-s.semaphore // Release semaphore
					s.mutex.Lock()
					s.activeTests--
					s.completedTests++
					s.mutex.Unlock()
				}()

				result, err := s.testExecutor.ExecuteTest(s.ctx, ep, epID)
				if err != nil {
					log.Ctx(s.ctx).Error().
						Str("endpoint_id", epID).
						Str("region", region).
						Err(err).
						Msg("Scheduled test execution failed")
				} else {
					log.Ctx(s.ctx).Debug().
						Str("endpoint_id", epID).
						Dur("latency", result.Latency).
						Str("status", result.Status).
						Msg("Scheduled test completed")
				}
			}(endpoint, endpointID, regionName)
		}
	}

	// Wait for all tests in this batch to complete
	wg.Wait()
	log.Ctx(s.ctx).Debug().Msg("Scheduled test batch completed")
}

// ExecuteTest implements TestExecutor interface
func (e *DefaultTestExecutor) ExecuteTest(ctx context.Context, endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error) {
	// Create a context with timeout
	testCtx, cancel := context.WithTimeout(ctx, time.Duration(endpoint.Timeout)*time.Millisecond)
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
	case "HTTP":
		networkTest = &network.HTTPTest{}
		testConfig = network.TestConfig{
			Name:     endpoint.Name,
			Address:  endpoint.Address,
			Timeout:  time.Duration(endpoint.Timeout) * time.Millisecond,
			Protocol: "HTTP",
			Config: &network.HTTPConfig{
				Method:          "GET",
				Headers:         make(map[string]string),
				FollowRedirects: true,
				ValidateSSL:     true,
				ExpectedStatus:  0,
			},
		}
	case "TCP":
		networkTest = &network.TCPTest{}
		testConfig = network.TestConfig{
			Name:     endpoint.Name,
			Address:  endpoint.Address,
			Timeout:  time.Duration(endpoint.Timeout) * time.Millisecond,
			Protocol: "TCP",
			Config: &network.TCPConfig{
				ExpectResponse: false,
			},
		}
	case "UDP":
		networkTest = &network.UDPTest{}
		testConfig = network.TestConfig{
			Name:     endpoint.Name,
			Address:  endpoint.Address,
			Timeout:  time.Duration(endpoint.Timeout) * time.Millisecond,
			Protocol: "UDP",
			Config: &network.UDPConfig{
				SendData:     "PING",
				WaitResponse: true,
				ResponseSize: 1024,
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
	networkResult, err := networkTest.Execute(testCtx, testConfig)
	if err != nil {
		// Return a failed test result even on error
		failedResult := &storage.TestResult{
			Timestamp:  time.Now(),
			EndpointID: endpointID,
			Protocol:   endpoint.Type,
			Latency:    0,
			Status:     string(network.TestStatusFailed),
			Error:      err.Error(),
		}
		// Still store failed results
		if storeErr := e.storage.StoreTestResult(failedResult); storeErr != nil {
			log.Ctx(ctx).Error().Err(storeErr).Msg("Failed to store failed test result")
		}
		return failedResult, err
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

	// Store the result
	if err := e.storage.StoreTestResult(storageResult); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to store test result")
		return storageResult, fmt.Errorf("failed to store test result: %w", err)
	}

	return storageResult, nil
}
