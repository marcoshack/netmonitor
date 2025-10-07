# T009: TCP Test Implementation

## Overview
Implement TCP connection testing to verify port connectivity and measure connection establishment time.

## Context
TCP connection tests are essential for monitoring service availability on specific ports. This allows testing database connections, service ports, and other TCP-based services.

## Task Description
Create a TCP connection test implementation that can establish connections to TCP ports and measure connection time.

## Acceptance Criteria
- [x] TCP test implementation satisfying NetworkTest interface
- [x] TCP connection establishment to specified host:port
- [x] Connection time measurement
- [x] Support for IPv4 and IPv6 addresses
- [x] Configurable connection timeout
- [x] Proper connection cleanup (close)
- [x] Port validation (1-65535)
- [x] Unit tests with mock TCP server
- [x] Integration tests with real services

## Implementation Requirements
- Use Go's `net.Dial` or `net.DialTimeout` functions
- Measure time from connection start to establishment
- Handle connection refused, timeout, and other errors
- Validate port numbers and addresses
- Clean up connections properly

## Example Usage
```go
tcpTest := &TCPTest{}
config := TestConfig{
    Name:     "Database Connection",
    Address:  "db.example.com:5432",
    Timeout:  5 * time.Second,
    Protocol: "TCP",
}
result, err := tcpTest.Execute(ctx, config)
```

## Test Scenarios
- Open port (should connect successfully)
- Closed port (should fail with connection refused)
- Filtered port (should timeout)
- Invalid address format
- Invalid port number
- Network unreachable

## Verification Steps
1. Connect to open port (e.g., 80, 443) - should succeed
2. Connect to closed port - should fail with appropriate error
3. Test connection timeout - should respect timeout setting
4. Test invalid port number - should fail validation
5. Test invalid address format - should fail validation
6. Verify connection cleanup - should not leak connections
7. Test concurrent connections - should handle multiple simultaneous tests

## Dependencies
- T006: Network Test Interfaces

## Notes
- Consider testing common ports (22, 80, 443, 3306, 5432)
- Handle different error types appropriately
- Implement connection pooling if needed for performance
- Consider adding basic banner grabbing for service identification

## Implementation Summary

### Files Created
- `internal/network/tcp.go` - TCP test implementation
- `internal/network/tcp_test.go` - Unit tests with mock TCP server
- `internal/network/tcp_integration_test.go` - Integration tests
- `cmd/netmonitor-tcp-example/main.go` - Example usage program

### Key Features Implemented
- Full NetworkTest interface implementation (Execute, GetProtocol, Validate)
- Connection establishment with configurable timeout using `net.DialContext`
- Precise connection time measurement
- IPv4 and IPv6 support
- Port validation (1-65535)
- Proper connection cleanup with defer
- Optional data sending and response validation via TCPConfig
- Comprehensive error handling for:
  - Connection refused (Windows and Unix)
  - Connection timeout
  - Network unreachable
  - Invalid address/port formats
  - Context cancellation

### Test Coverage
- **Unit Tests**: 15 test cases covering validation, successful connections, error conditions, data exchange, IPv6, and concurrent connections
- **Integration Tests**: 7 test suites covering public services, common ports, HTTP over TCP, IPv6, concurrent connections, database ports, and timeout behavior
- All tests passing on Windows

### Example Usage
The example program demonstrates:
1. Simple TCP connection tests (DNS, HTTPS ports)
2. HTTP request over raw TCP with response validation
3. Database port scanning (PostgreSQL, MySQL, Redis, MongoDB)
4. Closed port handling
5. IPv6 connection testing

### Status
✅ **COMPLETED** - All acceptance criteria met and verified