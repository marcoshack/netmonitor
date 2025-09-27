# T011: Test Scheduler

## Overview
Implement a scheduler system that executes network tests at configurable intervals according to the configuration settings.

## Context
NetMonitor needs to automatically run network tests at regular intervals (configurable from 1 minute to 24 hours, default 5 minutes). The scheduler must handle multiple endpoints efficiently and manage concurrent test execution.

## Task Description
Create a robust test scheduler that manages periodic execution of network tests across all configured endpoints and regions.

## Acceptance Criteria
- [ ] Scheduler starts/stops with application lifecycle
- [ ] Configurable test intervals (1 minute to 24 hours)
- [ ] Concurrent test execution with configurable limits
- [ ] Graceful handling of long-running tests
- [ ] Test result collection and forwarding
- [ ] Error handling and retry logic
- [ ] Configuration reload without restart
- [ ] Scheduler status reporting
- [ ] Unit tests with time mocking

## Implementation Requirements
- Use Go's `time.Ticker` or similar for scheduling
- Implement worker pool pattern for concurrent tests
- Handle context cancellation for graceful shutdown
- Support dynamic configuration changes
- Collect and aggregate test results

## Scheduler Interface
```go
type Scheduler interface {
    Start(ctx context.Context) error
    Stop() error
    UpdateInterval(interval time.Duration) error
    GetStatus() SchedulerStatus
}

type SchedulerStatus struct {
    Running        bool
    Interval       time.Duration
    LastRun        time.Time
    NextRun        time.Time
    ActiveTests    int
    CompletedTests int64
}
```

## Verification Steps
1. Start scheduler with 1-minute interval - should execute tests every minute
2. Update interval to 5 minutes - should reschedule appropriately
3. Stop scheduler gracefully - should complete running tests and stop
4. Test with long-running tests - should handle timeouts properly
5. Verify concurrent test execution - should run multiple tests simultaneously
6. Test configuration reload - should pick up new endpoints
7. Verify scheduler status reporting accuracy

## Dependencies
- T006: Network Test Interfaces
- T007: ICMP Ping Implementation
- T008: HTTP Test Implementation
- T009: TCP Test Implementation
- T010: UDP Test Implementation
- T003: Configuration System

## Notes
- Consider using a worker pool pattern for concurrent execution
- Implement proper cleanup on shutdown
- Log scheduler events for debugging
- Consider jitter to avoid thundering herd effects
- Handle system clock changes gracefully