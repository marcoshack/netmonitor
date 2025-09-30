package network

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
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

	// Validate that the address is a valid IP or can be resolved
	if net.ParseIP(config.Address) == nil {
		// Try to resolve hostname
		_, err := net.ResolveIPAddr("ip", config.Address)
		if err != nil {
			return fmt.Errorf("invalid address: %w", err)
		}
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

	// Resolve the address
	ipAddr, err := net.ResolveIPAddr("ip", config.Address)
	if err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "ICMP",
			Status:    TestStatusError,
			Error:     fmt.Sprintf("failed to resolve address: %v", err),
		}, err
	}

	// Determine if IPv4 or IPv6
	isIPv4 := ipAddr.IP.To4() != nil

	// Perform the ping
	latency, err := t.ping(ctx, ipAddr, isIPv4, config.Timeout, icmpConfig)

	result := &TestResult{
		Timestamp:  startTime,
		EndpointID: config.Name,
		Protocol:   "ICMP",
		Latency:    latency,
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = TestStatusTimeout
			result.Error = "ping timeout"
		} else {
			result.Status = TestStatusFailed
			result.Error = err.Error()
		}
		return result, err
	}

	result.Status = TestStatusSuccess
	return result, nil
}

// ping performs the actual ICMP echo request/reply
func (t *ICMPTest) ping(ctx context.Context, addr *net.IPAddr, isIPv4 bool, timeout time.Duration, config *ICMPConfig) (time.Duration, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var network string
	var icmpType icmp.Type

	if isIPv4 {
		if config.Privileged {
			network = "ip4:icmp"
		} else {
			network = "udp4"
		}
		icmpType = ipv4.ICMPTypeEcho
	} else {
		if config.Privileged {
			network = "ip6:ipv6-icmp"
		} else {
			network = "udp6"
		}
		icmpType = ipv6.ICMPTypeEchoRequest
	}

	// Create connection
	conn, err := icmp.ListenPacket(network, "")
	if err != nil {
		return 0, fmt.Errorf("failed to create ICMP connection: %w", err)
	}
	defer conn.Close()

	// Set TTL if specified
	if config.TTL > 0 {
		if isIPv4 {
			if ipv4Conn := conn.IPv4PacketConn(); ipv4Conn != nil {
				ipv4Conn.SetTTL(config.TTL)
			}
		} else {
			if ipv6Conn := conn.IPv6PacketConn(); ipv6Conn != nil {
				ipv6Conn.SetHopLimit(config.TTL)
			}
		}
	}

	// Prepare ICMP message
	msg := icmp.Message{
		Type: icmpType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1,
			Seq:  1,
			Data: make([]byte, config.PacketSize),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal ICMP message: %w", err)
	}

	// Send echo request and measure latency
	start := time.Now()

	if _, err := conn.WriteTo(msgBytes, addr); err != nil {
		return 0, fmt.Errorf("failed to send ICMP echo request: %w", err)
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Wait for reply
	reply := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			n, peer, err := conn.ReadFrom(reply)
			if err != nil {
				return 0, fmt.Errorf("failed to read ICMP reply: %w", err)
			}

			// Measure round-trip time
			rtt := time.Since(start)

			// Parse the reply
			var proto int
			if isIPv4 {
				proto = 1 // ICMP for IPv4
			} else {
				proto = 58 // ICMPv6
			}

			replyMsg, err := icmp.ParseMessage(proto, reply[:n])
			if err != nil {
				continue // Not a valid ICMP message, keep waiting
			}

			// Check if this is an echo reply
			isEchoReply := (isIPv4 && replyMsg.Type == ipv4.ICMPTypeEchoReply) ||
				(!isIPv4 && replyMsg.Type == ipv6.ICMPTypeEchoReply)

			if !isEchoReply {
				continue // Not an echo reply, keep waiting
			}

			// Verify it's from the expected peer
			if peer.String() != addr.String() {
				continue // Reply from different host, keep waiting
			}

			// Successfully received echo reply
			return rtt, nil
		}
	}
}