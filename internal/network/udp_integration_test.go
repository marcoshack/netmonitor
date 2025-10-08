package network

import (
	"context"
	"encoding/binary"
	"testing"
	"time"
)

// DNS query packet for google.com (A record)
// This is a simplified DNS query for testing purposes
func createDNSQuery() []byte {
	// DNS Header (12 bytes)
	query := make([]byte, 512)
	// Transaction ID
	binary.BigEndian.PutUint16(query[0:2], 0x1234)
	// Flags: standard query
	binary.BigEndian.PutUint16(query[2:4], 0x0100)
	// Questions: 1
	binary.BigEndian.PutUint16(query[4:6], 0x0001)
	// Answer RRs: 0
	binary.BigEndian.PutUint16(query[6:8], 0x0000)
	// Authority RRs: 0
	binary.BigEndian.PutUint16(query[8:10], 0x0000)
	// Additional RRs: 0
	binary.BigEndian.PutUint16(query[10:12], 0x0000)

	// Question section
	// Domain: google.com (encoded as length-prefixed labels)
	offset := 12
	// "google"
	query[offset] = 6
	offset++
	copy(query[offset:], "google")
	offset += 6
	// "com"
	query[offset] = 3
	offset++
	copy(query[offset:], "com")
	offset += 3
	// Root (null terminator)
	query[offset] = 0
	offset++

	// Type: A (1)
	binary.BigEndian.PutUint16(query[offset:offset+2], 0x0001)
	offset += 2
	// Class: IN (1)
	binary.BigEndian.PutUint16(query[offset:offset+2], 0x0001)
	offset += 2

	return query[:offset]
}

func TestUDPTest_Integration_GoogleDNS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &UDPTest{}
	dnsQuery := string(createDNSQuery())

	config := TestConfig{
		Name:    "Google Public DNS",
		Address: "8.8.8.8:53",
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     dnsQuery,
			WaitResponse: true,
			ResponseSize: 512,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Logf("DNS query failed (expected if no internet): %v", err)
		// Don't fail the test if there's no internet connectivity
		if result.Status == TestStatusTimeout || result.Status == TestStatusFailed {
			t.Skip("Skipping DNS test - no internet connectivity or DNS blocked")
		}
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Protocol != "UDP" {
		t.Errorf("Execute() protocol = %v, want UDP", result.Protocol)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	if result.ResponseSize == 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}

	t.Logf("DNS query successful: latency=%v, response_size=%d bytes", result.Latency, result.ResponseSize)
}

func TestUDPTest_Integration_CloudflareDNS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &UDPTest{}
	dnsQuery := string(createDNSQuery())

	config := TestConfig{
		Name:    "Cloudflare DNS",
		Address: "1.1.1.1:53",
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     dnsQuery,
			WaitResponse: true,
			ResponseSize: 512,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Logf("DNS query failed (expected if no internet): %v", err)
		if result.Status == TestStatusTimeout || result.Status == TestStatusFailed {
			t.Skip("Skipping DNS test - no internet connectivity or DNS blocked")
		}
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	if result.ResponseSize == 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}

	t.Logf("DNS query successful: latency=%v, response_size=%d bytes", result.Latency, result.ResponseSize)
}

func TestUDPTest_Integration_LocalhostDNS(t *testing.T) {
	test := &UDPTest{}
	dnsQuery := string(createDNSQuery())

	config := TestConfig{
		Name:    "Localhost DNS",
		Address: "127.0.0.1:53",
		Timeout: 2 * time.Second,
		Config: &UDPConfig{
			SendData:     dnsQuery,
			WaitResponse: true,
			ResponseSize: 512,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// This may fail if there's no local DNS server running
	// That's expected, so we just verify the test behaves correctly
	if err != nil {
		if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
			t.Errorf("Execute() status = %v, want timeout or failed when no local DNS", result.Status)
		}
		t.Logf("No local DNS server (expected): %v", err)
		return
	}

	// If there is a local DNS server, verify success
	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}
}

