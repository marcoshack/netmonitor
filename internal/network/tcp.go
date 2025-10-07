package network

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// TCPTest implements the NetworkTest interface for TCP connectivity testing
type TCPTest struct{}

// GetProtocol returns the protocol type this test implements
func (t *TCPTest) GetProtocol() string {
	return "TCP"
}

// Validate checks if the configuration is valid for TCP tests
func (t *TCPTest) Validate(config TestConfig) error {
	if config.Address == "" {
		return fmt.Errorf("address is required for TCP test")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}

	// Parse address to validate host:port format
	host, portStr, err := net.SplitHostPort(config.Address)
	if err != nil {
		return fmt.Errorf("invalid address format (expected host:port): %v", err)
	}

	// Validate host is not empty
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Validate port number
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port number: %v", err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}

	// Validate TCP-specific config if provided
	if config.Config != nil {
		tcpConfig, ok := config.Config.(*TCPConfig)
		if !ok {
			return fmt.Errorf("invalid config type for TCP test")
		}

		// Validate Port field if set (should match the port in Address)
		if tcpConfig.Port != 0 && tcpConfig.Port != port {
			return fmt.Errorf("port mismatch: config has %d but address has %d", tcpConfig.Port, port)
		}
	}

	return nil
}

// Execute runs the TCP test with the given configuration
func (t *TCPTest) Execute(ctx context.Context, config TestConfig) (*TestResult, error) {
	startTime := time.Now()

	// Validate configuration
	if err := t.Validate(config); err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "TCP",
			Status:    TestStatusError,
			Error:     err.Error(),
		}, err
	}

	// Get TCP-specific config or use defaults
	tcpConfig := &TCPConfig{
		ExpectResponse: false,
	}
	if config.Config != nil {
		if cfg, ok := config.Config.(*TCPConfig); ok {
			tcpConfig = cfg
		}
	}

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: config.Timeout,
	}

	// Measure connection time
	connStart := time.Now()

	// Establish TCP connection
	conn, err := dialer.DialContext(ctx, "tcp", config.Address)
	if err != nil {
		// Determine the type of error
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			return &TestResult{
				Timestamp: startTime,
				Protocol:  "TCP",
				Status:    TestStatusTimeout,
				Error:     "connection timeout",
				Latency:   0,
			}, err
		}

		// Check for specific error types
		errStr := err.Error()
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
			return &TestResult{
				Timestamp: startTime,
				Protocol:  "TCP",
				Status:    TestStatusTimeout,
				Error:     "connection timeout",
				Latency:   0,
			}, err
		}

		if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "actively refused") {
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "TCP",
				Status:     TestStatusFailed,
				Error:      "connection refused",
				Latency:    0,
			}, err
		}

		if strings.Contains(errStr, "network is unreachable") || strings.Contains(errStr, "no route to host") {
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "TCP",
				Status:     TestStatusFailed,
				Error:      "network unreachable",
				Latency:    0,
			}, err
		}

		return &TestResult{
			Timestamp:  startTime,
			EndpointID: config.Name,
			Protocol:   "TCP",
			Status:     TestStatusFailed,
			Error:      fmt.Sprintf("connection failed: %v", err),
			Latency:    0,
		}, err
	}
	defer conn.Close()

	// Calculate connection establishment time
	connLatency := time.Since(connStart)

	// If we need to send data and expect a response
	var responseSize int64
	if tcpConfig.SendData != "" {
		// Set write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(config.Timeout)); err != nil {
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "TCP",
				Status:     TestStatusError,
				Error:      fmt.Sprintf("failed to set write deadline: %v", err),
				Latency:    connLatency,
			}, err
		}

		// Send data
		_, err := conn.Write([]byte(tcpConfig.SendData))
		if err != nil {
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "TCP",
				Status:     TestStatusFailed,
				Error:      fmt.Sprintf("failed to send data: %v", err),
				Latency:    connLatency,
			}, err
		}

		// If expecting a response
		if tcpConfig.ExpectResponse {
			// Set read deadline
			if err := conn.SetReadDeadline(time.Now().Add(config.Timeout)); err != nil {
				return &TestResult{
					Timestamp:  startTime,
					EndpointID: config.Name,
					Protocol:   "TCP",
					Status:     TestStatusError,
					Error:      fmt.Sprintf("failed to set read deadline: %v", err),
					Latency:    connLatency,
				}, err
			}

			// Read response
			buffer := make([]byte, 4096)
			n, err := conn.Read(buffer)
			if err != nil {
				return &TestResult{
					Timestamp:  startTime,
					EndpointID: config.Name,
					Protocol:   "TCP",
					Status:     TestStatusFailed,
					Error:      fmt.Sprintf("failed to read response: %v", err),
					Latency:    connLatency,
				}, err
			}

			responseSize = int64(n)

			// Check if expected data matches
			if tcpConfig.ExpectedData != "" {
				responseStr := string(buffer[:n])
				if !strings.Contains(responseStr, tcpConfig.ExpectedData) {
					return &TestResult{
						Timestamp:    startTime,
						EndpointID:   config.Name,
						Protocol:     "TCP",
						Status:       TestStatusFailed,
						Error:        fmt.Sprintf("expected data '%s' not found in response", tcpConfig.ExpectedData),
						Latency:      connLatency,
						ResponseSize: responseSize,
					}, fmt.Errorf("expected data not found")
				}
			}
		}
	}

	// Connection successful
	return &TestResult{
		Timestamp:    startTime,
		EndpointID:   config.Name,
		Protocol:     "TCP",
		Status:       TestStatusSuccess,
		Latency:      connLatency,
		ResponseSize: responseSize,
	}, nil
}
