package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

func TestManager_StoreAndRetrieveResults(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Create test results
	now := time.Now()
	results := []*TestResult{
		{
			Timestamp:  now,
			EndpointID: "test-endpoint-1",
			Protocol:   "ICMP",
			Latency:    25 * time.Millisecond,
			Status:     "success",
		},
		{
			Timestamp:  now.Add(1 * time.Minute),
			EndpointID: "test-endpoint-2",
			Protocol:   "TCP",
			Latency:    50 * time.Millisecond,
			Status:     "success",
		},
	}

	// Store results
	for _, result := range results {
		if err := manager.StoreTestResult(result); err != nil {
			t.Fatalf("StoreTestResult() error = %v", err)
		}
	}

	// Retrieve results
	retrieved, err := manager.GetResults(now)
	if err != nil {
		t.Fatalf("GetResults() error = %v", err)
	}

	if len(retrieved) != len(results) {
		t.Errorf("Retrieved %d results, want %d", len(retrieved), len(results))
	}
}

func TestManager_GetResultsRange(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Create test results across multiple days
	baseDate := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		result := &TestResult{
			Timestamp:  baseDate.AddDate(0, 0, i),
			EndpointID: "test-endpoint",
			Protocol:   "ICMP",
			Latency:    time.Duration(20+i) * time.Millisecond,
			Status:     "success",
		}
		if err := manager.StoreTestResult(result); err != nil {
			t.Fatalf("StoreTestResult() error = %v", err)
		}
	}

	// Test range query
	start := baseDate
	end := baseDate.AddDate(0, 0, 4)
	results, err := manager.GetResultsRange(start, end)
	if err != nil {
		t.Fatalf("GetResultsRange() error = %v", err)
	}

	if len(results) != 5 {
		t.Errorf("GetResultsRange() returned %d results, want 5", len(results))
	}

	// Test partial range
	start = baseDate.AddDate(0, 0, 1)
	end = baseDate.AddDate(0, 0, 3)
	results, err = manager.GetResultsRange(start, end)
	if err != nil {
		t.Fatalf("GetResultsRange() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("GetResultsRange() returned %d results, want 3", len(results))
	}
}

func TestManager_CleanupOldFiles(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Create old test results
	oldDate := time.Now().AddDate(0, 0, -40)
	recentDate := time.Now().AddDate(0, 0, -5)

	oldResult := &TestResult{
		Timestamp:  oldDate,
		EndpointID: "test-endpoint-old",
		Protocol:   "ICMP",
		Latency:    25 * time.Millisecond,
		Status:     "success",
	}

	recentResult := &TestResult{
		Timestamp:  recentDate,
		EndpointID: "test-endpoint-recent",
		Protocol:   "ICMP",
		Latency:    30 * time.Millisecond,
		Status:     "success",
	}

	if err := manager.StoreTestResult(oldResult); err != nil {
		t.Fatalf("StoreTestResult() error = %v", err)
	}
	if err := manager.StoreTestResult(recentResult); err != nil {
		t.Fatalf("StoreTestResult() error = %v", err)
	}

	// Get initial stats
	stats, err := manager.GetStorageStats()
	if err != nil {
		t.Fatalf("GetStorageStats() error = %v", err)
	}
	initialFileCount := stats.TotalFiles

	// Cleanup files older than 30 days
	if err := manager.CleanupOldFiles(30); err != nil {
		t.Fatalf("CleanupOldFiles() error = %v", err)
	}

	// Get stats after cleanup
	stats, err = manager.GetStorageStats()
	if err != nil {
		t.Fatalf("GetStorageStats() error = %v", err)
	}

	if stats.TotalFiles >= initialFileCount {
		t.Errorf("Expected file count to decrease after cleanup, got %d, initial was %d",
			stats.TotalFiles, initialFileCount)
	}

	// Verify old file is gone
	_, err = manager.GetResults(oldDate)
	// Should have no results or error for old date
	if err == nil {
		results, _ := manager.GetResults(oldDate)
		if len(results) > 0 {
			t.Error("Expected old file to be deleted")
		}
	}

	// Verify recent file still exists
	recentResults, err := manager.GetResults(recentDate)
	if err != nil {
		t.Fatalf("GetResults() for recent date error = %v", err)
	}
	if len(recentResults) == 0 {
		t.Error("Expected recent file to still exist")
	}
}

func TestManager_ValidateDataFile(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Create a valid data file
	now := time.Now()
	result := &TestResult{
		Timestamp:  now,
		EndpointID: "test-endpoint",
		Protocol:   "ICMP",
		Latency:    25 * time.Millisecond,
		Status:     "success",
	}

	if err := manager.StoreTestResult(result); err != nil {
		t.Fatalf("StoreTestResult() error = %v", err)
	}

	// Validate the file
	filename := fmt.Sprintf("%s.json", now.Format("2006-01-02"))
	filepath := filepath.Join(tempDir, filename)

	if err := manager.ValidateDataFile(filepath); err != nil {
		t.Errorf("ValidateDataFile() error = %v, expected valid file", err)
	}
}

