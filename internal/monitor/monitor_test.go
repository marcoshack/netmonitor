package monitor

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"context"

	"github.com/marcoshack/netmonitor/internal/models"
)

func TestMonitorHTTP(t *testing.T) {
	// Start local server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	mon := NewMonitor(context.Background(), nil) // Config not needed for direct TestEndpoint

	ep := models.Endpoint{
		Name:    "Test HTTP",
		Type:    models.TypeHTTP,
		Address: ts.URL,
		Timeout: 1000,
	}

	res := mon.TestEndpoint(ep)
	if res.St != ResultSuccess {
		t.Errorf("Expected success, got %d (err: %s)", res.St, res.Err)
	}
	if res.Ms < 0 {
		t.Errorf("Invalid latency: %d", res.Ms)
	}

	// Test Failure
	ep.Address = "http://localhost:59999" // Unlikely port
	// Ensure fast fail?
	res = mon.TestEndpoint(ep)
	if res.St != ResultError {
		t.Errorf("Expected failure for bad port, got %d", res.St)
	}
}

func TestMonitorTCP(t *testing.T) {
	// Listen on random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	mon := NewMonitor(context.Background(), nil)

	ep := models.Endpoint{
		Name:    "Test TCP",
		Type:    models.TypeTCP,
		Address: ln.Addr().String(),
		Timeout: 1000,
	}

	res := mon.TestEndpoint(ep)
	if res.St != ResultSuccess {
		t.Errorf("Expected success for TCP, got %d (err: %s)", res.St, res.Err)
	}
}

func TestCheckICMP_Integration(t *testing.T) {
	// Pinging localhost should generally work, but might require privileges or specific setup on Windows.
	// Since we are switching to pro-bing with unprivileged support via API, this test is crucial.

	// Skip if short test? No, we want to run this.

	target := "127.0.0.1"
	timeout := 2 * time.Second

	fmt.Printf("Attempting to ping %s...\n", target)

	_, err := checkICMP(target, timeout)
	if err != nil {
		t.Logf("ICMP Ping to %s failed: %v", target, err)
		t.Logf("Note: This might be expected if running without sufficient privileges or OS support.")
		// We don't fail the test immediately because CI/Environments vary,
		// but for local dev this should pass.
		// Un-commenting Fatal to enforce it for now as per user request.
		t.Fatal(err)
	} else {
		t.Logf("ICMP Ping to %s succeeded", target)
	}
}
