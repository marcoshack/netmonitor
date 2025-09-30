package network

import (
	"context"
	"testing"
	"time"
)

// MockNetworkTest is a mock implementation of the NetworkTest interface for testing
type MockNetworkTest struct {
	protocol string
	result   *TestResult
	err      error
}

func (m *MockNetworkTest) Execute(ctx context.Context, config TestConfig) (*TestResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *MockNetworkTest) GetProtocol() string {
	return m.protocol
}

func (m *MockNetworkTest) Validate(config TestConfig) error {
	if config.Address == "" {
		return &ValidationError{Field: "Address", Message: "address cannot be empty"}
	}
	if config.Timeout <= 0 {
		return &ValidationError{Field: "Timeout", Message: "timeout must be positive"}
	}
	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// TestMockNetworkTest verifies the mock implementation satisfies the interface
func TestMockNetworkTest(t *testing.T) {
	var _ NetworkTest = (*MockNetworkTest)(nil)

	mock := &MockNetworkTest{
		protocol: "TEST",
		result: &TestResult{
			Timestamp:  time.Now(),
			EndpointID: "test-endpoint",
			Protocol:   "TEST",
			Latency:    50 * time.Millisecond,
			Status:     TestStatusSuccess,
		},
	}

	if mock.GetProtocol() != "TEST" {
		t.Errorf("Expected protocol TEST, got %s", mock.GetProtocol())
	}

	ctx := context.Background()
	config := TestConfig{
		Name:     "Test Config",
		Address:  "test.example.com",
		Timeout:  5 * time.Second,
		Protocol: "TEST",
	}

	result, err := mock.Execute(ctx, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Protocol != "TEST" {
		t.Errorf("Expected protocol TEST, got %s", result.Protocol)
	}
}

// TestTestResultCreation verifies TestResult can be created with all fields
func TestTestResultCreation(t *testing.T) {
	now := time.Now()
	result := TestResult{
		Timestamp:    now,
		EndpointID:   "endpoint-123",
		Protocol:     "HTTP",
		Latency:      100 * time.Millisecond,
		Status:       TestStatusSuccess,
		Error:        "",
		ResponseSize: 1024,
	}

	if result.Timestamp != now {
		t.Errorf("Expected timestamp %v, got %v", now, result.Timestamp)
	}
	if result.EndpointID != "endpoint-123" {
		t.Errorf("Expected endpoint ID endpoint-123, got %s", result.EndpointID)
	}
	if result.Protocol != "HTTP" {
		t.Errorf("Expected protocol HTTP, got %s", result.Protocol)
	}
	if result.Latency != 100*time.Millisecond {
		t.Errorf("Expected latency 100ms, got %v", result.Latency)
	}
	if result.Status != TestStatusSuccess {
		t.Errorf("Expected status success, got %s", result.Status)
	}
	if result.ResponseSize != 1024 {
		t.Errorf("Expected response size 1024, got %d", result.ResponseSize)
	}
}

// TestTestResultWithError verifies TestResult handles error conditions
func TestTestResultWithError(t *testing.T) {
	result := TestResult{
		Timestamp:  time.Now(),
		EndpointID: "failed-endpoint",
		Protocol:   "TCP",
		Status:     TestStatusFailed,
		Error:      "connection refused",
	}

	if result.Status != TestStatusFailed {
		t.Errorf("Expected status failed, got %s", result.Status)
	}
	if result.Error != "connection refused" {
		t.Errorf("Expected error 'connection refused', got %s", result.Error)
	}
}

// TestTestConfigCreation verifies TestConfig can be created properly
func TestTestConfigCreation(t *testing.T) {
	config := TestConfig{
		Name:     "Google DNS Ping",
		Address:  "8.8.8.8",
		Timeout:  5 * time.Second,
		Protocol: "ICMP",
		Config: &ICMPConfig{
			Count:      3,
			PacketSize: 64,
			TTL:        64,
			Privileged: false,
		},
	}

	if config.Name != "Google DNS Ping" {
		t.Errorf("Expected name 'Google DNS Ping', got %s", config.Name)
	}
	if config.Address != "8.8.8.8" {
		t.Errorf("Expected address 8.8.8.8, got %s", config.Address)
	}
	if config.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", config.Timeout)
	}
	if config.Protocol != "ICMP" {
		t.Errorf("Expected protocol ICMP, got %s", config.Protocol)
	}

	icmpConfig, ok := config.Config.(*ICMPConfig)
	if !ok {
		t.Fatal("Expected ICMPConfig type")
	}
	if icmpConfig.Count != 3 {
		t.Errorf("Expected count 3, got %d", icmpConfig.Count)
	}
}

// TestHTTPConfigCreation verifies HTTPConfig structure
func TestHTTPConfigCreation(t *testing.T) {
	httpConfig := HTTPConfig{
		Method: "GET",
		Headers: map[string]string{
			"User-Agent": "NetMonitor/1.0",
			"Accept":     "application/json",
		},
		FollowRedirects: true,
		ValidateSSL:     true,
		ExpectedStatus:  200,
	}

	if httpConfig.Method != "GET" {
		t.Errorf("Expected method GET, got %s", httpConfig.Method)
	}
	if httpConfig.Headers["User-Agent"] != "NetMonitor/1.0" {
		t.Errorf("Expected User-Agent header")
	}
	if !httpConfig.FollowRedirects {
		t.Error("Expected FollowRedirects to be true")
	}
	if httpConfig.ExpectedStatus != 200 {
		t.Errorf("Expected status 200, got %d", httpConfig.ExpectedStatus)
	}
}

// TestTCPConfigCreation verifies TCPConfig structure
func TestTCPConfigCreation(t *testing.T) {
	tcpConfig := TCPConfig{
		Port:           80,
		SendData:       "GET / HTTP/1.0\r\n\r\n",
		ExpectResponse: true,
		ExpectedData:   "HTTP/1",
	}

	if tcpConfig.Port != 80 {
		t.Errorf("Expected port 80, got %d", tcpConfig.Port)
	}
	if tcpConfig.SendData == "" {
		t.Error("Expected SendData to be set")
	}
	if !tcpConfig.ExpectResponse {
		t.Error("Expected ExpectResponse to be true")
	}
}

// TestUDPConfigCreation verifies UDPConfig structure
func TestUDPConfigCreation(t *testing.T) {
	udpConfig := UDPConfig{
		Port:         53,
		SendData:     "DNS query packet",
		WaitResponse: true,
		ResponseSize: 512,
	}

	if udpConfig.Port != 53 {
		t.Errorf("Expected port 53, got %d", udpConfig.Port)
	}
	if udpConfig.ResponseSize != 512 {
		t.Errorf("Expected response size 512, got %d", udpConfig.ResponseSize)
	}
}

// TestICMPConfigCreation verifies ICMPConfig structure
func TestICMPConfigCreation(t *testing.T) {
	icmpConfig := ICMPConfig{
		Count:      5,
		PacketSize: 64,
		TTL:        128,
		Privileged: false,
	}

	if icmpConfig.Count != 5 {
		t.Errorf("Expected count 5, got %d", icmpConfig.Count)
	}
	if icmpConfig.PacketSize != 64 {
		t.Errorf("Expected packet size 64, got %d", icmpConfig.PacketSize)
	}
	if icmpConfig.TTL != 128 {
		t.Errorf("Expected TTL 128, got %d", icmpConfig.TTL)
	}
}

// TestTestStatus verifies TestStatus constants
func TestTestStatus(t *testing.T) {
	tests := []struct {
		status   TestStatus
		expected string
	}{
		{TestStatusSuccess, "success"},
		{TestStatusFailed, "failed"},
		{TestStatusTimeout, "timeout"},
		{TestStatusError, "error"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("Expected status %s, got %s", tt.expected, string(tt.status))
		}
	}
}

// TestContextCancellation verifies context cancellation handling
func TestContextCancellation(t *testing.T) {
	mock := &MockNetworkTest{
		protocol: "TEST",
		result: &TestResult{
			Status: TestStatusSuccess,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := TestConfig{
		Address: "test.example.com",
		Timeout: 5 * time.Second,
	}

	// Mock should still execute even with cancelled context
	// Real implementations should check ctx.Done()
	_, err := mock.Execute(ctx, config)
	if err != nil {
		t.Errorf("Mock should execute regardless of context, got error: %v", err)
	}
}

// TestValidation verifies configuration validation
func TestValidation(t *testing.T) {
	mock := &MockNetworkTest{}

	tests := []struct {
		name        string
		config      TestConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: TestConfig{
				Address: "example.com",
				Timeout: 5 * time.Second,
			},
			expectError: false,
		},
		{
			name: "empty address",
			config: TestConfig{
				Address: "",
				Timeout: 5 * time.Second,
			},
			expectError: true,
		},
		{
			name: "zero timeout",
			config: TestConfig{
				Address: "example.com",
				Timeout: 0,
			},
			expectError: true,
		},
		{
			name: "negative timeout",
			config: TestConfig{
				Address: "example.com",
				Timeout: -1 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.Validate(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}