func TestManager_RecoverDataFile(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Create a data file with mismatched metadata
	filename := "2025-10-15.json"
	filepath := filepath.Join(tempDir, filename)

	corruptedData := DailyDataFile{
		Date: "2025-10-15",
		Results: []*TestResult{
			{
				Timestamp:  time.Date(2025, 10, 15, 10, 0, 0, 0, time.UTC),
				EndpointID: "test",
				Protocol:   "ICMP",
				Latency:    25 * time.Millisecond,
				Status:     "success",
			},
		},
		Metadata: &FileMetadata{
			Version:      "1.0.0",
			CreatedAt:    time.Now(),
			LastModified: time.Now(),
			ResultCount:  999, // Wrong count
		},
	}

	data, _ := json.MarshalIndent(corruptedData, "", "  ")
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify file is invalid
	if err := manager.ValidateDataFile(filepath); err == nil {
		t.Error("Expected validation error for corrupted file")
	}

	// Recover the file
	if err := manager.RecoverDataFile(filepath); err != nil {
		t.Fatalf("RecoverDataFile() error = %v", err)
	}

	// Verify file is now valid
	if err := manager.ValidateDataFile(filepath); err != nil {
		t.Errorf("ValidateDataFile() after recovery error = %v", err)
	}
}

func TestManager_ConfigurationStorage(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Test configuration structure
	type TestConfig struct {
		Name    string `json:"name"`
		Value   int    `json:"value"`
		Enabled bool   `json:"enabled"`
	}

	originalConfig := TestConfig{
		Name:    "test-config",
		Value:   42,
		Enabled: true,
	}

	// Save configuration
	if err := manager.SaveConfiguration(originalConfig); err != nil {
		t.Fatalf("SaveConfiguration() error = %v", err)
	}

	// Load configuration
	var loadedConfig TestConfig
	if err := manager.LoadConfiguration(&loadedConfig); err != nil {
		t.Fatalf("LoadConfiguration() error = %v", err)
	}

	// Verify configuration matches
	if loadedConfig.Name != originalConfig.Name {
		t.Errorf("Name = %v, want %v", loadedConfig.Name, originalConfig.Name)
	}
	if loadedConfig.Value != originalConfig.Value {
		t.Errorf("Value = %v, want %v", loadedConfig.Value, originalConfig.Value)
	}
	if loadedConfig.Enabled != originalConfig.Enabled {
		t.Errorf("Enabled = %v, want %v", loadedConfig.Enabled, originalConfig.Enabled)
	}
}

func TestManager_GetStorageStats(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Get initial stats
	stats, err := manager.GetStorageStats()
	if err != nil {
		t.Fatalf("GetStorageStats() error = %v", err)
	}

	if stats.DataDirectory != tempDir {
		t.Errorf("DataDirectory = %v, want %v", stats.DataDirectory, tempDir)
	}

	initialFiles := stats.TotalFiles

	// Add some data
	now := time.Now()
	result := &TestResult{
		Timestamp:  now,
		EndpointID: "test",
		Protocol:   "ICMP",
		Latency:    25 * time.Millisecond,
		Status:     "success",
	}

	if err := manager.StoreTestResult(result); err != nil {
		t.Fatalf("StoreTestResult() error = %v", err)
	}

	// Get stats again
	stats, err = manager.GetStorageStats()
	if err != nil {
		t.Fatalf("GetStorageStats() error = %v", err)
	}

	if stats.TotalFiles <= initialFiles {
		t.Errorf("Expected file count to increase, got %d", stats.TotalFiles)
	}

	if stats.TotalSizeBytes == 0 {
		t.Error("Expected TotalSizeBytes to be greater than 0")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	// Test concurrent writes
	now := time.Now()
	numGoroutines := 10
	resultsPerGoroutine := 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*resultsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < resultsPerGoroutine; j++ {
				result := &TestResult{
					Timestamp:  now,
					EndpointID: fmt.Sprintf("endpoint-%d-%d", goroutineID, j),
					Protocol:   "ICMP",
					Latency:    time.Duration(20+j) * time.Millisecond,
					Status:     "success",
				}
				if err := manager.StoreTestResult(result); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent write error: %v", err)
	}

	// Verify all results were stored
	results, err := manager.GetResults(now)
	if err != nil {
		t.Fatalf("GetResults() error = %v", err)
	}

	expectedCount := numGoroutines * resultsPerGoroutine
	if len(results) != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, len(results))
	}

	// Test concurrent reads while writing
	wg = sync.WaitGroup{}
	readErrors := make(chan error, numGoroutines)

	// Start writers
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			result := &TestResult{
				Timestamp:  now,
				EndpointID: fmt.Sprintf("concurrent-endpoint-%d", goroutineID),
				Protocol:   "TCP",
				Latency:    30 * time.Millisecond,
				Status:     "success",
			}
			if err := manager.StoreTestResult(result); err != nil {
				readErrors <- err
			}
		}(i)
	}

	// Start readers
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := manager.GetResults(now); err != nil {
				readErrors <- err
			}
		}()
	}

	wg.Wait()
	close(readErrors)

	// Check for errors
	for err := range readErrors {
		t.Errorf("Concurrent read/write error: %v", err)
	}
}

func TestManager_AtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	manager, err := NewManager(ctx, tempDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Close()

	now := time.Now()
	result := &TestResult{
		Timestamp:  now,
		EndpointID: "test-endpoint",
		Protocol:   "ICMP",
		Latency:    25 * time.Millisecond,
		Status:     "success",
	}

	// Store result
	if err := manager.StoreTestResult(result); err != nil {
		t.Fatalf("StoreTestResult() error = %v", err)
	}

	// Verify no .tmp files exist (atomic write should clean up)
	filename := fmt.Sprintf("%s.json", now.Format("2006-01-02"))
	tempFilePath := filepath.Join(tempDir, filename+".tmp")

	if _, err := os.Stat(tempFilePath); err == nil {
		t.Error("Temporary file should not exist after successful write")
	}

	// Verify actual file exists
	actualFilePath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(actualFilePath); err != nil {
		t.Errorf("Data file should exist after write: %v", err)
	}
}
