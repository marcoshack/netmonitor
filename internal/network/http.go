package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

// HTTPTest implements the NetworkTest interface for HTTP/HTTPS connectivity testing
type HTTPTest struct{}

// GetProtocol returns the protocol type this test implements
func (t *HTTPTest) GetProtocol() string {
	return "HTTP"
}

// Validate checks if the configuration is valid for HTTP tests
func (t *HTTPTest) Validate(config TestConfig) error {
	if config.Address == "" {
		return fmt.Errorf("address is required for HTTP test")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}

	// Validate HTTP-specific config if provided
	if config.Config != nil {
		httpConfig, ok := config.Config.(*HTTPConfig)
		if !ok {
			return fmt.Errorf("invalid config type for HTTP test")
		}

		if httpConfig.Method != "" && httpConfig.Method != "GET" && httpConfig.Method != "HEAD" && httpConfig.Method != "POST" && httpConfig.Method != "PUT" {
			return fmt.Errorf("unsupported HTTP method: %s", httpConfig.Method)
		}
	}

	return nil
}

// Execute runs the HTTP test with the given configuration
func (t *HTTPTest) Execute(ctx context.Context, config TestConfig) (*TestResult, error) {
	startTime := time.Now()

	// Validate configuration
	if err := t.Validate(config); err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "HTTP",
			Status:    TestStatusError,
			Error:     err.Error(),
		}, err
	}

	// Get HTTP-specific config or use defaults
	httpConfig := &HTTPConfig{
		Method:          "GET",
		Headers:         make(map[string]string),
		FollowRedirects: true,
		ValidateSSL:     true,
		ExpectedStatus:  0, // 0 means any 2xx is acceptable
	}
	if config.Config != nil {
		if cfg, ok := config.Config.(*HTTPConfig); ok {
			httpConfig = cfg
			if httpConfig.Method == "" {
				httpConfig.Method = "GET"
			}
			if httpConfig.Headers == nil {
				httpConfig.Headers = make(map[string]string)
			}
		}
	}

	// Track timing metrics - these can be used for detailed analysis
	// They are captured via httptrace but not currently exposed in the result
	var dnsStart, dnsEnd, connectStart, connectEnd, tlsStart, tlsEnd, requestStart, responseStart time.Time
	_ = dnsStart    // Mark as intentionally unused for now
	_ = dnsEnd      // Mark as intentionally unused for now
	_ = connectStart // Mark as intentionally unused for now
	_ = connectEnd  // Mark as intentionally unused for now
	_ = tlsStart    // Mark as intentionally unused for now
	_ = tlsEnd      // Mark as intentionally unused for now
	_ = requestStart // Mark as intentionally unused for now
	_ = responseStart // Mark as intentionally unused for now

	// Create HTTP trace to measure different phases
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
		},
		ConnectStart: func(_, _ string) {
			connectStart = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			connectEnd = time.Now()
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			tlsEnd = time.Now()
		},
		WroteRequest: func(_ httptrace.WroteRequestInfo) {
			requestStart = time.Now()
		},
		GotFirstResponseByte: func() {
			responseStart = time.Now()
		},
	}

	// Create context with trace
	ctx = httptrace.WithClientTrace(ctx, trace)

	// Create HTTP client with timeout and SSL configuration
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !httpConfig.ValidateSSL,
			},
		},
	}

	// Configure redirect policy
	if !httpConfig.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, httpConfig.Method, config.Address, nil)
	if err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "HTTP",
			Status:    TestStatusError,
			Error:     fmt.Sprintf("failed to create request: %v", err),
		}, err
	}

	// Set custom User-Agent
	req.Header.Set("User-Agent", "NetworkMonitor/1.0")

	// Add custom headers
	for key, value := range httpConfig.Headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		// Determine the type of error
		// Check for context cancellation or timeout errors
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			return &TestResult{
				Timestamp: startTime,
				Protocol:  "HTTP",
				Status:    TestStatusTimeout,
				Error:     "request timeout",
				Latency:   0,
			}, err
		}
		// Also check for http.Client timeout
		errStr := err.Error()
		if strings.Contains(errStr, "Client.Timeout exceeded") || strings.Contains(errStr, "context deadline exceeded") {
			return &TestResult{
				Timestamp: startTime,
				Protocol:  "HTTP",
				Status:    TestStatusTimeout,
				Error:     "request timeout",
				Latency:   0,
			}, err
		}
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "HTTP",
			Status:    TestStatusFailed,
			Error:     fmt.Sprintf("request failed: %v", err),
			Latency:   0,
		}, err
	}
	defer resp.Body.Close()

	// Read response body to measure size
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestResult{
			Timestamp: startTime,
			Protocol:  "HTTP",
			Status:    TestStatusFailed,
			Error:     fmt.Sprintf("failed to read response: %v", err),
			Latency:   0,
		}, err
	}

	// Calculate total latency
	totalLatency := time.Since(startTime)

	// Prepare result
	result := &TestResult{
		Timestamp:    startTime,
		EndpointID:   config.Name,
		Protocol:     "HTTP",
		Latency:      totalLatency,
		ResponseSize: int64(len(bodyBytes)),
	}

	// Check HTTP status code
	expectedStatus := httpConfig.ExpectedStatus
	if expectedStatus == 0 {
		// Accept any 2xx status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result.Status = TestStatusSuccess
		} else {
			result.Status = TestStatusFailed
			result.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		}
	} else {
		// Check for specific status code
		if resp.StatusCode == expectedStatus {
			result.Status = TestStatusSuccess
		} else {
			result.Status = TestStatusFailed
			result.Error = fmt.Sprintf("expected status %d, got %d", expectedStatus, resp.StatusCode)
		}
	}

	return result, nil
}

// GetTimingBreakdown returns a detailed breakdown of request timing
// This is a helper function that can be used to get detailed timing information
func GetTimingBreakdown(dnsTime, connectTime, tlsTime, responseTime time.Duration) map[string]time.Duration {
	return map[string]time.Duration{
		"dns":      dnsTime,
		"connect":  connectTime,
		"tls":      tlsTime,
		"response": responseTime,
	}
}
