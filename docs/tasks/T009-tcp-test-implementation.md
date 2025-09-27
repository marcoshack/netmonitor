# T009: TCP Test Implementation

## Overview
Implement TCP connection testing to verify port connectivity and measure connection establishment time.

## Context
TCP connection tests are essential for monitoring service availability on specific ports. This allows testing database connections, service ports, and other TCP-based services.

## Task Description
Create a TCP connection test implementation that can establish connections to TCP ports and measure connection time.

## Acceptance Criteria
- [ ] TCP test implementation satisfying NetworkTest interface
- [ ] TCP connection establishment to specified host:port
- [ ] Connection time measurement
- [ ] Support for IPv4 and IPv6 addresses
- [ ] Configurable connection timeout
- [ ] Proper connection cleanup (close)
- [ ] Port validation (1-65535)
- [ ] Unit tests with mock TCP server
- [ ] Integration tests with real services

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