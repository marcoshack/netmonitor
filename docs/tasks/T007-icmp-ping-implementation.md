# T007: ICMP Ping Implementation

## Overview
Implement ICMP ping functionality for testing network connectivity and measuring round-trip time to endpoints.

## Context
ICMP ping is a fundamental network diagnostic tool that sends packets to a destination and measures response time. This will be one of the core monitoring protocols in NetMonitor.

## Task Description
Create a complete ICMP ping implementation that can test connectivity to IP addresses and measure latency accurately.

## Acceptance Criteria
- [ ] ICMP ping implementation satisfying NetworkTest interface
- [ ] Support for both IPv4 and IPv6 addresses
- [ ] Accurate latency measurement in milliseconds
- [ ] Proper handling of unreachable hosts
- [ ] Timeout handling and cancellation support
- [ ] Cross-platform compatibility (Windows, macOS, Linux)
- [ ] Unit tests with mock scenarios
- [ ] Integration tests with real endpoints

## Implementation Requirements
- Use raw sockets or appropriate OS-specific libraries
- Handle ICMP echo request/reply packets
- Calculate round-trip time accurately
- Support configurable timeout values
- Handle network permission requirements

## Example Usage
```go
icmpTest := &ICMPTest{}
config := TestConfig{
    Name:     "Google DNS",
    Address:  "8.8.8.8",
    Timeout:  5 * time.Second,
    Protocol: "ICMP",
}
result, err := icmpTest.Execute(ctx, config)
```

## Verification Steps
1. Ping reachable host (e.g., 8.8.8.8) - should return success with latency
2. Ping unreachable host - should return failure with timeout
3. Test with invalid IP address - should return validation error
4. Test timeout cancellation - should respect context deadline
5. Test concurrent pings - should handle multiple simultaneous tests
6. Verify cross-platform functionality

## Dependencies
- T006: Network Test Interfaces

## Notes
- May require elevated privileges on some systems
- Consider using a Go ICMP library (e.g., golang.org/x/net/icmp)
- Handle platform-specific socket requirements
- Implement proper cleanup of network resources