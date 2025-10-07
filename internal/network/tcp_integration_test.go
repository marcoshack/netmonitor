package network

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestTCPTest_Integration_PublicServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &TCPTest{}
	ctx := context.Background()

	tests := []struct {
		name       string
		address    string
		wantStatus TestStatus
		timeout    time.Duration
	}{
		{
			name:       "Google DNS - Port 53",
			address:    "8.8.8.8:53",
			wantStatus: TestStatusSuccess,
			timeout:    5 * time.Second,
		},
		{
			name:       "Google Public DNS HTTPS",
			address:    "8.8.8.8:443",
			wantStatus: TestStatusSuccess,
			timeout:    5 * time.Second,
		},
		{
			name:       "Cloudflare DNS",
			address:    "1.1.1.1:53",
			wantStatus: TestStatusSuccess,
			timeout:    5 * time.Second,
		},
		{
			name:       "Localhost SSH (likely closed)",
			address:    "127.0.0.1:22",
			wantStatus: TestStatusFailed, // Expected to fail on most systems
			timeout:    1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := TestConfig{
				Name:    tt.name,
				Address: tt.address,
				Timeout: tt.timeout,
			}

			result, err := test.Execute(ctx, config)

			// For public services, we might get success or timeout based on network
			// We mainly verify that the test runs without panicking
			if result == nil {
				t.Fatal("Execute() returned nil result")
			}

			if result.Protocol != "TCP" {
				t.Errorf("Execute() protocol = %v, want TCP", result.Protocol)
			}

			t.Logf("Result: Status=%s, Latency=%v, Error=%s", result.Status, result.Latency, result.Error)

			// For successful connections, verify latency is measured
			if result.Status == TestStatusSuccess && result.Latency <= 0 {
				t.Errorf("Execute() latency = %v, want > 0 for successful connection", result.Latency)
			}

			// Verify error handling
			if result.Status == TestStatusFailed || result.Status == TestStatusTimeout {
				if err == nil {
					t.Error("Execute() expected error for failed/timeout status, got nil")
				}
			}
		})
	}
}

func TestTCPTest_Integration_CommonPorts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &TCPTest{}
	ctx := context.Background()

	// Test common ports on localhost (most should be closed)
	commonPorts := []struct {
		port        int
		serviceName string
	}{
		{80, "HTTP"},
		{443, "HTTPS"},
		{22, "SSH"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{6379, "Redis"},
		{27017, "MongoDB"},
		{8080, "HTTP Alt"},
	}

	for _, p := range commonPorts {
		t.Run(p.serviceName, func(t *testing.T) {
			config := TestConfig{
				Name:    p.serviceName + " Test",
				Address: fmt.Sprintf("127.0.0.1:%d", p.port),
				Timeout: 500 * time.Millisecond,
			}

			// We expect most to fail, but they should fail gracefully
			result, _ := test.Execute(ctx, config)

			if result == nil {
				t.Fatal("Execute() returned nil result")
			}

			// Should be either failed or timeout, not error
			if result.Status != TestStatusFailed && result.Status != TestStatusTimeout && result.Status != TestStatusError {
				t.Logf("Unexpected status for port %d: %s", p.port, result.Status)
			}

			t.Logf("Port %d (%s): Status=%s, Error=%s", p.port, p.serviceName, result.Status, result.Error)
		})
	}
}

func TestTCPTest_Integration_HTTPService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &TCPTest{}
	ctx := context.Background()

	// Test actual HTTP communication over TCP
	config := TestConfig{
		Name:    "HTTP Request Test",
		Address: "www.google.com:80",
		Timeout: 10 * time.Second,
		Config: &TCPConfig{
			SendData:       "GET / HTTP/1.0\r\nHost: www.google.com\r\n\r\n",
			ExpectResponse: true,
			ExpectedData:   "HTTP/",
		},
	}

	result, err := test.Execute(ctx, config)

	// This might fail due to network issues, but we verify the behavior
	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	t.Logf("HTTP over TCP Result: Status=%s, Latency=%v, ResponseSize=%d, Error=%s",
		result.Status, result.Latency, result.ResponseSize, result.Error)

	// If successful, verify we got a response
	if result.Status == TestStatusSuccess {
		if result.ResponseSize == 0 {
			t.Error("Execute() successful but response size is 0")
		}
		if result.Latency <= 0 {
			t.Error("Execute() successful but latency is 0")
		}
		if err != nil {
			t.Errorf("Execute() successful but returned error: %v", err)
		}
	}
}

