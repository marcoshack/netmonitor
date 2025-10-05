package main

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoshack/netmonitor/internal/network"
)

func main() {
	// Create an HTTP test instance
	httpTest := &network.HTTPTest{}

	// Example 1: Simple HTTP test
	fmt.Println("=== HTTP Test Example ===")
	config1 := network.TestConfig{
		Name:     "Cloudflare DNS",
		Address:  "https://1.1.1.1",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
	}

	ctx := context.Background()
	result1, err := httpTest.Execute(ctx, config1)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("HTTP Test Results:\n")
		fmt.Printf("  Endpoint: %s\n", config1.Address)
		fmt.Printf("  Status: %s\n", result1.Status)
		fmt.Printf("  LatencyInMs: %.2f\n", float64(result1.Latency.Nanoseconds())/1_000_000.0)
		fmt.Printf("  Response Size: %d bytes\n", result1.ResponseSize)
		fmt.Printf("  Protocol: %s\n", result1.Protocol)
		fmt.Printf("  Timestamp: %v\n", result1.Timestamp)
		if result1.Error != "" {
			fmt.Printf("  Error: %s\n", result1.Error)
		}
	}

	// Example 2: HTTP test with custom configuration
	fmt.Println("\n=== HTTP Test with Custom Configuration ===")
	config2 := network.TestConfig{
		Name:     "Google Homepage",
		Address:  "https://www.google.com",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
		Config: &network.HTTPConfig{
			Method:          "GET",
			FollowRedirects: true,
			ValidateSSL:     true,
			Headers: map[string]string{
				"X-Custom-Header": "MyValue",
			},
		},
	}

	result2, err := httpTest.Execute(ctx, config2)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("HTTP Test Results:\n")
		fmt.Printf("  Endpoint: %s\n", config2.Address)
		fmt.Printf("  Status: %s\n", result2.Status)
		fmt.Printf("  LatencyInMs: %.2f\n", float64(result2.Latency.Nanoseconds())/1_000_000.0)
		fmt.Printf("  Response Size: %d bytes\n", result2.ResponseSize)
		if result2.Error != "" {
			fmt.Printf("  Error: %s\n", result2.Error)
		}
	}

	// Example 3: Testing for specific status code
	fmt.Println("\n=== HTTP Test Expecting 404 ===")
	config3 := network.TestConfig{
		Name:     "Not Found Test",
		Address:  "https://httpbin.org/status/404",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
		Config: &network.HTTPConfig{
			ExpectedStatus: 404, // Expect a 404 response
		},
	}

	result3, err := httpTest.Execute(ctx, config3)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("HTTP Test Results:\n")
		fmt.Printf("  Endpoint: %s\n", config3.Address)
		fmt.Printf("  Status: %s\n", result3.Status)
		fmt.Printf("  LatencyInMs: %.2f\n", float64(result3.Latency.Nanoseconds())/1_000_000.0)
		if result3.Error != "" {
			fmt.Printf("  Error: %s\n", result3.Error)
		}
	}
}
