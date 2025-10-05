package storage

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTestResult_MarshalJSON(t *testing.T) {
	result := &TestResult{
		Timestamp:  time.Date(2025, 10, 5, 14, 31, 15, 0, time.UTC),
		EndpointID: "EU-West-Cloudflare DNS",
		Protocol:   "ICMP",
		Latency:    25 * time.Millisecond,
		Status:     "success",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Check that latencyInMs is present and correct
	latencyInMs, ok := unmarshaled["latencyInMs"].(float64)
	if !ok {
		t.Errorf("latencyInMs not found or not a float64: %v", unmarshaled["latencyInMs"])
	}

	if latencyInMs != 25.0 {
		t.Errorf("latencyInMs = %v, want 25.0", latencyInMs)
	}

	// Check that latency field is not present
	if _, exists := unmarshaled["latency"]; exists {
		t.Errorf("latency field should not be present in JSON")
	}
}

func TestTestResult_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"timestamp": "2025-10-05T14:31:15Z",
		"endpoint_id": "EU-West-Cloudflare DNS",
		"protocol": "ICMP",
		"latencyInMs": 25.0,
		"status": "success"
	}`

	var result TestResult
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	expectedLatency := 25 * time.Millisecond
	if result.Latency != expectedLatency {
		t.Errorf("Latency = %v, want %v", result.Latency, expectedLatency)
	}

	if result.EndpointID != "EU-West-Cloudflare DNS" {
		t.Errorf("EndpointID = %v, want EU-West-Cloudflare DNS", result.EndpointID)
	}
}

func TestTestResult_RoundTrip(t *testing.T) {
	original := &TestResult{
		Timestamp:  time.Date(2025, 10, 5, 14, 31, 15, 0, time.UTC),
		EndpointID: "EU-West-Cloudflare DNS",
		Protocol:   "ICMP",
		Latency:    35500 * time.Microsecond, // 35.5ms
		Status:     "success",
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal
	var decoded TestResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Compare - allow small rounding error due to float conversion
	diff := decoded.Latency - original.Latency
	if diff < -time.Microsecond || diff > time.Microsecond {
		t.Errorf("Latency after round-trip = %v, want %v (diff: %v)", decoded.Latency, original.Latency, diff)
	}

	if decoded.EndpointID != original.EndpointID {
		t.Errorf("EndpointID = %v, want %v", decoded.EndpointID, original.EndpointID)
	}

	if decoded.Status != original.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, original.Status)
	}
}
