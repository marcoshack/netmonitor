# T010: UDP Test Implementation

## Overview
Implement UDP connectivity testing with packet sending and response validation for UDP-based services.

## Context
UDP testing is important for monitoring DNS servers, game servers, and other UDP-based services. Unlike TCP, UDP is connectionless, so testing requires sending packets and waiting for responses.

## Task Description
Create a UDP test implementation that can send packets to UDP services and measure response time or detect service availability.

## Acceptance Criteria
- [x] UDP test implementation satisfying NetworkTest interface
- [x] UDP packet sending to specified host:port
- [x] Response time measurement for services that respond
- [x] Timeout handling for non-responsive services
- [x] Support for IPv4 and IPv6 addresses
- [x] Custom payload support for different UDP services
- [x] Port validation (1-65535)
- [x] Unit tests with mock UDP server
- [x] Integration tests with real UDP services (e.g., DNS)

## Implementation Requirements
- Use Go's `net.DialUDP` or `net.UDPConn`
- Send test packets and measure response time
- Handle services that don't respond (like NTP, SNMP)
- Support custom payloads for specific protocols
- Proper connection cleanup

## Example Usage
```go
udpTest := &UDPTest{}
config := TestConfig{
    Name:     "DNS Server",
    Address:  "8.8.8.8:53",
    Timeout:  3 * time.Second,
    Protocol: "UDP",
    Config: &UDPConfig{
        Payload: dnsQueryPacket, // Custom DNS query
    },
}
result, err := udpTest.Execute(ctx, config)
```

## UDP Service Types to Support
- **DNS queries** - Send DNS request, expect response
- **NTP servers** - Send NTP request, measure response
- **Echo services** - Send data, expect echo back
- **Generic UDP** - Send packet, measure if response received

## Verification Steps
1. Test DNS server (8.8.8.8:53) with DNS query - should get response
2. Test non-responsive UDP port - should timeout appropriately
3. Test invalid port number - should fail validation
4. Test timeout behavior - should respect timeout setting
5. Test custom payload - should send correct data
6. Verify response time measurement accuracy
7. Test concurrent UDP tests - should handle simultaneous tests

## Dependencies
- T006: Network Test Interfaces

## Notes
- UDP is connectionless, so "success" means response received
- Some UDP services may not respond to generic packets
- Consider implementing protocol-specific payloads (DNS, NTP)
- Handle ICMP port unreachable responses
- Be careful with UDP flood protection on target systems

## Implementation Summary

**Completed**: All acceptance criteria met

### Files Created
- `internal/network/udp.go` - UDPTest implementation
- `internal/network/udp_test.go` - Unit tests (13 test cases)
- `internal/network/udp_integration_test.go` - Integration tests (9 test cases)

### Key Features Implemented
1. **NetworkTest Interface**: Implements `Execute()`, `GetProtocol()`, and `Validate()` methods
2. **Address Validation**: Port range (1-65535), host:port format, IPv4/IPv6 support
3. **UDP Operations**:
   - Send UDP packets with custom or default payload
   - Optional response waiting with configurable buffer size
   - Context cancellation support
   - Proper timeout handling
4. **Error Handling**: Timeout detection, port unreachable (ICMP), network errors
5. **Configuration**: UDPConfig with SendData, WaitResponse, ResponseSize fields

### Test Coverage
- **Unit Tests**: Validation, successful send/receive, timeouts, context cancellation, IPv6, error cases
- **Integration Tests**: Real DNS queries (Google DNS, Cloudflare), IPv6, latency measurement
- All 22 test cases passing

### Example Usage
```go
// DNS query test
udpTest := &UDPTest{}
config := TestConfig{
    Name:    "Google DNS",
    Address: "8.8.8.8:53",
    Timeout: 3 * time.Second,
    Config: &UDPConfig{
        SendData:     dnsQueryPacket,
        WaitResponse: true,
        ResponseSize: 512,
    },
}
result, err := udpTest.Execute(ctx, config)
```