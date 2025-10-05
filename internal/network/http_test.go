package network

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPTest_GetProtocol(t *testing.T) {
	test := &HTTPTest{}
	if got := test.GetProtocol(); got != "HTTP" {
		t.Errorf("GetProtocol() = %v, want HTTP", got)
	}
}

func TestHTTPTest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TestConfig
		wantErr bool
	}{
		{
			name: "valid HTTP URL",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid HTTPS URL",
			config: TestConfig{
				Address: "https://example.com",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid with IP address",
			config: TestConfig{
				Address: "http://192.168.1.1",
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
			name: "zero timeout",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 0,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "valid with HTTP config GET",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 5 * time.Second,
				Config: &HTTPConfig{
					Method:          "GET",
					FollowRedirects: true,
					ValidateSSL:     true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid with HTTP config POST",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 5 * time.Second,
				Config: &HTTPConfig{
					Method:          "POST",
					FollowRedirects: true,
					ValidateSSL:     true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid with HTTP config HEAD",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 5 * time.Second,
				Config: &HTTPConfig{
					Method:          "HEAD",
					FollowRedirects: false,
					ValidateSSL:     false,
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported HTTP method",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 5 * time.Second,
				Config: &HTTPConfig{
					Method: "DELETE",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config type",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 5 * time.Second,
				Config:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := &HTTPTest{}
			err := test.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPTest_Execute_Success(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header
		if ua := r.Header.Get("User-Agent"); ua != "NetworkMonitor/1.0" {
			t.Errorf("User-Agent = %v, want NetworkMonitor/1.0", ua)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	test := &HTTPTest{}
	config := TestConfig{
		Name:    "Test Server",
		Address: server.URL,
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, TestStatusSuccess, result.Error)
	}

	if result.Protocol != "HTTP" {
		t.Errorf("Execute() protocol = %v, want HTTP", result.Protocol)
	}

	if result.Latency <= 0 {
		t.Errorf("Execute() latency = %v, want > 0", result.Latency)
	}

	if result.ResponseSize != 13 { // "Hello, World!" is 13 bytes
		t.Errorf("Execute() response size = %v, want 13", result.ResponseSize)
	}

	if result.EndpointID != "Test Server" {
		t.Errorf("Execute() endpoint ID = %v, want Test Server", result.EndpointID)
	}
}

func TestHTTPTest_Execute_HTTPSSuccess(t *testing.T) {
	// Create a test HTTPS server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Secure!"))
	}))
	defer server.Close()

	test := &HTTPTest{}
	config := TestConfig{
		Name:    "HTTPS Test Server",
		Address: server.URL,
		Timeout: 5 * time.Second,
		Config: &HTTPConfig{
			ValidateSSL: false, // Test server uses self-signed cert
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

	if result.ResponseSize != 7 { // "Secure!" is 7 bytes
		t.Errorf("Execute() response size = %v, want 7", result.ResponseSize)
	}
}

func TestHTTPTest_Execute_CustomHeaders(t *testing.T) {
	// Create a test HTTP server that checks custom headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "CustomValue" {
			t.Errorf("Custom header not found or incorrect")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	test := &HTTPTest{}
	config := TestConfig{
		Name:    "Custom Headers Test",
		Address: server.URL,
		Timeout: 5 * time.Second,
		Config: &HTTPConfig{
			Headers: map[string]string{
				"X-Custom-Header": "CustomValue",
			},
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

func TestHTTPTest_Execute_StatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
		wantStatus     TestStatus
	}{
		{
			name:           "200 OK - default expectation",
			statusCode:     200,
			expectedStatus: 0, // 0 means any 2xx
			wantStatus:     TestStatusSuccess,
		},
		{
			name:           "201 Created - default expectation",
			statusCode:     201,
			expectedStatus: 0,
			wantStatus:     TestStatusSuccess,
		},
		{
			name:           "404 Not Found - default expectation",
			statusCode:     404,
			expectedStatus: 0,
			wantStatus:     TestStatusFailed,
		},
		{
			name:           "500 Internal Server Error - default expectation",
			statusCode:     500,
			expectedStatus: 0,
			wantStatus:     TestStatusFailed,
		},
		{
			name:           "404 expected and received",
			statusCode:     404,
			expectedStatus: 404,
			wantStatus:     TestStatusSuccess,
		},
		{
			name:           "200 expected but 404 received",
			statusCode:     404,
			expectedStatus: 200,
			wantStatus:     TestStatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			test := &HTTPTest{}
			config := TestConfig{
				Name:    tt.name,
				Address: server.URL,
				Timeout: 5 * time.Second,
				Config: &HTTPConfig{
					ExpectedStatus: tt.expectedStatus,
				},
			}

			ctx := context.Background()
			result, _ := test.Execute(ctx, config)

			if result.Status != tt.wantStatus {
				t.Errorf("Execute() status = %v, want %v, error: %v", result.Status, tt.wantStatus, result.Error)
			}
		})
	}
}

func TestHTTPTest_Execute_Redirects(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			redirectCount++
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Final destination"))
	}))
	defer server.Close()

	t.Run("follow redirects", func(t *testing.T) {
		redirectCount = 0
		test := &HTTPTest{}
		config := TestConfig{
			Name:    "Redirect Test - Follow",
			Address: server.URL + "/redirect",
			Timeout: 5 * time.Second,
			Config: &HTTPConfig{
				FollowRedirects: true,
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

		if redirectCount != 1 {
			t.Errorf("Expected 1 redirect, got %d", redirectCount)
		}
	})

	t.Run("don't follow redirects", func(t *testing.T) {
		redirectCount = 0
		test := &HTTPTest{}
		config := TestConfig{
			Name:    "Redirect Test - Don't Follow",
			Address: server.URL + "/redirect",
			Timeout: 5 * time.Second,
			Config: &HTTPConfig{
				FollowRedirects: false,
				ExpectedStatus:  302, // Expect the redirect status
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
	})
}

func TestHTTPTest_Execute_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	test := &HTTPTest{}
	config := TestConfig{
		Name:    "Timeout Test",
		Address: server.URL,
		Timeout: 100 * time.Millisecond,
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err == nil {
		t.Error("Execute() expected timeout error, got nil")
	}

	if result.Status != TestStatusTimeout {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusTimeout)
	}

	if result.Latency != 0 {
		t.Errorf("Execute() latency = %v, want 0 for timeout", result.Latency)
	}
}

func TestHTTPTest_Execute_ValidationErrors(t *testing.T) {
	test := &HTTPTest{}
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
			wantProtocol: "HTTP",
		},
		{
			name: "zero timeout",
			config: TestConfig{
				Address: "http://example.com",
				Timeout: 0,
			},
			wantStatus:   TestStatusError,
			wantProtocol: "HTTP",
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

func TestHTTPTest_Execute_NetworkError(t *testing.T) {
	test := &HTTPTest{}
	config := TestConfig{
		Name:    "Network Error Test",
		Address: "http://192.0.2.1:9999", // TEST-NET-1, should be unreachable
		Timeout: 1 * time.Second,
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err == nil {
		t.Error("Execute() expected network error, got nil")
	}

	// Should be either timeout or failed
	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v or %v", result.Status, TestStatusTimeout, TestStatusFailed)
	}
}

func TestHTTPTest_Execute_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cancellation test in short mode")
	}

	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	test := &HTTPTest{}
	config := TestConfig{
		Name:    "Cancellation Test",
		Address: server.URL,
		Timeout: 30 * time.Second,
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

	// Should complete quickly (within 2 seconds, not the full timeout)
	if elapsed > 2*time.Second {
		t.Errorf("Execute() took %v, expected quick cancellation", elapsed)
	}

	if result.Status != TestStatusTimeout && result.Status != TestStatusFailed {
		t.Errorf("Execute() status = %v, want %v or %v", result.Status, TestStatusTimeout, TestStatusFailed)
	}
}

func TestHTTPTest_Execute_DefaultConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify default method is GET
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	test := &HTTPTest{}
	config := TestConfig{
		Name:    "Default Config Test",
		Address: server.URL,
		Timeout: 5 * time.Second,
		// No Config provided - should use defaults
	}

	ctx := context.Background()
	result, err := test.Execute(ctx, config)

	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	if result.Status != TestStatusSuccess {
		t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusSuccess)
	}

	if result.Protocol != "HTTP" {
		t.Errorf("Execute() protocol = %v, want HTTP", result.Protocol)
	}
}

func TestHTTPTest_Execute_SSLValidation(t *testing.T) {
	// Create a test HTTPS server (with self-signed cert)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Run("skip SSL validation", func(t *testing.T) {
		test := &HTTPTest{}
		config := TestConfig{
			Name:    "SSL Skip Validation",
			Address: server.URL,
			Timeout: 5 * time.Second,
			Config: &HTTPConfig{
				ValidateSSL: false,
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
	})

	t.Run("enforce SSL validation", func(t *testing.T) {
		test := &HTTPTest{}
		config := TestConfig{
			Name:    "SSL Enforce Validation",
			Address: server.URL,
			Timeout: 5 * time.Second,
			Config: &HTTPConfig{
				ValidateSSL: true,
			},
		}

		ctx := context.Background()
		result, err := test.Execute(ctx, config)

		// Should fail because of self-signed certificate
		if err == nil {
			t.Error("Execute() expected SSL validation error, got nil")
		}

		if result.Status != TestStatusFailed {
			t.Errorf("Execute() status = %v, want %v", result.Status, TestStatusFailed)
		}
	})
}

func TestHTTPTest_Execute_HTTPMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"GET method", "GET"},
		{"POST method", "POST"},
		{"HEAD method", "HEAD"},
		{"PUT method", "PUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method {
					t.Errorf("Method = %v, want %v", r.Method, tt.method)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			test := &HTTPTest{}
			config := TestConfig{
				Name:    fmt.Sprintf("%s Test", tt.method),
				Address: server.URL,
				Timeout: 5 * time.Second,
				Config: &HTTPConfig{
					Method: tt.method,
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
		})
	}
}

func TestGetTimingBreakdown(t *testing.T) {
	dns := 10 * time.Millisecond
	connect := 20 * time.Millisecond
	tls := 30 * time.Millisecond
	response := 40 * time.Millisecond

	breakdown := GetTimingBreakdown(dns, connect, tls, response)

	if breakdown["dns"] != dns {
		t.Errorf("DNS time = %v, want %v", breakdown["dns"], dns)
	}
	if breakdown["connect"] != connect {
		t.Errorf("Connect time = %v, want %v", breakdown["connect"], connect)
	}
	if breakdown["tls"] != tls {
		t.Errorf("TLS time = %v, want %v", breakdown["tls"], tls)
	}
	if breakdown["response"] != response {
		t.Errorf("Response time = %v, want %v", breakdown["response"], response)
	}
}
