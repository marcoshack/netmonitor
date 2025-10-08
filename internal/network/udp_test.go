package network

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestUDPTest_GetProtocol(t *testing.T) {
	test := &UDPTest{}
	if got := test.GetProtocol(); got != "UDP" {
		t.Errorf("GetProtocol() = %v, want UDP", got)
	}
}

func TestUDPTest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TestConfig
		wantErr bool
	}{
		{
			name: "valid IPv4 address with port",
			config: TestConfig{
				Address: "192.168.1.1:53",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid hostname with port",
			config: TestConfig{
				Address: "dns.google.com:53",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid IPv6 address with port",
			config: TestConfig{
				Address: "[::1]:5353",
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
				Address: "example.com:53",
				Timeout: 0,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: TestConfig{
				Address: "example.com:53",
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "valid with UDP config",
			config: TestConfig{
				Address: "example.com:53",
				Timeout: 5 * time.Second,
				Config: &UDPConfig{
					Port:         53,
					SendData:     "test",
					WaitResponse: true,
					ResponseSize: 512,
				},
			},
			wantErr: false,
		},
		{
			name: "port mismatch in config",
			config: TestConfig{
				Address: "example.com:53",
				Timeout: 5 * time.Second,
				Config: &UDPConfig{
					Port: 5353, // Doesn't match address port
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config type",
			config: TestConfig{
				Address: "example.com:53",
				Timeout: 5 * time.Second,
				Config:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative response size",
			config: TestConfig{
				Address: "example.com:53",
				Timeout: 5 * time.Second,
				Config: &UDPConfig{
					ResponseSize: -1,
				},
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
			test := &UDPTest{}
			err := test.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUDPTest_Execute_Success(t *testing.T) {
	// Create a test UDP server
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer conn.Close()

	// Get the actual address the server is listening on
	serverAddr := conn.LocalAddr().String()

	// Echo server - respond to incoming packets
	go func() {
		buffer := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		// Echo back the data
		conn.WriteToUDP(buffer[:n], addr)
	}()

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Test UDP Echo Server",
		Address: serverAddr,
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     "Hello UDP!",
			WaitResponse: true,
			ResponseSize: 1024,
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

	if result.Protocol != "UDP" {
		t.Errorf("Execute() protocol = %v, want UDP", result.Protocol)
	}

	if result.Latency < 0 {
		t.Errorf("Execute() latency = %v, want >= 0", result.Latency)
	}

	if result.ResponseSize == 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}

	if result.EndpointID != "Test UDP Echo Server" {
		t.Errorf("Execute() endpoint ID = %v, want Test UDP Echo Server", result.EndpointID)
	}
}

func TestUDPTest_Execute_NoResponse(t *testing.T) {
	test := &UDPTest{}
	// Send to a likely closed port
	config := TestConfig{
		Name:    "No Response Test",
		Address: "127.0.0.1:54321",
		Timeout: 500 * time.Millisecond,
		Config: &UDPConfig{
			SendData:     "Hello",
			WaitResponse: true,
			ResponseSize: 1024,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err == nil {
		t.Error("Execute() expected error for no response, got nil")
	}

	// Should timeout since no response expected
	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want timeout or failed", result.Status)
	}
}

func TestUDPTest_Execute_SendOnly(t *testing.T) {
	test := &UDPTest{}
	config := TestConfig{
		Name:    "Send Only Test",
		Address: "127.0.0.1:54322",
		Timeout: 1 * time.Second,
		Config: &UDPConfig{
			SendData:     "PROBE",
			WaitResponse: false, // Don't wait for response
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// Should succeed even without response
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	if result.Protocol != "UDP" {
		t.Errorf("Execute() protocol = %v, want UDP", result.Protocol)
	}
}

func TestUDPTest_Execute_DefaultPayload(t *testing.T) {
	// Create a test UDP server
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()
	receivedData := make(chan string, 1)

	// Server to capture received data
	go func() {
		buffer := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		receivedData <- string(buffer[:n])
		conn.WriteToUDP([]byte("ACK"), addr)
	}()

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Default Payload Test",
		Address: serverAddr,
		Timeout: 2 * time.Second,
		Config: &UDPConfig{
			// No SendData specified - should use default "PROBE"
			WaitResponse: true,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	// Verify default payload was sent
	select {
	case data := <-receivedData:
		if data != "PROBE" {
			t.Errorf("Received data = %v, want PROBE", data)
		}
	case <-time.After(1 * time.Second):
		t.Error("Server did not receive data")
	}
}

func TestUDPTest_Execute_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	// Create a server that receives but doesn't respond
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()

	// Server that receives but doesn't respond
	go func() {
		buffer := make([]byte, 1024)
		conn.ReadFromUDP(buffer)
		// Don't send response
	}()

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Timeout Test",
		Address: serverAddr,
		Timeout: 100 * time.Millisecond,
		Config: &UDPConfig{
			SendData:     "test",
			WaitResponse: true,
		},
	}

	ctx := context.Background()
	start := time.Now()
	result, err := test.Execute(ctx, config)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Execute() expected timeout error, got nil")
	}

	if result.Status != TestStatusTimeout {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusTimeout)
	}

	// Should respect timeout (with some margin for processing)
	if elapsed > 2*time.Second {
		t.Errorf("Execute() took %v, expected quick timeout", elapsed)
	}
}

func TestUDPTest_Execute_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cancellation test in short mode")
	}

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Cancellation Test",
		Address: "192.0.2.1:53", // TEST-NET-1, should be unreachable
		Timeout: 30 * time.Second,
		Config: &UDPConfig{
			WaitResponse: true,
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
		t.Error("Execute() expected cancellation error, got nil")
	}

	// Should complete quickly due to cancellation
	if elapsed > 2*time.Second {
		t.Errorf("Execute() took %v, expected quick cancellation", elapsed)
	}

	if result.Status != TestStatusTimeout {
		t.Errorf("Execute() status = %v, want timeout", result.Status)
	}
}

func TestUDPTest_Execute_ValidationErrors(t *testing.T) {
	test := &UDPTest{}
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
				Address: "example.com:53",
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

func TestUDPTest_Execute_IPv6(t *testing.T) {
	// Create a test UDP server on IPv6
	udpAddr, err := net.ResolveUDPAddr("udp6", "[::1]:0")
	if err != nil {
		t.Skip("IPv6 not available, skipping test")
	}

	conn, err := net.ListenUDP("udp6", udpAddr)
	if err != nil {
		t.Skip("IPv6 not available, skipping test")
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()

	// Echo server
	go func() {
		buffer := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		conn.WriteToUDP(buffer[:n], addr)
	}()

	test := &UDPTest{}
	config := TestConfig{
		Name:    "IPv6 Test",
		Address: serverAddr,
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     "IPv6 test",
			WaitResponse: true,
		},
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

func TestUDPTest_Execute_MultiplePackets(t *testing.T) {
	// Create a test UDP server
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()
	packetCount := 0

	// Echo server that counts packets
	go func() {
		buffer := make([]byte, 1024)
		for i := 0; i < 3; i++ {
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			packetCount++
			conn.WriteToUDP(buffer[:n], addr)
		}
	}()

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Multiple Packets Test",
		Address: serverAddr,
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     "test",
			WaitResponse: true,
		},
	}

	ctx := context.Background()

	// Send 3 packets
	for i := 0; i < 3; i++ {
		result, err := test.Execute(ctx, config)
		if err != nil {
			t.Fatalf("Execute() iteration %d unexpected error: %v", i, err)
		}
		if result.Status != TestStatusSuccess {
			t.Errorf("Execute() iteration %d status = %v, want %v", i, result.Status, TestStatusSuccess)
		}
	}

	// Give time for all packets to be processed
	time.Sleep(100 * time.Millisecond)

	if packetCount != 3 {
		t.Errorf("Expected 3 packets, got %d", packetCount)
	}
}

func TestUDPTest_Execute_LargePayload(t *testing.T) {
	// Create a test UDP server
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()

	// Echo server
	go func() {
		buffer := make([]byte, 65536) // Max UDP packet size
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		conn.WriteToUDP(buffer[:n], addr)
	}()

	// Create a large payload (but within UDP limits)
	largeData := string(make([]byte, 8192))

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Large Payload Test",
		Address: serverAddr,
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     largeData,
			WaitResponse: true,
			ResponseSize: 65536,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	if result.ResponseSize != int64(len(largeData)) {
		t.Errorf("Execute() response size = %v, want %v", result.ResponseSize, len(largeData))
	}
}

func TestUDPTest_Execute_PortRange(t *testing.T) {
	test := &UDPTest{}
	ctx := context.Background()

	// Test common UDP port numbers
	ports := []int{53, 123, 161, 514, 5353}

	for _, port := range ports {
		config := TestConfig{
			Name:    fmt.Sprintf("Port %d Test", port),
			Address: fmt.Sprintf("127.0.0.1:%d", port),
			Timeout: 100 * time.Millisecond,
			Config: &UDPConfig{
				WaitResponse: false, // Just send, don't wait
			},
		}

		result, _ := test.Execute(ctx, config)

		// We expect these to succeed (send only, no response expected)
		if result.Protocol != "UDP" {
			t.Errorf("Port %d: protocol = %v, want UDP", port, result.Protocol)
		}

		if result.Status != TestStatusSuccess {
			t.Errorf("Port %d: status = %v, want success (send only mode)", port, result.Status)
		}
	}
}

func TestUDPTest_Execute_CustomResponseSize(t *testing.T) {
	// Create a test UDP server
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()

	// Server that sends a specific size response
	go func() {
		buffer := make([]byte, 1024)
		_, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		// Send 100 bytes
		conn.WriteToUDP(make([]byte, 100), addr)
	}()

	test := &UDPTest{}
	config := TestConfig{
		Name:    "Custom Response Size Test",
		Address: serverAddr,
		Timeout: 5 * time.Second,
		Config: &UDPConfig{
			SendData:     "test",
			WaitResponse: true,
			ResponseSize: 512, // Custom buffer size
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	if result.ResponseSize != 100 {
		t.Errorf("Execute() response size = %v, want 100", result.ResponseSize)
	}
}
