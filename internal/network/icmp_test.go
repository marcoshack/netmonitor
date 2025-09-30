package network

import (
	"context"
	"testing"
	"time"
)

func TestICMPTest_GetProtocol(t *testing.T) {
	test := &ICMPTest{}
	if got := test.GetProtocol(); got != "ICMP" {
		t.Errorf("GetProtocol() = %v, want ICMP", got)
	}
}

func TestICMPTest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TestConfig
		wantErr bool
	}{
		{
			name: "valid IPv4 address",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid IPv6 address",
			config: TestConfig{
				Address: "2001:4860:4860::8888",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid hostname",
			config: TestConfig{
				Address: "localhost",
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
			name: "invalid address",
			config: TestConfig{
				Address: "invalid..address",
				Timeout: 5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 0,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "valid with ICMP config",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
				Config: &ICMPConfig{
					Count:      3,
					PacketSize: 64,
					TTL:        64,
				},
			},
			wantErr: false,
		},
		{
			name: "negative count",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
				Config: &ICMPConfig{
					Count:      -1,
					PacketSize: 64,
					TTL:        64,
				},
			},
			wantErr: true,
		},
		{
			name: "packet size too large",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
				Config: &ICMPConfig{
					Count:      1,
					PacketSize: 70000,
					TTL:        64,
				},
			},
			wantErr: true,
		},
		{
			name: "negative packet size",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
				Config: &ICMPConfig{
					Count:      1,
					PacketSize: -10,
					TTL:        64,
				},
			},
			wantErr: true,
		},
		{
			name: "TTL too large",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
				Config: &ICMPConfig{
					Count:      1,
					PacketSize: 64,
					TTL:        300,
				},
			},
			wantErr: true,
		},
		{
			name: "negative TTL",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
				Config: &ICMPConfig{
					Count:      1,
					PacketSize: 64,
					TTL:        -5,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config type",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 5 * time.Second,
				Config:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := &ICMPTest{}
			err := test.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestICMPTest_Execute_ValidationErrors(t *testing.T) {
	test := &ICMPTest{}
	ctx := context.Background()

	tests := []struct {
		name         string
		config       TestConfig
		wantStatus   TestStatus
		wantProtocol string
	}{
		{
			name: "empty address",
			config: TestConfig{
				Address: "",
				Timeout: 5 * time.Second,
			},
			wantStatus:   TestStatusError,
			wantProtocol: "ICMP",
		},
		{
			name: "invalid address",
			config: TestConfig{
				Address: "not-a-valid-address...",
				Timeout: 5 * time.Second,
			},
			wantStatus:   TestStatusError,
			wantProtocol: "ICMP",
		},
		{
			name: "zero timeout",
			config: TestConfig{
				Address: "8.8.8.8",
				Timeout: 0,
			},
			wantStatus:   TestStatusError,
			wantProtocol: "ICMP",
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
			if result.Protocol != tt.wantProtocol {
				t.Errorf("Execute() protocol = %v, want %v", result.Protocol, tt.wantProtocol)
			}
		})
	}
}

func TestICMPTest_Execute_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	test := &ICMPTest{}

	// Use a very short timeout and an unreachable IP
	config := TestConfig{
		Name:    "Timeout Test",
		Address: "192.0.2.1", // TEST-NET-1, should be unreachable
		Timeout: 100 * time.Millisecond,
		Config: &ICMPConfig{
			Count:      1,
			PacketSize: 64,
			TTL:        64,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err == nil {
		t.Error("Execute() expected error for timeout, got nil")
	}

	// Should be either timeout or failed (depending on network configuration)
	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v or %v", result.Status, TestStatusTimeout, TestStatusFailed)
	}

	if result.Protocol != "ICMP" {
		t.Errorf("Execute() protocol = %v, want ICMP", result.Protocol)
	}
}

func TestICMPTest_Execute_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cancellation test in short mode")
	}

	test := &ICMPTest{}

	config := TestConfig{
		Name:    "Cancel Test",
		Address: "192.0.2.1", // TEST-NET-1, should be unreachable
		Timeout: 10 * time.Second,
		Config: &ICMPConfig{
			Count:      1,
			PacketSize: 64,
			TTL:        64,
		},
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
		t.Error("Execute() expected error for cancellation, got nil")
	}

	// Should complete quickly (within 1 second, not the full 10 second timeout)
	if elapsed > 2*time.Second {
		t.Errorf("Execute() took %v, expected quick cancellation", elapsed)
	}

	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v or %v", result.Status, TestStatusTimeout, TestStatusFailed)
	}
}

func TestICMPTest_Execute_DefaultConfig(t *testing.T) {
	test := &ICMPTest{}

	config := TestConfig{
		Name:    "Default Config Test",
		Address: "127.0.0.1",
		Timeout: 5 * time.Second,
		// No Config provided - should use defaults
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// This test might fail depending on permissions and system configuration
	// We're mainly testing that it doesn't panic with default config
	if result == nil {
		t.Fatal("Execute() returned nil result")
	}

	if result.Protocol != "ICMP" {
		t.Errorf("Execute() protocol = %v, want ICMP", result.Protocol)
	}

	// Either success or permission error is acceptable
	if result.Status == TestStatusError && err != nil {
		// Permission error is acceptable
		t.Logf("Execute() failed with error: %v (this is acceptable for unprivileged tests)", err)
	}
}