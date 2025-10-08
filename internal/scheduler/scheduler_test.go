package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/marcoshack/netmonitor/internal/config"
	"github.com/marcoshack/netmonitor/internal/storage"
)

// MockTestExecutor is a mock implementation of TestExecutor for testing
type MockTestExecutor struct {
	executeFunc func(ctx context.Context, endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error)
	callCount   int
	mutex       sync.Mutex
}

func (m *MockTestExecutor) ExecuteTest(ctx context.Context, endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error) {
	m.mutex.Lock()
	m.callCount++
	m.mutex.Unlock()

	if m.executeFunc != nil {
		return m.executeFunc(ctx, endpoint, endpointID)
	}

	return &storage.TestResult{
		Timestamp:  time.Now(),
		EndpointID: endpointID,
		Protocol:   endpoint.Type,
		Latency:    10 * time.Millisecond,
		Status:     "success",
	}, nil
}

func (m *MockTestExecutor) GetCallCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.callCount
}

// setupTestScheduler creates a scheduler for testing
func setupTestScheduler(t *testing.T) (*TestScheduler, *config.Manager, *storage.Manager, context.Context) {
	ctx := context.Background()

	// Create temporary directory for config
	configMgr, err := config.NewManager(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create temporary directory for storage
	storageMgr, err := storage.NewManager(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	scheduler, err := NewScheduler(ctx, configMgr, storageMgr, 5)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	return scheduler, configMgr, storageMgr, ctx
}

func TestNewScheduler(t *testing.T) {
	scheduler, _, _, _ := setupTestScheduler(t)

	if scheduler == nil {
		t.Fatal("Expected scheduler to be created")
	}

	if scheduler.maxConcurrent != 5 {
		t.Errorf("Expected maxConcurrent to be 5, got %d", scheduler.maxConcurrent)
	}

	if scheduler.running {
		t.Error("Expected scheduler to not be running initially")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	scheduler, _, _, ctx := setupTestScheduler(t)

	// Test starting scheduler
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	status := scheduler.GetStatus()
	if !status.Running {
		t.Error("Expected scheduler to be running after Start()")
	}

	// Test starting already running scheduler
	err = scheduler.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running scheduler")
	}

	// Test stopping scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}

	status = scheduler.GetStatus()
	if status.Running {
		t.Error("Expected scheduler to not be running after Stop()")
	}

	// Test stopping already stopped scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Error("Expected no error when stopping already stopped scheduler")
	}
}

func TestSchedulerUpdateInterval(t *testing.T) {
	scheduler, _, _, _ := setupTestScheduler(t)

	tests := []struct {
		name        string
		interval    time.Duration
		expectError bool
	}{
		{"Valid 1 minute", 1 * time.Minute, false},
		{"Valid 5 minutes", 5 * time.Minute, false},
		{"Valid 1 hour", 1 * time.Hour, false},
		{"Valid 24 hours", 24 * time.Hour, false},
		{"Invalid too short", 30 * time.Second, true},
		{"Invalid too long", 25 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scheduler.UpdateInterval(tt.interval)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError {
				status := scheduler.GetStatus()
				if status.Interval != tt.interval {
					t.Errorf("Expected interval %v, got %v", tt.interval, status.Interval)
				}
			}
		})
	}
}

func TestSchedulerGetStatus(t *testing.T) {
	scheduler, _, _, ctx := setupTestScheduler(t)

	// Get status before starting
	status := scheduler.GetStatus()
	if status.Running {
		t.Error("Expected Running to be false before starting")
	}
	if status.ActiveTests != 0 {
		t.Error("Expected ActiveTests to be 0")
	}
	if status.CompletedTests != 0 {
		t.Error("Expected CompletedTests to be 0")
	}

	// Start scheduler and check status
	scheduler.Start(ctx)
	defer scheduler.Stop()

	status = scheduler.GetStatus()
	if !status.Running {
		t.Error("Expected Running to be true after starting")
	}
}

