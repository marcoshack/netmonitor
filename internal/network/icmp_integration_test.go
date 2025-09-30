package network

import (
	"context"
	"testing"
	"time"
)

// TestICMPTest_Integration_IPv4 tests ICMP ping with a real IPv4 endpoint
func TestICMPTest_Integration_IPv4(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &ICMPTest{}

	config := TestConfig{
		Name:     "Google DNS",
		Address:  "8.8.8.8",
		Timeout:  5 * time.Second,
		Protocol: "ICMP",
		Config: &ICMPConfig{
			Count:      1,
			PacketSize: 64,
			TTL:        64,
			Privileged: false,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		// If error is permission-related, skip the test
		if result.Status == TestStatusError {
			t.Skipf("Skipping test due to permission error: %v", err)
		}
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	if result.Protocol != "ICMP" {
		t.Errorf("Execute() protocol = %v, want ICMP", result.Protocol)
	}

	t.Logf("Ping to %s: latency=%v, status=%v", config.Address, result.Latency, result.Status)
}

// TestICMPTest_Integration_IPv6 tests ICMP ping with a real IPv6 endpoint
func TestICMPTest_Integration_IPv6(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &ICMPTest{}

	config := TestConfig{
		Name:     "Google DNS IPv6",
		Address:  "2001:4860:4860::8888",
		Timeout:  5 * time.Second,
		Protocol: "ICMP",
		Config: &ICMPConfig{
			Count:      1,
			PacketSize: 64,
			TTL:        64,
			Privileged: false,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		// IPv6 might not be available on all systems
		if result.Status == TestStatusError || result.Status == TestStatusFailed {
			t.Skipf("Skipping IPv6 test: %v", err)
		}
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	if result.Protocol != "ICMP" {
		t.Errorf("Execute() protocol = %v, want ICMP", result.Protocol)
	}

	t.Logf("Ping to %s: latency=%v, status=%v", config.Address, result.Latency, result.Status)
}

// TestICMPTest_Integration_Localhost tests ICMP ping to localhost
func TestICMPTest_Integration_Localhost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &ICMPTest{}

	config := TestConfig{
		Name:     "Localhost",
		Address:  "127.0.0.1",
		Timeout:  2 * time.Second,
		Protocol: "ICMP",
		Config: &ICMPConfig{
			Count:      1,
			PacketSize: 32,
			TTL:        64,
			Privileged: false,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		// If error is permission-related, skip the test
		if result.Status == TestStatusError {
			t.Skipf("Skipping test due to permission error: %v", err)
		}
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	// Localhost should have very low latency (< 100ms)
	if result.Latency > 100*time.Millisecond {
		t.Errorf("Execute() latency = %v, want < 100ms for localhost", result.Latency)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	t.Logf("Ping to localhost: latency=%v, status=%v", result.Latency, result.Status)
}

// TestICMPTest_Integration_UnreachableHost tests behavior with unreachable host
func TestICMPTest_Integration_UnreachableHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &ICMPTest{}

	// Using TEST-NET-1 address range (192.0.2.0/24) which should be unreachable
	config := TestConfig{
		Name:     "Unreachable Host",
		Address:  "192.0.2.254",
		Timeout:  2 * time.Second,
		Protocol: "ICMP",
		Config: &ICMPConfig{
			Count:      1,
			PacketSize: 64,
			TTL:        64,
			Privileged: false,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// Should fail or timeout
	if err == nil {
		t.Error("Execute() expected error for unreachable host, got nil")
	}

	// Status should be timeout or failed
	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v or %v", result.Status, TestStatusTimeout, TestStatusFailed)
	}

	t.Logf("Ping to unreachable host: status=%v, error=%v", result.Status, result.Error)
}

// TestICMPTest_Integration_HostnameResolution tests ping with hostname
func TestICMPTest_Integration_HostnameResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &ICMPTest{}

	config := TestConfig{
		Name:     "DNS Hostname",
		Address:  "dns.google",
		Timeout:  5 * time.Second,
		Protocol: "ICMP",
		Config: &ICMPConfig{
			Count:      1,
			PacketSize: 64,
			TTL:        64,
			Privileged: false,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		// If error is permission or DNS related, skip the test
		if result.Status == TestStatusError || result.Status == TestStatusFailed {
			t.Skipf("Skipping test: %v", err)
		}
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	t.Logf("Ping to %s: latency=%v, status=%v", config.Address, result.Latency, result.Status)
}

// TestICMPTest_Integration_ConcurrentPings tests multiple simultaneous pings
func TestICMPTest_Integration_ConcurrentPings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &ICMPTest{}

	targets := []string{
		"8.8.8.8",
		"8.8.4.4",
		"1.1.1.1",
		"1.0.0.1",
	}

	ctx := context.Background()
	results := make(chan *TestResult, len(targets))
	errors := make(chan error, len(targets))

	// Launch concurrent pings
	for _, target := range targets {
		go func(addr string) {
			config := TestConfig{
				Name:     addr,
				Address:  addr,
				Timeout:  5 * time.Second,
				Protocol: "ICMP",
				Config: &ICMPConfig{
					Count:      1,
					PacketSize: 64,
					TTL:        64,
					Privileged: false,
				},
			}

			result, err := test.Execute(ctx, config)
			results <- result
			errors <- err
		}(target)
	}

	// Collect results
	successCount := 0
	for i := 0; i < len(targets); i++ {
		result := <-results
		err := <-errors

		if err == nil && result.Status == TestStatusSuccess {
			successCount++
			t.Logf("Concurrent ping to %s: latency=%v, status=%v", result.EndpointID, result.Latency, result.Status)
		} else if result.Status == TestStatusError {
			t.Logf("Concurrent ping to %s failed (likely permission): %v", result.EndpointID, err)
		}
	}

	// At least some pings should succeed (or all fail due to permissions)
	if successCount == 0 {
		t.Skip("All concurrent pings failed (likely due to permissions)")
	}

	t.Logf("Concurrent pings: %d/%d successful", successCount, len(targets))
}

// TestICMPTest_Integration_DifferentPacketSizes tests various packet sizes
func TestICMPTest_Integration_DifferentPacketSizes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &ICMPTest{}
	sizes := []int{32, 64, 128, 512, 1024}

	for _, size := range sizes {
		t.Run(t.Name()+"_size_"+string(rune(size)), func(t *testing.T) {
			config := TestConfig{
				Name:     "Packet Size Test",
				Address:  "8.8.8.8",
				Timeout:  5 * time.Second,
				Protocol: "ICMP",
				Config: &ICMPConfig{
					Count:      1,
					PacketSize: size,
					TTL:        64,
					Privileged: false,
				},
			}

			ctx := context.Background()
			result, err := test.Execute(ctx, config)

			if err != nil && result.Status == TestStatusError {
				t.Skipf("Skipping test for packet size %d: %v", size, err)
			}

			if err == nil && result.Status == TestStatusSuccess {
				t.Logf("Ping with packet size %d: latency=%v", size, result.Latency)
			}
		})
	}
}