package network

import (
	"context"
	"testing"
	"time"
)

// TestHTTPTest_Integration_HTTP tests HTTP connectivity with a real HTTP endpoint
func TestHTTPTest_Integration_HTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "HTTP Test - Example.com",
		Address:  "http://example.com",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	if result.ResponseSize <= 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}

	if result.Protocol != "HTTP" {
		t.Errorf("Execute() protocol = %v, want HTTP", result.Protocol)
	}

	t.Logf("HTTP test to %s: latencyInMs=%.2f, status=%v, size=%d bytes",
		config.Address, float64(result.Latency.Nanoseconds())/1_000_000.0, result.Status, result.ResponseSize)
}

// TestHTTPTest_Integration_HTTPS tests HTTPS connectivity with SSL validation
func TestHTTPTest_Integration_HTTPS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "HTTPS Test - Cloudflare DNS",
		Address:  "https://1.1.1.1",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
		Config: &HTTPConfig{
			ValidateSSL:     true,
			FollowRedirects: true,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	if result.ResponseSize <= 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}

	t.Logf("HTTPS test to %s: latencyInMs=%.2f, status=%v, size=%d bytes",
		config.Address, float64(result.Latency.Nanoseconds())/1_000_000.0, result.Status, result.ResponseSize)
}

// TestHTTPTest_Integration_HTTPS_Google tests HTTPS with another real endpoint
func TestHTTPTest_Integration_HTTPS_Google(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "HTTPS Test - Google",
		Address:  "https://www.google.com",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
		Config: &HTTPConfig{
			ValidateSSL:     true,
			FollowRedirects: true,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	if result.ResponseSize <= 0 {
		t.Errorf("Execute() response size = %v, want > 0", result.ResponseSize)
	}

	t.Logf("HTTPS test to %s: latencyInMs=%.2f, status=%v, size=%d bytes",
		config.Address, float64(result.Latency.Nanoseconds())/1_000_000.0, result.Status, result.ResponseSize)
}

// TestHTTPTest_Integration_404NotFound tests handling of 404 status
func TestHTTPTest_Integration_404NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "404 Test",
		Address:  "https://httpbin.org/status/404",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should fail because 404 is not a 2xx status
	if result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusFailed)
	}

	if result.Error == "" {
		t.Error("Execute() error message is empty, want error description")
	}

	t.Logf("404 test: status=%v, error=%v", result.Status, result.Error)
}

// TestHTTPTest_Integration_ExpectedStatus tests expecting specific status codes
func TestHTTPTest_Integration_ExpectedStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "Expected 404 Test",
		Address:  "https://httpbin.org/status/404",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
		Config: &HTTPConfig{
			ExpectedStatus: 404,
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should succeed because we expect 404
	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	t.Logf("Expected 404 test: status=%v", result.Status)
}

// TestHTTPTest_Integration_UnreachableHost tests behavior with unreachable host
func TestHTTPTest_Integration_UnreachableHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	// Using TEST-NET-1 address range which should be unreachable
	config := TestConfig{
		Name:     "Unreachable Host",
		Address:  "http://192.0.2.1",
		Timeout:  3 * time.Second,
		Protocol: "HTTP",
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

	t.Logf("Unreachable host test: status=%v, error=%v", result.Status, result.Error)
}

// TestHTTPTest_Integration_InvalidDomain tests behavior with invalid domain
func TestHTTPTest_Integration_InvalidDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "Invalid Domain",
		Address:  "http://this-domain-definitely-does-not-exist-12345.invalid",
		Timeout:  5 * time.Second,
		Protocol: "HTTP",
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// Should fail
	if err == nil {
		t.Error("Execute() expected error for invalid domain, got nil")
	}

	if result.Status != TestStatusFailed && result.Status != TestStatusTimeout {
		t.Errorf("Execute() status = %v, want %v or %v", result.Status, TestStatusFailed, TestStatusTimeout)
	}

	t.Logf("Invalid domain test: status=%v, error=%v", result.Status, result.Error)
}

// TestHTTPTest_Integration_Redirects tests redirect handling
func TestHTTPTest_Integration_Redirects(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("follow redirects", func(t *testing.T) {
		test := &HTTPTest{}

		config := TestConfig{
			Name:     "Redirect Test - Follow",
			Address:  "http://httpbin.org/redirect/2",
			Timeout:  10 * time.Second,
			Protocol: "HTTP",
			Config: &HTTPConfig{
				FollowRedirects: true,
			},
		}

		ctx := context.Background()
		result, err := test.Execute(ctx, config)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.Status != TestStatusSuccess {
			t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
		}

		t.Logf("Follow redirects test: status=%v, latencyInMs=%.2f",
			result.Status, float64(result.Latency.Nanoseconds())/1_000_000.0)
	})

	t.Run("don't follow redirects", func(t *testing.T) {
		test := &HTTPTest{}

		config := TestConfig{
			Name:     "Redirect Test - Don't Follow",
			Address:  "http://httpbin.org/redirect/2",
			Timeout:  10 * time.Second,
			Protocol: "HTTP",
			Config: &HTTPConfig{
				FollowRedirects: false,
				ExpectedStatus:  302, // Expect redirect status
			},
		}

		ctx := context.Background()
		result, err := test.Execute(ctx, config)

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.Status != TestStatusSuccess {
			t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
		}

		t.Logf("Don't follow redirects test: status=%v", result.Status)
	})
}

