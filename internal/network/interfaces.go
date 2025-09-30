package network

import (
	"context"
	"time"
)

// TestStatus represents the status of a network test
type TestStatus string

const (
	TestStatusSuccess TestStatus = "success"
	TestStatusFailed  TestStatus = "failed"
	TestStatusTimeout TestStatus = "timeout"
	TestStatusError   TestStatus = "error"
)

// NetworkTest interface defines the contract for network testing implementations
type NetworkTest interface {
	// Execute runs the network test with the given configuration
	Execute(ctx context.Context, config TestConfig) (*TestResult, error)

	// GetProtocol returns the protocol type this test implements
	GetProtocol() string

	// Validate checks if the configuration is valid for this test type
	Validate(config TestConfig) error
}

// TestResult contains the outcome of a network test
type TestResult struct {
	Timestamp    time.Time     // When the test was executed
	EndpointID   string        // Unique identifier for the endpoint
	Protocol     string        // Protocol used (HTTP, TCP, UDP, ICMP)
	Latency      time.Duration // Round-trip time or response latency
	Status       TestStatus    // Test outcome status
	Error        string        // Error message if test failed
	ResponseSize int64         // Size of response in bytes (if applicable)
}

// TestConfig contains parameters for executing a network test
type TestConfig struct {
	Name     string        // Human-readable name for the test
	Address  string        // Target address (URL, IP, hostname)
	Timeout  time.Duration // Maximum time to wait for response
	Protocol string        // Protocol type (HTTP, TCP, UDP, ICMP)
	Config   interface{}   // Protocol-specific configuration
}

// HTTPConfig contains configuration specific to HTTP/HTTPS tests
type HTTPConfig struct {
	Method          string            // HTTP method (GET, POST, etc.)
	Headers         map[string]string // Custom HTTP headers
	Body            string            // Request body for POST/PUT
	FollowRedirects bool              // Whether to follow HTTP redirects
	ValidateSSL     bool              // Whether to validate SSL certificates
	ExpectedStatus  int               // Expected HTTP status code (0 = any 2xx)
}

// TCPConfig contains configuration specific to TCP connection tests
type TCPConfig struct {
	Port           int    // TCP port to connect to
	SendData       string // Optional data to send after connection
	ExpectResponse bool   // Whether to wait for a response
	ExpectedData   string // Expected response data (if any)
}

// UDPConfig contains configuration specific to UDP tests
type UDPConfig struct {
	Port         int    // UDP port to send to
	SendData     string // Data to send in UDP packet
	WaitResponse bool   // Whether to wait for a response
	ResponseSize int    // Maximum response size to read
}

// ICMPConfig contains configuration specific to ICMP ping tests
type ICMPConfig struct {
	Count       int  // Number of ping packets to send
	PacketSize  int  // Size of ICMP packet in bytes
	TTL         int  // Time to live for packets
	Privileged  bool // Whether to use privileged raw sockets
}