func TestSchedulerConcurrentExecution(t *testing.T) {
	scheduler, _, _, ctx := setupTestScheduler(t)

	// Use mock executor to track calls
	mockExecutor := &MockTestExecutor{
		executeFunc: func(ctx context.Context, endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error) {
			// Simulate some work
			time.Sleep(50 * time.Millisecond)
			return &storage.TestResult{
				Timestamp:  time.Now(),
				EndpointID: endpointID,
				Protocol:   endpoint.Type,
				Latency:    10 * time.Millisecond,
				Status:     "success",
			}, nil
		},
	}
	scheduler.testExecutor = mockExecutor

	// Start scheduler
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// Wait for tests to execute
	time.Sleep(200 * time.Millisecond)

	// Check that tests were executed
	callCount := mockExecutor.GetCallCount()
	if callCount == 0 {
		t.Error("Expected at least one test to be executed")
	}

	status := scheduler.GetStatus()
	if status.CompletedTests == 0 {
		t.Error("Expected completed tests count to be greater than 0")
	}
}

func TestSchedulerConfigurationReload(t *testing.T) {
	scheduler, configMgr, _, _ := setupTestScheduler(t)

	initialInterval := scheduler.interval

	// Update configuration with new interval
	cfg := configMgr.GetConfig()
	cfg.Settings.TestIntervalSeconds = 120
	err := configMgr.UpdateConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Give callback time to execute
	time.Sleep(100 * time.Millisecond)

	// Check that interval was updated
	newInterval := scheduler.interval
	if newInterval == initialInterval {
		t.Error("Expected interval to be updated after configuration change")
	}

	expectedInterval := 120 * time.Second
	if newInterval != expectedInterval {
		t.Errorf("Expected interval %v, got %v", expectedInterval, newInterval)
	}
}

func TestSchedulerGracefulShutdown(t *testing.T) {
	scheduler, _, _, ctx := setupTestScheduler(t)

	// Use mock executor with longer execution time
	mockExecutor := &MockTestExecutor{
		executeFunc: func(ctx context.Context, endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error) {
			// Simulate long-running test
			select {
			case <-time.After(500 * time.Millisecond):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			return &storage.TestResult{
				Timestamp:  time.Now(),
				EndpointID: endpointID,
				Protocol:   endpoint.Type,
				Latency:    10 * time.Millisecond,
				Status:     "success",
			}, nil
		},
	}
	scheduler.testExecutor = mockExecutor

	// Start scheduler
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for tests to start
	time.Sleep(100 * time.Millisecond)

	// Stop scheduler (should wait for active tests)
	stopStart := time.Now()
	err = scheduler.Stop()
	stopDuration := time.Since(stopStart)

	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}

	// Should have waited for tests to complete (but less than timeout)
	if stopDuration < 100*time.Millisecond {
		t.Error("Stop() returned too quickly, may not have waited for active tests")
	}
	if stopDuration > 30*time.Second {
		t.Error("Stop() took too long, may have exceeded timeout")
	}
}

func TestSchedulerMaxConcurrent(t *testing.T) {
	testCtx := context.Background()

	configMgr, err := config.NewManager(testCtx, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	storageMgr, err := storage.NewManager(testCtx, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create scheduler with max 2 concurrent tests
	scheduler, err := NewScheduler(testCtx, configMgr, storageMgr, 2)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	var maxActive int
	var mutex sync.Mutex

	mockExecutor := &MockTestExecutor{
		executeFunc: func(ctx context.Context, endpoint *config.Endpoint, endpointID string) (*storage.TestResult, error) {
			scheduler.mutex.RLock()
			active := scheduler.activeTests
			scheduler.mutex.RUnlock()

			mutex.Lock()
			if active > maxActive {
				maxActive = active
			}
			mutex.Unlock()

			time.Sleep(100 * time.Millisecond)
			return &storage.TestResult{
				Timestamp:  time.Now(),
				EndpointID: endpointID,
				Protocol:   endpoint.Type,
				Latency:    10 * time.Millisecond,
				Status:     "success",
			}, nil
		},
	}
	scheduler.testExecutor = mockExecutor

	scheduler.Start(testCtx)
	time.Sleep(300 * time.Millisecond)
	scheduler.Stop()

	mutex.Lock()
	finalMaxActive := maxActive
	mutex.Unlock()

	// Max concurrent should not exceed 2
	if finalMaxActive > 2 {
		t.Errorf("Expected max active tests to be <= 2, got %d", finalMaxActive)
	}
}
