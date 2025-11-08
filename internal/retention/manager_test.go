package retention

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	policy := DefaultRetentionPolicy()
	policy.AutoCleanupEnabled = false // Disable for testing

	manager, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Verify policy
	retrievedPolicy := manager.GetPolicy()
	if retrievedPolicy.RawDataDays != policy.RawDataDays {
		t.Errorf("Expected RawDataDays %d, got %d", policy.RawDataDays, retrievedPolicy.RawDataDays)
	}
}

func TestNewManagerWithInvalidPolicy(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	invalidPolicy := &RetentionPolicy{
		RawDataDays:        5, // Invalid (too low)
		AggregatedDataDays: 365,
		ConfigBackupDays:   30,
		AutoCleanupEnabled: false,
		CleanupTime:        "02:00",
	}

	_, err := NewManager(ctx, tempDir, invalidPolicy)
	if err == nil {
		t.Error("Expected error when creating manager with invalid policy")
	}
}

func TestUpdatePolicy(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	initialPolicy := DefaultRetentionPolicy()
	initialPolicy.AutoCleanupEnabled = false

	manager, err := NewManager(ctx, tempDir, initialPolicy)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Update policy
	newPolicy := &RetentionPolicy{
		RawDataDays:        30,
		AggregatedDataDays: 180,
		ConfigBackupDays:   14,
		AutoCleanupEnabled: false,
		CleanupTime:        "03:00",
	}

	err = manager.UpdatePolicy(newPolicy)
	if err != nil {
		t.Fatalf("Failed to update policy: %v", err)
	}

	// Verify policy was updated
	retrievedPolicy := manager.GetPolicy()
	if retrievedPolicy.RawDataDays != 30 {
		t.Errorf("Expected RawDataDays 30, got %d", retrievedPolicy.RawDataDays)
	}
	if retrievedPolicy.CleanupTime != "03:00" {
		t.Errorf("Expected CleanupTime '03:00', got '%s'", retrievedPolicy.CleanupTime)
	}
}

func TestGetStorageStats(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	manager, err := NewManager(ctx, tempDir, DefaultRetentionPolicy())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create some test data files
	testFiles := []struct {
		name    string
		content string
	}{
		{"2025-01-01.json", `{"date":"2025-01-01","results":[]}`},
		{"2025-01-02.json", `{"date":"2025-01-02","results":[]}`},
		{"2025-01-03.json", `{"date":"2025-01-03","results":[]}`},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		if err := os.WriteFile(filePath, []byte(tf.content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Get storage stats
	stats, err := manager.GetStorageStats()
	if err != nil {
		t.Fatalf("Failed to get storage stats: %v", err)
	}

	if stats.TotalFiles != 3 {
		t.Errorf("Expected 3 files, got %d", stats.TotalFiles)
	}

	if stats.TotalSizeBytes == 0 {
		t.Error("Expected non-zero total size")
	}

	if stats.DaysOfData != 3 {
		t.Errorf("Expected 3 days of data, got %d", stats.DaysOfData)
	}
}

func TestTriggerManualCleanup(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	policy := &RetentionPolicy{
		RawDataDays:        30,
		AggregatedDataDays: 180,
		ConfigBackupDays:   14,
		AutoCleanupEnabled: false,
		CleanupTime:        "02:00",
	}

	manager, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create old and recent test files
	now := time.Now()
	oldDate := now.AddDate(0, 0, -40).Format("2006-01-02") // 40 days old (should be deleted)
	recentDate := now.AddDate(0, 0, -20).Format("2006-01-02") // 20 days old (should be kept)

	oldFilePath := filepath.Join(tempDir, oldDate+".json")
	recentFilePath := filepath.Join(tempDir, recentDate+".json")

	testContent := `{"date":"test","results":[]}`
	if err := os.WriteFile(oldFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create old test file: %v", err)
	}
	if err := os.WriteFile(recentFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create recent test file: %v", err)
	}

	// Trigger cleanup
	report, err := manager.TriggerManualCleanup()
	if err != nil {
		t.Fatalf("Failed to trigger cleanup: %v", err)
	}

	if report.FilesDeleted != 1 {
		t.Errorf("Expected 1 file deleted, got %d", report.FilesDeleted)
	}

	if report.SpaceFreed == 0 {
		t.Error("Expected non-zero space freed")
	}

	// Verify old file was deleted
	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Error("Old file should have been deleted")
	}

	// Verify recent file still exists
	if _, err := os.Stat(recentFilePath); os.IsNotExist(err) {
		t.Error("Recent file should not have been deleted")
	}
}

func TestCleanupProtectsCurrentDay(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Set retention to 0 days (extreme case)
	policy := &RetentionPolicy{
		RawDataDays:        7, // Minimum allowed
		AggregatedDataDays: 30,
		ConfigBackupDays:   7,
		AutoCleanupEnabled: false,
		CleanupTime:        "02:00",
	}

	manager, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create today's file and yesterday's file
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	todayFilePath := filepath.Join(tempDir, today+".json")
	yesterdayFilePath := filepath.Join(tempDir, yesterday+".json")

	testContent := `{"date":"test","results":[]}`
	if err := os.WriteFile(todayFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create today's test file: %v", err)
	}
	if err := os.WriteFile(yesterdayFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create yesterday's test file: %v", err)
	}

	// Trigger cleanup
	_, err = manager.TriggerManualCleanup()
	if err != nil {
		t.Fatalf("Failed to trigger cleanup: %v", err)
	}

	// Verify today's file still exists (should be protected)
	if _, err := os.Stat(todayFilePath); os.IsNotExist(err) {
		t.Error("Today's file should be protected from deletion")
	}

	// Yesterday's file should also exist (within retention period)
	if _, err := os.Stat(yesterdayFilePath); os.IsNotExist(err) {
		t.Error("Yesterday's file should not have been deleted")
	}
}

func TestGetCleanupHistory(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	policy := DefaultRetentionPolicy()
	policy.AutoCleanupEnabled = false

	manager, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Initial history should be empty
	history := manager.GetCleanupHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}

	// Trigger a cleanup
	_, err = manager.TriggerManualCleanup()
	if err != nil {
		t.Fatalf("Failed to trigger cleanup: %v", err)
	}

	// History should now have one entry
	history = manager.GetCleanupHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}

	if history[0].Success != true {
		t.Error("Expected cleanup to be successful")
	}
}

