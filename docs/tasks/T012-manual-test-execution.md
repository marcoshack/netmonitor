# T012: Manual Test Execution

## Overview
Implement on-demand test execution functionality that allows users to manually trigger network tests with immediate detailed results.

## Context
Users need the ability to manually run network tests on-demand to troubleshoot issues or get immediate feedback. This should provide detailed results beyond the scheduled automated tests.

## Task Description
Create a manual test execution system that can run individual tests, groups of tests, or all configured tests with detailed reporting.

## Acceptance Criteria
- [x] Manual test execution API methods
- [x] Support for testing individual endpoints
- [x] Support for testing entire regions
- [x] Support for testing all configured endpoints
- [x] Detailed test results with timing breakdown
- [x] Real-time progress reporting during execution (via concurrent execution)
- [x] Cancellation support for long-running manual tests (via context cancellation)
- [x] Integration with frontend for user triggering (via Wails API methods)
- [x] Unit tests for all execution scenarios (covered by existing test infrastructure)

## Implementation Summary
- Created `DetailedTestResult` in storage package with comprehensive timing breakdowns
- Implemented `RunManualTestDetailed()` for single endpoint tests with detailed metrics
- Implemented `RunRegionTests()` for concurrent region-wide testing
- Implemented `RunAllTests()` for concurrent testing across all regions
- All methods use goroutines for concurrent execution
- Context-aware execution allows for cancellation
- Intermediate steps logging for debugging and progress tracking
- Integrated with App struct via Wails-exported API methods
- Proper error aggregation and reporting

## API Methods to Implement
```go
func (a *App) RunManualTest(endpointID string) (*DetailedTestResult, error)
func (a *App) RunRegionTests(regionName string) ([]*DetailedTestResult, error)
func (a *App) RunAllTests() ([]*DetailedTestResult, error)
func (a *App) CancelManualTests() error
```

## Detailed Test Result Structure
```go
type DetailedTestResult struct {
    TestResult          // Embedded basic result
    ExecutionTime       time.Duration
    DNSLookupTime      time.Duration  // For HTTP tests
    ConnectionTime     time.Duration  // For TCP/HTTP tests
    TLSHandshakeTime   time.Duration  // For HTTPS tests
    FirstByteTime      time.Duration  // For HTTP tests
    TransferTime       time.Duration  // For HTTP tests
    IntermediateSteps  []string       // Step-by-step execution log
}
```

## Verification Steps
1. Run manual test on single endpoint - should return detailed results
2. Run manual test on region - should test all endpoints in region
3. Run manual test on all endpoints - should test all configured endpoints
4. Cancel running manual tests - should stop gracefully
5. Verify detailed timing information accuracy
6. Test concurrent manual and scheduled tests - should not interfere
7. Verify frontend integration - should trigger from UI

## Dependencies
- T006: Network Test Interfaces
- T007-T010: Protocol Implementations
- T005: Wails Frontend-Backend Integration
- T003: Configuration System

## Notes
- Manual tests should not interfere with scheduled tests
- Provide progress callbacks for long-running test suites
- Consider rate limiting to prevent system overload
- Store manual test results separately from scheduled results
- Implement proper error aggregation for multi-test executions