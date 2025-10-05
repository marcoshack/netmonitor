package network

import (
	"context"
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// ICMPTest implements the NetworkTest interface for ICMP ping functionality
type ICMPTest struct{}

// GetProtocol returns the protocol type this test implements
func (t *ICMPTest) GetProtocol() string {
	return "ICMP"
}

// Validate checks if the configuration is valid for ICMP tests
func (t *ICMPTest) Validate(config TestConfig) error {
	if config.Address == "" {
		return fmt.Errorf("address is required for ICMP test")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}

	// Validate ICMP-specific config if provided
	if config.Config != nil {
		icmpConfig, ok := config.Config.(*ICMPConfig)
		if !ok {
			return fmt.Errorf("invalid config type for ICMP test")
		}

		if icmpConfig.Count < 0 {
			return fmt.Errorf("count must be non-negative")
		}

		if icmpConfig.PacketSize < 0 || icmpConfig.PacketSize > 65507 {
			return fmt.Errorf("packet size must be between 0 and 65507")
		}

		if icmpConfig.TTL < 0 || icmpConfig.TTL > 255 {
			return fmt.Errorf("TTL must be between 0 and 255")
		}
	}

	return nil
}

// Execute runs the ICMP ping test with the given configuration
func (t *ICMPTest) Execute(ctx context.Context, config TestConfig) (*TestResult, error) {
	startTime := time.Now()

	// Validate configuration
	if err := t.Validate(config); err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "ICMP",
			Status:    TestStatusError,
			Error:     err.Error(),
		}, err
	}

	// Get ICMP-specific config or use defaults
	icmpConfig := &ICMPConfig{
		Count:      1,
		PacketSize: 64,
		TTL:        64,
		Privileged: false,
	}
	if config.Config != nil {
		if cfg, ok := config.Config.(*ICMPConfig); ok {
			icmpConfig = cfg
		}
	}

	// Create a pinger
	pinger, err := probing.NewPinger(config.Address)
	if err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "ICMP",
			Status:    TestStatusError,
			Error:     fmt.Sprintf("failed to create pinger: %v", err),
		}, err
	}

	// IMPORTANT: On Windows, we must use SetPrivileged(true) even though it doesn't require admin
	// See: https://github.com/prometheus-community/pro-bing documentation
	// This works on Windows 10+ without elevated privileges
	pinger.SetPrivileged(true)

	// Configure pinger
	pinger.Count = icmpConfig.Count
	if icmpConfig.Count <= 0 {
		pinger.Count = 1
	}
	pinger.Size = icmpConfig.PacketSize
	pinger.Timeout = config.Timeout

	// Resolve the address to ensure it's valid
	if err := pinger.Resolve(); err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "ICMP",
			Status:    TestStatusError,
			Error:     fmt.Sprintf("failed to resolve address: %v", err),
		}, err
	}

	// Run the ping
	err = pinger.RunWithContext(ctx)
	stats := pinger.Statistics()

	result := &TestResult{
		Timestamp:  startTime,
		EndpointID: config.Name,
		Protocol:   "ICMP",
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = TestStatusTimeout
			result.Error = "ping timeout"
		} else {
			result.Status = TestStatusFailed
			result.Error = err.Error()
		}
		result.Latency = 0
		return result, err
	}

	// Check if we received any packets
	if stats.PacketsRecv == 0 {
		result.Status = TestStatusTimeout
		result.Error = "no ICMP reply received"
		result.Latency = 0
		return result, fmt.Errorf("no ICMP reply received")
	}

	result.Status = TestStatusSuccess
	result.Latency = stats.AvgRtt
	return result, nil
}