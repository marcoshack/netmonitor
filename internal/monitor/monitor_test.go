package monitor

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marcoshack/netmonitor/internal/models"
)

func TestMonitorHTTP(t *testing.T) {
	// Start local server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	mon := NewMonitor(nil) // Config not needed for direct TestEndpoint

	ep := models.Endpoint{
		Name:    "Test HTTP",
		Type:    models.TypeHTTP,
		Address: ts.URL,
		Timeout: 1000,
	}

	res := mon.TestEndpoint(ep)
	if res.Status != "success" {
		t.Errorf("Expected success, got %s (err: %s)", res.Status, res.Error)
	}
	if res.LatencyMs < 0 {
		t.Errorf("Invalid latency: %d", res.LatencyMs)
	}

	// Test Failure
	ep.Address = "http://localhost:59999" // Unlikely port
	// Ensure fast fail?
	res = mon.TestEndpoint(ep)
	if res.Status != "failure" {
		t.Errorf("Expected failure for bad port, got %s", res.Status)
	}
}

func TestMonitorTCP(t *testing.T) {
	// Listen on random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	mon := NewMonitor(nil)

	ep := models.Endpoint{
		Name:    "Test TCP",
		Type:    models.TypeTCP,
		Address: ln.Addr().String(),
		Timeout: 1000,
	}

	res := mon.TestEndpoint(ep)
	if res.Status != "success" {
		t.Errorf("Expected success for TCP, got %s (err: %s)", res.Status, res.Error)
	}
}
