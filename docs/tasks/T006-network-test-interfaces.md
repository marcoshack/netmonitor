# T006: Network Test Interfaces

## Overview
Define the core interfaces and data structures for network monitoring tests supporting HTTP, TCP, UDP, and ICMP protocols.

## Context
NetMonitor needs to support multiple network protocols for testing connectivity and measuring latency. Each protocol requires specific implementation but should share common interfaces for unified handling.

## Task Description
Create the foundational interfaces and data structures that will be used by all network testing implementations.

## Acceptance Criteria
- [ ] `NetworkTest` interface defined with common methods
- [ ] `TestResult` struct for storing test outcomes
- [ ] `TestConfig` struct for test parameters
- [ ] Protocol-specific configuration structures:
  - `HTTPConfig` for HTTP/HTTPS tests
  - `TCPConfig` for TCP connection tests
  - `UDPConfig` for UDP tests
  - `ICMPConfig` for ping tests
- [ ] Test timeout and error handling interfaces
- [ ] Unit tests for data structures

## Core Interfaces
```go
type NetworkTest interface {
    Execute(ctx context.Context, config TestConfig) (*TestResult, error)
    GetProtocol() string
    Validate(config TestConfig) error
}

type TestResult struct {
    Timestamp    time.Time
    EndpointID   string
    Protocol     string
    Latency      time.Duration
    Status       TestStatus
    Error        string
    ResponseSize int64
}

type TestConfig struct {
    Name     string
    Address  string
    Timeout  time.Duration
    Protocol string
    Config   interface{} // Protocol-specific config
}
```

## Verification Steps
1. Create test instances of each config type - should compile
2. Implement a mock NetworkTest - should satisfy interface
3. Create TestResult instances - should handle all fields correctly
4. Test timeout handling - should respect context cancellation
5. Verify error handling for invalid configurations

## Dependencies
- T002: Basic Application Structure

## Notes
- Use Go interfaces for testability and extensibility
- Consider using context.Context for cancellation
- Design for concurrent execution of multiple tests
- Prepare for metrics collection and aggregation