func TestUDPTest_Integration_ClosedPort(t *testing.T) {
	test := &UDPTest{}
	config := TestConfig{
		Name:    "Closed Port Test",
		Address: "127.0.0.1:9",
		Timeout: 1 * time.Second,
		Config: &UDPConfig{
			SendData:     "test",
			WaitResponse: true,
			ResponseSize: 512,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// Should timeout or fail (port likely closed)
	if err == nil {
		t.Error("Execute() expected error for closed port")
	}

	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want timeout or failed", result.Status)
	}
}

func TestUDPTest_Integration_UnreachableHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Unreachable Host",
		Address: "192.0.2.1:53", // TEST-NET-1, reserved for documentation
		Timeout: 2 * time.Second,
		Config: &UDPConfig{
			SendData:     "test",
			WaitResponse: true,
			ResponseSize: 512,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// Should timeout or fail
	if err == nil {
		t.Error("Execute() expected error for unreachable host")
	}

	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want timeout or failed", result.Status)
	}
}

func TestUDPTest_Integration_IPv6DNS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &UDPTest{}
	dnsQuery := string(createDNSQuery())

	config := TestConfig{
		Name:    "Google DNS IPv6",
		Address: "[2001:4860:4860::8888]:53",
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     dnsQuery,
			WaitResponse: true,
			ResponseSize: 512,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		// IPv6 might not be available or DNS might be blocked
		if result.Status == TestStatusTimeout || result.Status == TestStatusFailed {
			t.Skip("Skipping IPv6 DNS test - no IPv6 connectivity or DNS blocked")
		}
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	if result.ResponseSize == 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}

	t.Logf("IPv6 DNS query successful: latency=%v, response_size=%d bytes", result.Latency, result.ResponseSize)
}

func TestUDPTest_Integration_SendOnlyMode(t *testing.T) {
	test := &UDPTest{}
	config := TestConfig{
		Name:    "Send Only Integration",
		Address: "8.8.8.8:53",
		Timeout: 2 * time.Second,
		Config: &UDPConfig{
			SendData:     "test data",
			WaitResponse: false, // Don't wait for response
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// Should succeed immediately without waiting
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	if result.Protocol != "UDP" {
		t.Errorf("Execute() protocol = %v, want UDP", result.Protocol)
	}

	// Latency should be very small since we're not waiting
	if result.Latency > 1*time.Second {
		t.Errorf("Execute() latency = %v, expected < 1s for send-only", result.Latency)
	}
}

func TestUDPTest_Integration_MultipleQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &UDPTest{}
	dnsQuery := string(createDNSQuery())

	// Test multiple queries in sequence
	for i := 0; i < 5; i++ {
		config := TestConfig{
			Name:    "Google DNS Multiple Queries",
			Address: "8.8.8.8:53",
			Timeout: 5 * time.Second,
			Config: &UDPConfig{
				SendData:     dnsQuery,
				WaitResponse: true,
				ResponseSize: 512,
			},
		}

		ctx := context.Background()
		result, err := test.Execute(ctx, config)

		if err != nil {
			// Skip if network unavailable
			if result.Status == TestStatusTimeout || result.Status == TestStatusFailed {
				t.Skip("Skipping test - no internet connectivity")
			}
			t.Fatalf("Query %d failed: %v", i+1, err)
		}

		if result.Status != TestStatusSuccess {
			t.Errorf("Query %d: status = %v, want %v", i+1, result.Status, TestStatusSuccess)
		}

		t.Logf("Query %d successful: latency=%v", i+1, result.Latency)
	}
}

func TestUDPTest_Integration_LatencyMeasurement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &UDPTest{}
	dnsQuery := string(createDNSQuery())

	config := TestConfig{
		Name:    "DNS Latency Test",
		Address: "8.8.8.8:53",
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     dnsQuery,
			WaitResponse: true,
			ResponseSize: 512,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		if result.Status == TestStatusTimeout || result.Status == TestStatusFailed {
			t.Skip("Skipping test - no internet connectivity")
		}
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	// Verify latency is reasonable (should be < 1 second for DNS)
	if result.Latency > 1*time.Second {
		t.Errorf("Execute() latency = %v, expected < 1s for DNS query", result.Latency)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, expected > 0", result.Latency)
	}

	t.Logf("DNS latency: %v", result.Latency)
}