func TestCleanupHistoryPersistence(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	policy := DefaultRetentionPolicy()
	policy.AutoCleanupEnabled = false

	// Create manager and trigger cleanup
	manager1, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create first manager: %v", err)
	}

	_, err = manager1.TriggerManualCleanup()
	if err != nil {
		t.Fatalf("Failed to trigger cleanup: %v", err)
	}

	// Close first manager
	if err := manager1.Close(); err != nil {
		t.Fatalf("Failed to close first manager: %v", err)
	}

	// Create new manager with same directory
	manager2, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}
	defer manager2.Close()

	// History should be loaded
	history := manager2.GetCleanupHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 history entry after reload, got %d", len(history))
	}
}

func TestCleanupHistoryMaxLength(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	policy := DefaultRetentionPolicy()
	policy.AutoCleanupEnabled = false

	manager, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Set max history length to something small for testing
	manager.maxHistoryLength = 5

	// Trigger more cleanups than max length
	for i := 0; i < 10; i++ {
		_, err = manager.TriggerManualCleanup()
		if err != nil {
			t.Fatalf("Failed to trigger cleanup %d: %v", i, err)
		}
	}

	// History should be capped at max length
	history := manager.GetCleanupHistory()
	if len(history) != 5 {
		t.Errorf("Expected history capped at 5 entries, got %d", len(history))
	}
}

func TestStorageStatsWithEmptyDirectory(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	manager, err := NewManager(ctx, tempDir, DefaultRetentionPolicy())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	stats, err := manager.GetStorageStats()
	if err != nil {
		t.Fatalf("Failed to get storage stats: %v", err)
	}

	if stats.TotalFiles != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", stats.TotalFiles)
	}

	if stats.TotalSizeBytes != 0 {
		t.Errorf("Expected 0 bytes in empty directory, got %d", stats.TotalSizeBytes)
	}
}

func TestCleanupWithInvalidFilenames(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	policy := &RetentionPolicy{
		RawDataDays:        30,
		AggregatedDataDays: 180,
		ConfigBackupDays:   14,
		AutoCleanupEnabled: false,
		CleanupTime:        "02:00",
	}

	manager, err := NewManager(ctx, tempDir, policy)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create files with invalid names
	invalidFiles := []string{
		"invalid.json",
		"not-a-date.json",
		"2025-13-01.json", // Invalid month
		"README.md",
	}

	for _, filename := range invalidFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Trigger cleanup - should not crash
	report, err := manager.TriggerManualCleanup()
	if err != nil {
		t.Fatalf("Failed to trigger cleanup: %v", err)
	}

	// Should not delete invalid files
	if report.FilesDeleted != 0 {
		t.Errorf("Expected 0 files deleted, got %d", report.FilesDeleted)
	}

	// Verify files still exist
	for _, filename := range invalidFiles {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %s should not have been deleted", filename)
		}
	}
}

func TestCleanupReportSerialization(t *testing.T) {
	report := &CleanupReport{
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(5 * time.Second),
		FilesDeleted: 10,
		SpaceFreed:   1024000,
		ErrorCount:   2,
		Errors:       []string{"error 1", "error 2"},
	}

	// Test JSON serialization
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal cleanup report: %v", err)
	}

	var decoded CleanupReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal cleanup report: %v", err)
	}

	if decoded.FilesDeleted != report.FilesDeleted {
		t.Errorf("Expected FilesDeleted %d, got %d", report.FilesDeleted, decoded.FilesDeleted)
	}

	if decoded.SpaceFreed != report.SpaceFreed {
		t.Errorf("Expected SpaceFreed %d, got %d", report.SpaceFreed, decoded.SpaceFreed)
	}
}

func TestStorageStatsSerialization(t *testing.T) {
	stats := &StorageStats{
		TotalFiles:     100,
		TotalSizeBytes: 1024000,
		OldestDataDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		NewestDataDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DaysOfData:     365,
	}

	// Test JSON serialization
	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal storage stats: %v", err)
	}

	var decoded StorageStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal storage stats: %v", err)
	}

	if decoded.TotalFiles != stats.TotalFiles {
		t.Errorf("Expected TotalFiles %d, got %d", stats.TotalFiles, decoded.TotalFiles)
	}

	if decoded.DaysOfData != stats.DaysOfData {
		t.Errorf("Expected DaysOfData %d, got %d", stats.DaysOfData, decoded.DaysOfData)
	}
}
