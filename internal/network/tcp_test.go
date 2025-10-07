package network

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestTCPTest_GetProtocol(t *testing.T) {
	test := &TCPTest{}
	if got := test.GetProtocol(); got != "TCP" {
		t.Errorf("GetProtocol() = %v, want TCP", got)
	}
}

func TestTCPTest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TestConfig
		wantErr bool
	}{
		{
			name: "valid IPv4 address with port",
			config: TestConfig{
				Address: "192.168.1.1:80",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid hostname with port",
			config: TestConfig{
				Address: "example.com:443",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid IPv6 address with port",
			config: TestConfig{
				Address: "[::1]:8080",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty address",
			config: TestConfig{
				Address: "",
				Timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing port",
			config: TestConfig{
				Address: "example.com",
				Timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too low",
			config: TestConfig{
				Address: "example.com:0",
				Timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: TestConfig{
				Address: "example.com:65536",
				Timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - not a number",
			config: TestConfig{
				Address: "example.com:abc",
				Timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: TestConfig{
				Address: "example.com:80",
				Timeout: 0,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: TestConfig{
				Address: "example.com:80",
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "valid with TCP config",
			config: TestConfig{
				Address: "example.com:80",
				Timeout: 5 * time.Second,
				Config: &TCPConfig{
					Port:           80,
					SendData:       "GET / HTTP/1.0\r\n\r\n",
					ExpectResponse: true,
				},
			},
			wantErr: false,
		},
		{
			name: "port mismatch in config",
			config: TestConfig{
				Address: "example.com:80",
				Timeout: 5 * time.Second,
				Config: &TCPConfig{
					Port: 443, // Doesn't match address port
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config type",
			config: TestConfig{
				Address: "example.com:80",
				Timeout: 5 * time.Second,
				Config:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid high port number",
			config: TestConfig{
				Address: "example.com:65535",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid low port number",
			config: TestConfig{
				Address: "example.com:1",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := &TCPTest{}
			err := test.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTCPTest_Execute_Success(t *testing.T) {
	// Create a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer listener.Close()

	// Accept connections in background
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	test := &TCPTest{}
	config := TestConfig{
		Name:    "Test TCP Server",
		Address: listener.Addr().String(),
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Protocol != "TCP" {
		t.Errorf("Execute() protocol = %v, want TCP", result.Protocol)
	}

	if result.Latency < 0 {
		t.Errorf("Execute() latency = %v, want >= 0", result.Latency)
	}

	if result.EndpointID != "Test TCP Server" {
		t.Errorf("Execute() endpoint ID = %v, want Test TCP Server", result.EndpointID)
	}
}

func TestTCPTest_Execute_ConnectionRefused(t *testing.T) {
	test := &TCPTest{}
	// Use a port that's likely not in use
	config := TestConfig{
		Name:    "Closed Port Test",
		Address: "127.0.0.1:54321",
		Timeout: 1 * time.Second,
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err == nil {
		t.Error("Execute() expected error for closed port, got nil")
	}

	if result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusFailed)
	}

	if result.Error != "connection refused" {
		t.Errorf("Execute() error = %v, want 'connection refused'", result.Error)
	}
}

func TestTCPTest_Execute_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	test := &TCPTest{}
	// Use an IP address that will timeout (filtered/blackhole)
	// 192.0.2.1 is from TEST-NET-1, should be unreachable
	config := TestConfig{
		Name:    "Timeout Test",
		Address: "192.0.2.1:80",
		Timeout: 100 * time.Millisecond,
	}

	ctx := context.Background()
	start := time.Now()
	result, err := test.Execute(ctx, config)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Execute() expected timeout error, got nil")
	}

	// Should timeout or fail
	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v or %v", result.Status, TestStatusTimeout, TestStatusFailed)
	}

	// Should respect timeout (with some margin for processing)
	if elapsed > 2*time.Second {
		t.Errorf("Execute() took %v, expected quick timeout", elapsed)
	}
}

func TestTCPTest_Execute_SendData(t *testing.T) {
	// Create a test TCP server that echoes data
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer listener.Close()

	// Accept connections and echo data
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}
		conn.Write(buffer[:n]) // Echo back
	}()

	test := &TCPTest{}
	config := TestConfig{
		Name:    "Echo Test",
		Address: listener.Addr().String(),
		Timeout: 5 * time.Second,
		Config: &TCPConfig{
			SendData:       "Hello, TCP!",
			ExpectResponse: true,
			ExpectedData:   "Hello, TCP!",
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.ResponseSize == 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}
}

func TestTCPTest_Execute_ExpectedDataMismatch(t *testing.T) {
	// Create a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer listener.Close()

	// Accept connections and send specific data
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read request
		buffer := make([]byte, 1024)
		_, err = conn.Read(buffer)
		if err != nil {
			return
		}

		// Send different response
		conn.Write([]byte("Wrong response"))
	}()

	test := &TCPTest{}
	config := TestConfig{
		Name:    "Data Mismatch Test",
		Address: listener.Addr().String(),
		Timeout: 5 * time.Second,
		Config: &TCPConfig{
			SendData:       "Request",
			ExpectResponse: true,
			ExpectedData:   "Expected response",
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err == nil {
		t.Error("Execute() expected error for data mismatch, got nil")
	}

	if result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusFailed)
	}
}

func TestTCPTest_Execute_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cancellation test in short mode")
	}

	test := &TCPTest{}
	// Use an address that will be slow to timeout (filtered/blackhole)
	config := TestConfig{
		Name:    "Cancellation Test",
		Address: "192.0.2.1:80", // TEST-NET-1, should be unreachable
		Timeout: 30 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result, err := test.Execute(ctx, config)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Execute() expected cancellation error, got nil")
	}

	// Should timeout/fail quickly due to cancellation (within 2 seconds, not the full 30s timeout)
	if elapsed > 2*time.Second {
		t.Errorf("Execute() took %v, expected quick cancellation", elapsed)
	}

	// Status should be timeout or failed
	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want timeout or failed", result.Status)
	}
}

func TestTCPTest_Execute_ValidationErrors(t *testing.T) {
	test := &TCPTest{}
	ctx := context.Background()

	tests := []struct {
		name       string
		config     TestConfig
		wantStatus TestStatus
	}{
		{
			name: "empty address",
			config: TestConfig{
				Address: "",
				Timeout: 5 * time.Second,
			},
			wantStatus: TestStatusError,
		},
		{
			name: "zero timeout",
			config: TestConfig{
				Address: "example.com:80",
				Timeout: 0,
			},
			wantStatus: TestStatusError,
		},
		{
			name: "invalid port",
			config: TestConfig{
				Address: "example.com:99999",
				Timeout: 5 * time.Second,
			},
			wantStatus: TestStatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := test.Execute(ctx, tt.config)
			if err == nil {
				t.Error("Execute() expected error, got nil")
			}
			if result.Status != tt.wantStatus {
				t.Errorf("Execute() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestTCPTest_Execute_IPv6(t *testing.T) {
	// Create a test TCP server on IPv6
	listener, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skip("IPv6 not available, skipping test")
	}
	defer listener.Close()

	// Accept connections in background
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	test := &TCPTest{}
	config := TestConfig{
		Name:    "IPv6 Test",
		Address: listener.Addr().String(),
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}
}

func TestTCPTest_Execute_HTTPRequest(t *testing.T) {
	// Create a simple HTTP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer listener.Close()

	// Simple HTTP response
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read HTTP request
		buffer := make([]byte, 1024)
		_, err = conn.Read(buffer)
		if err != nil {
			return
		}

		// Send HTTP response
		response := "HTTP/1.0 200 OK\r\nContent-Length: 2\r\n\r\nOK"
		conn.Write([]byte(response))
	}()

	test := &TCPTest{}
	config := TestConfig{
		Name:    "HTTP over TCP Test",
		Address: listener.Addr().String(),
		Timeout: 5 * time.Second,
		Config: &TCPConfig{
			SendData:       "GET / HTTP/1.0\r\n\r\n",
			ExpectResponse: true,
			ExpectedData:   "HTTP/1",
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.ResponseSize == 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}
}

func TestTCPTest_Execute_MultipleConnections(t *testing.T) {
	// Create a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer listener.Close()

	// Accept multiple connections
	connectionCount := 0
	go func() {
		for i := 0; i < 3; i++ {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			connectionCount++
			conn.Close()
		}
	}()

	test := &TCPTest{}
	config := TestConfig{
		Name:    "Multiple Connections Test",
		Address: listener.Addr().String(),
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()

	// Execute test 3 times
	for i := 0; i < 3; i++ {
		result, err := test.Execute(ctx, config)
		if err != nil {
			t.Fatalf("Execute() iteration %d unexpected error: %v", i, err)
		}
		if result.Status != TestStatusSuccess {
			t.Errorf("Execute() iteration %d status = %v, want %v", i, result.Status, TestStatusSuccess)
		}
	}

	// Give time for all accepts to complete
	time.Sleep(100 * time.Millisecond)

	if connectionCount != 3 {
		t.Errorf("Expected 3 connections, got %d", connectionCount)
	}
}

func TestTCPTest_Execute_PortRange(t *testing.T) {
	test := &TCPTest{}
	ctx := context.Background()

	// Test common port numbers
	ports := []int{80, 443, 22, 3306, 5432, 8080}

	for _, port := range ports {
		config := TestConfig{
			Name:    fmt.Sprintf("Port %d Test", port),
			Address: fmt.Sprintf("127.0.0.1:%d", port),
			Timeout: 100 * time.Millisecond,
		}

		result, _ := test.Execute(ctx, config)

		// We expect these to fail (ports not open), but should validate correctly
		if result.Protocol != "TCP" {
			t.Errorf("Port %d: protocol = %v, want TCP", port, result.Protocol)
		}

		if result.Status != TestStatusFailed && result.Status != TestStatusTimeout {
			t.Errorf("Port %d: status = %v, want failed or timeout", port, result.Status)
		}
	}
}
