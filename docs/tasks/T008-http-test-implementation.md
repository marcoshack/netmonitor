# T008: HTTP Test Implementation

## Overview
Implement HTTP/HTTPS connectivity testing with support for GET requests and response time measurement.

## Context
HTTP testing allows monitoring of web services and endpoints. This is crucial for testing web-based services and measuring their response times and availability.

## Task Description
Create a comprehensive HTTP test implementation that can perform GET requests to HTTP and HTTPS endpoints and measure response characteristics.

## Acceptance Criteria
- [ ] HTTP test implementation satisfying NetworkTest interface
- [ ] Support for both HTTP and HTTPS protocols
- [ ] GET request execution with configurable timeout
- [ ] Response time measurement (DNS lookup, connection, response)
- [ ] HTTP status code handling and validation
- [ ] Response size measurement
- [ ] Custom User-Agent header
- [ ] TLS/SSL certificate validation
- [ ] Redirect handling (configurable)
- [ ] Unit tests and integration tests

## Implementation Requirements
- Use Go's standard `net/http` package
- Measure different phases of the request (DNS, connect, response)
- Handle various HTTP status codes appropriately
- Support custom headers and timeouts
- Validate SSL certificates properly
- Handle network errors gracefully

## Example Usage
```go
httpTest := &HTTPTest{}
config := TestConfig{
    Name:     "Cloudflare HTTP",
    Address:  "https://1.1.1.1",
    Timeout:  10 * time.Second,
    Protocol: "HTTP",
}
result, err := httpTest.Execute(ctx, config)
```

## Verification Steps
1. Test valid HTTP endpoint - should return success with response time
2. Test valid HTTPS endpoint - should return success with SSL validation
3. Test non-existent endpoint - should return connection error
4. Test timeout with slow endpoint - should respect timeout
5. Test various HTTP status codes (200, 404, 500) - should handle appropriately
6. Test invalid SSL certificate - should fail validation
7. Verify response size measurement accuracy

## Dependencies
- T006: Network Test Interfaces

## Notes
- Consider using `httptrace` package for detailed timing
- Handle redirects according to configuration
- Implement proper connection pooling and cleanup
- Support for HTTP/2 where available
- Consider implementing HEAD requests for bandwidth efficiency