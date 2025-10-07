package main

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoshack/netmonitor/internal/network"
)

func main() {
	// Create a TCP test instance
	tcpTest := &network.TCPTest{}

	// Example 1: Simple TCP connection test
	fmt.Println("=== TCP Connection Test Example ===")
	config1 := network.TestConfig{
		Name:     "Google DNS",
		Address:  "8.8.8.8:53",
		Timeout:  5 * time.Second,
		Protocol: "TCP",
	}

	ctx := context.Background()
	result1, err := tcpTest.Execute(ctx, config1)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TCP Test Results:\n")
		fmt.Printf("  Endpoint: %s\n", config1.Address)
		fmt.Printf("  Status: %s\n", result1.Status)
		fmt.Printf("  LatencyInMs: %.2f\n", float64(result1.Latency.Nanoseconds())/1_000_000.0)
		fmt.Printf("  Protocol: %s\n", result1.Protocol)
		fmt.Printf("  Timestamp: %v\n", result1.Timestamp)
		if result1.Error != "" {
			fmt.Printf("  Error: %s\n", result1.Error)
		}
	}

	// Example 2: TCP connection test to HTTPS port
	fmt.Println("\n=== TCP Test to HTTPS Port ===")
	config2 := network.TestConfig{
		Name:     "Cloudflare HTTPS",
		Address:  "1.1.1.1:443",
		Timeout:  5 * time.Second,
		Protocol: "TCP",
	}

	result2, err := tcpTest.Execute(ctx, config2)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TCP Test Results:\n")
		fmt.Printf("  Endpoint: %s\n", config2.Address)
		fmt.Printf("  Status: %s\n", result2.Status)
		fmt.Printf("  LatencyInMs: %.2f\n", float64(result2.Latency.Nanoseconds())/1_000_000.0)
		if result2.Error != "" {
			fmt.Printf("  Error: %s\n", result2.Error)
		}
	}

	// Example 3: TCP test with data exchange (HTTP request)
	fmt.Println("\n=== TCP Test with HTTP Request ===")
	config3 := network.TestConfig{
		Name:     "HTTP over TCP",
		Address:  "www.google.com:80",
		Timeout:  10 * time.Second,
		Protocol: "TCP",
		Config: &network.TCPConfig{
			SendData:       "GET / HTTP/1.0\r\nHost: www.google.com\r\n\r\n",
			ExpectResponse: true,
			ExpectedData:   "HTTP/",
		},
	}

	result3, err := tcpTest.Execute(ctx, config3)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TCP Test Results:\n")
		fmt.Printf("  Endpoint: %s\n", config3.Address)
		fmt.Printf("  Status: %s\n", result3.Status)
		fmt.Printf("  LatencyInMs: %.2f\n", float64(result3.Latency.Nanoseconds())/1_000_000.0)
		fmt.Printf("  Response Size: %d bytes\n", result3.ResponseSize)
		if result3.Error != "" {
			fmt.Printf("  Error: %s\n", result3.Error)
		}
	}

	// Example 4: Testing common database ports
	fmt.Println("\n=== Testing Common Database Ports ===")
	databasePorts := []struct {
		name    string
		address string
	}{
		{"PostgreSQL", "localhost:5432"},
		{"MySQL", "localhost:3306"},
		{"Redis", "localhost:6379"},
		{"MongoDB", "localhost:27017"},
	}

	for _, db := range databasePorts {
		config := network.TestConfig{
			Name:     db.name,
			Address:  db.address,
			Timeout:  1 * time.Second,
			Protocol: "TCP",
		}

		result, err := tcpTest.Execute(ctx, config)
		status := "✗"
		if result.Status == network.TestStatusSuccess {
			status = "✓"
		}

		fmt.Printf("  %s %s (%s): %s", status, db.name, db.address, result.Status)
		if err != nil {
			fmt.Printf(" - %s", result.Error)
		} else if result.Status == network.TestStatusSuccess {
			fmt.Printf(" - %.2fms", float64(result.Latency.Nanoseconds())/1_000_000.0)
		}
		fmt.Println()
	}

	// Example 5: Testing a closed port (should fail)
	fmt.Println("\n=== Testing Closed Port ===")
	config5 := network.TestConfig{
		Name:     "Closed Port Test",
		Address:  "localhost:54321",
		Timeout:  2 * time.Second,
		Protocol: "TCP",
	}

	result5, err := tcpTest.Execute(ctx, config5)

	fmt.Printf("TCP Test Results:\n")
	fmt.Printf("  Endpoint: %s\n", config5.Address)
	fmt.Printf("  Status: %s\n", result5.Status)
	if err != nil {
		fmt.Printf("  Error: %s (This is expected for a closed port)\n", result5.Error)
	}

	// Example 6: IPv6 connection test
	fmt.Println("\n=== IPv6 Connection Test ===")
	config6 := network.TestConfig{
		Name:     "IPv6 Localhost",
		Address:  "[::1]:80",
		Timeout:  2 * time.Second,
		Protocol: "TCP",
	}

	result6, err := tcpTest.Execute(ctx, config6)

	fmt.Printf("TCP Test Results:\n")
	fmt.Printf("  Endpoint: %s\n", config6.Address)
	fmt.Printf("  Status: %s\n", result6.Status)
	if err != nil {
		fmt.Printf("  Error: %s\n", result6.Error)
	} else if result6.Status == network.TestStatusSuccess {
		fmt.Printf("  LatencyInMs: %.2f\n", float64(result6.Latency.Nanoseconds())/1_000_000.0)
	}
}
