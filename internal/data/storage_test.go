package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcoshack/netmonitor/internal/models"
)

func TestStorage(t *testing.T) {
	tmpDir := "test_data"
	defer os.RemoveAll(tmpDir)

	s := NewStorage(tmpDir)

	ts := time.Date(2023, 11, 15, 12, 0, 0, 0, time.UTC)
	res1 := models.TestResult{
		Timestamp:  ts,
		EndpointID: "test-ep",
		LatencyMs:  50,
		Status:     "success",
	}

	// Test Save
	err := s.SaveResult(res1)
	if err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	// Check file exists
	fp := filepath.Join(tmpDir, "2023-11-15.json")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Errorf("File %s not created", fp)
	}

	// Test Load
	results, err := s.GetResultsForDay(ts)
	if err != nil {
		t.Fatalf("GetResultsForDay failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].EndpointID != "test-ep" {
		t.Errorf("Expected endpoint ID test-ep, got %s", results[0].EndpointID)
	}

	// Append another
	res2 := models.TestResult{
		Timestamp:  ts.Add(1 * time.Minute),
		EndpointID: "test-ep-2",
		LatencyMs:  60,
		Status:     "success",
	}
	_ = s.SaveResult(res2)

	results, _ = s.GetResultsForDay(ts)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}