func TestTCPTest_Integration_IPv6(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &TCPTest{}
	ctx := context.Background()

	// Test IPv6 connectivity to Google DNS
	config := TestConfig{
		Name:    "IPv6 Google DNS",
		Address: "[2001:4860:4860::8888]:53",
		Timeout: 5 * time.Second,
	}

	result, err := test.Execute(ctx, config)

	// IPv6 might not be available on all systems
	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	t.Logf("IPv6 Result: Status=%s, Latency=%v, Error=%s", result.Status, result.Latency, result.Error)

	// If it succeeds, verify the connection details
	if result.Status == TestStatusSuccess {
		if result.Latency <= 0 {
			t.Error("Execute() successful but latency is 0")
		}
		if err != nil {
			t.Errorf("Execute() successful but returned error: %v", err)
		}
	} else {
		// If it fails, it should be a proper failure, not an error
		t.Logf("IPv6 test failed (this is expected on IPv4-only networks): %s", result.Error)
	}
}

func TestTCPTest_Integration_ConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &TCPTest{}
	ctx := context.Background()

	// Test multiple concurrent connections to the same service
	numConnections := 5
	results := make(chan *TestResult, numConnections)
	errors := make(chan error, numConnections)

	config := TestConfig{
		Name:    "Concurrent Connection Test",
		Address: "8.8.8.8:53",
		Timeout: 5 * time.Second,
	}

	// Execute tests concurrently
	for i := 0; i < numConnections; i++ {
		go func() {
			result, err := test.Execute(ctx, config)
			results <- result
			errors <- err
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConnections; i++ {
		result := <-results
		err := <-errors

		if result.Status == TestStatusSuccess {
			successCount++
		}

		t.Logf("Connection %d: Status=%s, Latency=%v, Error=%v", i+1, result.Status, result.Latency, err)
	}

	// At least some connections should succeed (unless network is down)
	t.Logf("Successful concurrent connections: %d/%d", successCount, numConnections)
}

func TestTCPTest_Integration_DatabasePorts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &TCPTest{}
	ctx := context.Background()

	// Test common database ports on localhost
	databases := []struct {
		name    string
		port    int
		timeout time.Duration
	}{
		{"PostgreSQL", 5432, 1 * time.Second},
		{"MySQL", 3306, 1 * time.Second},
		{"Redis", 6379, 1 * time.Second},
		{"MongoDB", 27017, 1 * time.Second},
	}

	for _, db := range databases {
		t.Run(db.name, func(t *testing.T) {
			config := TestConfig{
				Name:    db.name + " Connection Test",
				Address: fmt.Sprintf("127.0.0.1:%d", db.port),
				Timeout: db.timeout,
			}

			result, _ := test.Execute(ctx, config)

			if result == nil {
				t.Fatal("Execute() returned nil result")
			}

			// Most will fail (not running), but should fail cleanly
			t.Logf("%s: Status=%s, Error=%s", db.name, result.Status, result.Error)

			// Verify protocol is correct
			if result.Protocol != "TCP" {
				t.Errorf("Execute() protocol = %v, want TCP", result.Protocol)
			}
		})
	}
}

func TestTCPTest_Integration_Timeouts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &TCPTest{}
	ctx := context.Background()

	// Test with very short timeout on a potentially slow connection
	config := TestConfig{
		Name:    "Short Timeout Test",
		Address: "8.8.8.8:53",
		Timeout: 1 * time.Millisecond, // Very short timeout
	}

	start := time.Now()
	result, _ := test.Execute(ctx, config)
	elapsed := time.Since(start)

	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	t.Logf("Short timeout result: Status=%s, Elapsed=%v, Error=%s", result.Status, elapsed, result.Error)

	// Should complete quickly (within reasonable margin)
	if elapsed > 2*time.Second {
		t.Errorf("Execute() took %v, expected quick timeout", elapsed)
	}

	// Status should be timeout or failed
	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed && result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, expected timeout/failed/success", result.Status)
	}
}