// TestHTTPTest_Integration_ConcurrentRequests tests multiple simultaneous HTTP requests
func TestHTTPTest_Integration_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	targets := []string{
		"https://www.google.com",
		"https://www.cloudflare.com",
		"https://1.1.1.1",
		"http://example.com",
	}

	ctx := context.Background()
	results := make(chan *TestResult, len(targets))
	errors := make(chan error, len(targets))

	// Launch concurrent requests
	for _, target := range targets {
		go func(addr string) {
			config := TestConfig{
				Name:     addr,
				Address:  addr,
				Timeout:  10 * time.Second,
				Protocol: "HTTP",
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
			t.Logf("Concurrent request to %s: latencyInMs=%.2f, status=%v, size=%d bytes",
				result.EndpointID, float64(result.Latency.Nanoseconds())/1_000_000.0, result.Status, result.ResponseSize)
		} else {
			t.Logf("Concurrent request to %s failed: status=%v, error=%v", result.EndpointID, result.Status, err)
		}
	}

	// At least some requests should succeed
	if successCount == 0 {
		t.Error("All concurrent requests failed")
	}

	t.Logf("Concurrent requests: %d/%d successful", successCount, len(targets))
}

// TestHTTPTest_Integration_CustomHeaders tests custom header support
func TestHTTPTest_Integration_CustomHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "Custom Headers Test",
		Address:  "https://httpbin.org/headers",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
		Config: &HTTPConfig{
			Headers: map[string]string{
				"X-Custom-Header": "TestValue",
			},
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	// Response should contain our custom header
	if result.ResponseSize <= 0 {
		t.Error("Execute() response size is 0, expected response body")
	}

	t.Logf("Custom headers test: status=%v, size=%d bytes", result.Status, result.ResponseSize)
}

// TestHTTPTest_Integration_HEAD_Method tests HEAD request method
func TestHTTPTest_Integration_HEAD_Method(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "HEAD Method Test",
		Address:  "https://www.google.com",
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
		Config: &HTTPConfig{
			Method: "HEAD",
		},
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	// HEAD responses should have no body
	if result.ResponseSize != 0 {
		t.Logf("HEAD response size = %d (some servers may send body)", result.ResponseSize)
	}

	t.Logf("HEAD method test: status=%v, latencyInMs=%.2f, size=%d",
		result.Status, float64(result.Latency.Nanoseconds())/1_000_000.0, result.ResponseSize)
}

// TestHTTPTest_Integration_IPv6 tests HTTP over IPv6
func TestHTTPTest_Integration_IPv6(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	config := TestConfig{
		Name:     "IPv6 Test",
		Address:  "http://[2606:2800:220:1:248:1893:25c8:1946]", // example.com IPv6
		Timeout:  10 * time.Second,
		Protocol: "HTTP",
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		// IPv6 might not be available on all systems
		if result.Status == TestStatusFailed || result.Status == TestStatusTimeout {
			t.Skipf("Skipping IPv6 test: %v", err)
		}
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Skipf("IPv6 test failed (IPv6 may not be available): status=%v, error=%v", result.Status, result.Error)
	}

	t.Logf("IPv6 test: status=%v, latencyInMs=%.2f", result.Status, float64(result.Latency.Nanoseconds())/1_000_000.0)
}

// TestHTTPTest_Integration_LargeResponse tests handling of large responses
func TestHTTPTest_Integration_LargeResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	// Request a large response (100KB)
	config := TestConfig{
		Name:     "Large Response Test",
		Address:  "https://httpbin.org/bytes/102400",
		Timeout:  15 * time.Second,
		Protocol: "HTTP",
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	// Should be approximately 100KB
	if result.ResponseSize < 100000 || result.ResponseSize > 105000 {
		t.Errorf("Execute() response size = %d, want ~102400", result.ResponseSize)
	}

	t.Logf("Large response test: status=%v, latencyInMs=%.2f, size=%d bytes",
		result.Status, float64(result.Latency.Nanoseconds())/1_000_000.0, result.ResponseSize)
}

// TestHTTPTest_Integration_SlowServer tests timeout handling with slow servers
func TestHTTPTest_Integration_SlowServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	test := &HTTPTest{}

	// Server delays for 5 seconds, but we only wait 2
	config := TestConfig{
		Name:     "Slow Server Test",
		Address:  "https://httpbin.org/delay/5",
		Timeout:  2 * time.Second,
		Protocol: "HTTP",
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	// Should timeout
	if err == nil {
		t.Error("Execute() expected timeout error, got nil")
	}

	if result.Status != TestStatusTimeout {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusTimeout)
	}

	t.Logf("Slow server test: status=%v, error=%v", result.Status, result.Error)
}
