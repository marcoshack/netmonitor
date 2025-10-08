package network

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// UDPTest implements the NetworkTest interface for UDP connectivity testing
type UDPTest struct{}

// GetProtocol returns the protocol type this test implements
func (u *UDPTest) GetProtocol() string {
	return "UDP"
}

// Validate checks if the configuration is valid for UDP tests
func (u *UDPTest) Validate(config TestConfig) error {
	if config.Address == "" {
		return fmt.Errorf("address is required for UDP test")
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

	// Validate UDP-specific config if provided
	if config.Config != nil {
		udpConfig, ok := config.Config.(*UDPConfig)
		if !ok {
			return fmt.Errorf("invalid config type for UDP test")
		}

		// Validate Port field if set (should match the port in Address)
		if udpConfig.Port != 0 && udpConfig.Port != port {
			return fmt.Errorf("port mismatch: config has %d but address has %d", udpConfig.Port, port)
		}

		// Validate ResponseSize if set
		if udpConfig.ResponseSize < 0 {
			return fmt.Errorf("response size cannot be negative")
		}
	}

	return nil
}

// Execute runs the UDP test with the given configuration
func (u *UDPTest) Execute(ctx context.Context, config TestConfig) (*TestResult, error) {
	startTime := time.Now()

	// Validate configuration
	if err := u.Validate(config); err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "UDP",
			Status:    TestStatusError,
			Error:     err.Error(),
		}, err
	}

	// Get UDP-specific config or use defaults
	udpConfig := &UDPConfig{
		SendData:     "",
		WaitResponse: false,
		ResponseSize: 1024, // Default 1KB buffer
	}
	if config.Config != nil {
		if cfg, ok := config.Config.(*UDPConfig); ok {
			udpConfig = cfg
			// Ensure ResponseSize has a reasonable default
			if udpConfig.ResponseSize == 0 {
				udpConfig.ResponseSize = 1024
			}
		}
	}

	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", config.Address)
	if err != nil {
		return &TestResult{
			Timestamp:  startTime,
			EndpointID: config.Name,
			Protocol:   "UDP",
			Status:     TestStatusFailed,
			Error:      fmt.Sprintf("failed to resolve address: %v", err),
			Latency:    0,
		}, err
	}

	// Create UDP connection
	connStart := time.Now()

	// Dial UDP with context support
	var conn *net.UDPConn
	dialErr := make(chan error, 1)

	go func() {
		var err error
		conn, err = net.DialUDP("udp", nil, udpAddr)
		dialErr <- err
	}()

	// Wait for connection or context cancellation
	select {
	case <-ctx.Done():
		return &TestResult{
			Timestamp:  startTime,
			EndpointID: config.Name,
			Protocol:   "UDP",
			Status:     TestStatusTimeout,
			Error:      "connection timeout",
			Latency:    0,
		}, ctx.Err()
	case err := <-dialErr:
		if err != nil {
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "UDP",
				Status:     TestStatusFailed,
				Error:      fmt.Sprintf("failed to create UDP connection: %v", err),
				Latency:    0,
			}, err
		}
	}
	defer conn.Close()

	// For UDP, we need to send data to actually test the connection
	// If no data specified, use a default probe
	dataToSend := []byte(udpConfig.SendData)
	if len(dataToSend) == 0 {
		dataToSend = []byte("PROBE") // Default probe packet
	}

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(config.Timeout)); err != nil {
		return &TestResult{
			Timestamp:  startTime,
			EndpointID: config.Name,
			Protocol:   "UDP",
			Status:     TestStatusError,
			Error:      fmt.Sprintf("failed to set write deadline: %v", err),
			Latency:    time.Since(connStart),
		}, err
	}

	// Send UDP packet
	_, err = conn.Write(dataToSend)
	if err != nil {
		// Check for timeout
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "UDP",
				Status:     TestStatusTimeout,
				Error:      "write timeout",
				Latency:    0,
			}, err
		}

		return &TestResult{
			Timestamp:  startTime,
			EndpointID: config.Name,
			Protocol:   "UDP",
			Status:     TestStatusFailed,
			Error:      fmt.Sprintf("failed to send UDP packet: %v", err),
			Latency:    time.Since(connStart),
		}, err
	}

	// If we're waiting for a response
	var responseSize int64
	if udpConfig.WaitResponse {
		// Set read deadline
		if err := conn.SetReadDeadline(time.Now().Add(config.Timeout)); err != nil {
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "UDP",
				Status:     TestStatusError,
				Error:      fmt.Sprintf("failed to set read deadline: %v", err),
				Latency:    time.Since(connStart),
			}, err
		}

		// Read response in a goroutine to support context cancellation
		type readResult struct {
			n   int
			err error
		}
		readChan := make(chan readResult, 1)

		go func() {
			buffer := make([]byte, udpConfig.ResponseSize)
			n, err := conn.Read(buffer)
			readChan <- readResult{n: n, err: err}
		}()

		// Wait for response or context cancellation
		var n int
		select {
		case <-ctx.Done():
			// Context cancelled, close connection to interrupt read
			conn.Close()
			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "UDP",
				Status:     TestStatusTimeout,
				Error:      "context cancelled",
				Latency:    time.Since(connStart),
			}, ctx.Err()
		case result := <-readChan:
			n = result.n
			err = result.err
		}

		if err != nil {
			// Check for timeout - this is expected for non-responsive UDP services
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return &TestResult{
					Timestamp:  startTime,
					EndpointID: config.Name,
					Protocol:   "UDP",
					Status:     TestStatusTimeout,
					Error:      "no response received (timeout)",
					Latency:    config.Timeout,
				}, err
			}

			// Check for connection refused (ICMP port unreachable)
			if strings.Contains(err.Error(), "connection refused") {
				return &TestResult{
					Timestamp:  startTime,
					EndpointID: config.Name,
					Protocol:   "UDP",
					Status:     TestStatusFailed,
					Error:      "port unreachable",
					Latency:    time.Since(connStart),
				}, err
			}

			return &TestResult{
				Timestamp:  startTime,
				EndpointID: config.Name,
				Protocol:   "UDP",
				Status:     TestStatusFailed,
				Error:      fmt.Sprintf("failed to read response: %v", err),
				Latency:    time.Since(connStart),
			}, err
		}

		responseSize = int64(n)
		latency := time.Since(connStart)

		// If we got a response, it's a success
		return &TestResult{
			Timestamp:    startTime,
			EndpointID:   config.Name,
			Protocol:     "UDP",
			Status:       TestStatusSuccess,
			Latency:      latency,
			ResponseSize: responseSize,
		}, nil
	}

	// If not waiting for response, consider successful send as success
	// (UDP is connectionless, so we can't guarantee delivery without a response)
	return &TestResult{
		Timestamp:    startTime,
		EndpointID:   config.Name,
		Protocol:     "UDP",
		Status:       TestStatusSuccess,
		Latency:      time.Since(connStart),
		ResponseSize: responseSize,
	}, nil
}
