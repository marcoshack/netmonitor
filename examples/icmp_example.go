package main

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoshack/netmonitor/internal/network"
)

func main() {
	// Create an ICMP test instance
	icmpTest := &network.ICMPTest{}

	// Configure the test
	config := network.TestConfig{
		Name:     "Google DNS",
		Address:  "8.8.8.8",
		Timeout:  5 * time.Second,
		Protocol: "ICMP",
		Config: &network.ICMPConfig{
			Count:      1,
			PacketSize: 64,
			TTL:        64,
			Privileged: false, // Set to true if you have elevated privileges
		},
	}

	// Execute the ping test
	ctx := context.Background()
	result, err := icmpTest.Execute(ctx, config)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Display the results
	fmt.Printf("Ping Test Results:\n")
	fmt.Printf("  Endpoint: %s\n", config.Address)
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Printf("  LatencyInMs: %.2f\n", float64(result.Latency.Nanoseconds())/1_000_000.0)
	fmt.Printf("  Protocol: %s\n", result.Protocol)
	fmt.Printf("  Timestamp: %v\n", result.Timestamp)

	if result.Error != "" {
		fmt.Printf("  Error: %s\n", result.Error)
	}
}