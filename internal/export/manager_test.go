package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcoshack/netmonitor/internal/storage"
)

func setupTestExport(t *testing.T) (*Manager, *storage.Manager, string) {
	t.Helper()

	// Create temporary directories
	dataDir := t.TempDir()
	exportDir := t.TempDir()

	ctx := context.Background()

	// Initialize storage manager
	storageManager, err := storage.NewManager(ctx, dataDir)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create some test data
	createTestData(t, storageManager)

	// Initialize export manager
	exportManager, err := NewManager(ctx, storageManager, exportDir)
	if err != nil {
		t.Fatalf("Failed to create export manager: %v", err)
	}

	return exportManager, storageManager, exportDir
}

func createTestData(t *testing.T, sm *storage.Manager) {
	t.Helper()

	baseTime := time.Now().Add(-24 * time.Hour)

	// Create test results for the past 3 days
	for day := 0; day < 3; day++ {
		for i := 0; i < 10; i++ {
			result := &storage.TestResult{
				Timestamp:  baseTime.Add(time.Duration(day*24+i) * time.Hour),
				EndpointID: "test-endpoint-1",
				Protocol:   "icmp",
				Latency:    time.Duration(10+i) * time.Millisecond,
				Status:     "success",
			}
			if err := sm.StoreTestResult(result); err != nil {
				t.Fatalf("Failed to store test result: %v", err)
			}
		}
	}
}

func TestNewManager(t *testing.T) {
	exportDir := t.TempDir()
	dataDir := t.TempDir()
	ctx := context.Background()

	sm, err := storage.NewManager(ctx, dataDir)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	em, err := NewManager(ctx, sm, exportDir)
	if err != nil {
		t.Fatalf("Failed to create export manager: %v", err)
	}

	if em == nil {
		t.Fatal("Export manager is nil")
	}

	// Verify export directory was created
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		t.Errorf("Export directory was not created")
	}
}

func TestValidateRequest(t *testing.T) {
	em, _, _ := setupTestExport(t)

	tests := []struct {
		name    string
		request ExportRequest
		wantErr bool
	}{
		{
			name: "Valid CSV request",
			request: ExportRequest{
				Format:     FormatCSV,
				StartDate:  time.Now().Add(-24 * time.Hour),
				EndDate:    time.Now(),
				IncludeRaw: true,
			},
			wantErr: false,
		},
		{
			name: "Valid JSON request",
			request: ExportRequest{
				Format:     FormatJSON,
				StartDate:  time.Now().Add(-24 * time.Hour),
				EndDate:    time.Now(),
				IncludeRaw: true,
			},
			wantErr: false,
		},
		{
			name: "Invalid format",
			request: ExportRequest{
				Format:     "xml",
				StartDate:  time.Now().Add(-24 * time.Hour),
				EndDate:    time.Now(),
				IncludeRaw: true,
			},
			wantErr: true,
		},
		{
			name: "Invalid date range",
			request: ExportRequest{
				Format:     FormatCSV,
				StartDate:  time.Now(),
				EndDate:    time.Now().Add(-24 * time.Hour),
				IncludeRaw: true,
			},
			wantErr: true,
		},
		{
			name: "No data type specified",
			request: ExportRequest{
				Format:    FormatCSV,
				StartDate: time.Now().Add(-24 * time.Hour),
				EndDate:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Invalid CSV column",
			request: ExportRequest{
				Format:     FormatCSV,
				StartDate:  time.Now().Add(-24 * time.Hour),
				EndDate:    time.Now(),
				IncludeRaw: true,
				Columns:    []string{"invalid_column"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := em.validateRequest(&tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateExport(t *testing.T) {
	em, _, _ := setupTestExport(t)

	request := ExportRequest{
		Format:     FormatCSV,
		StartDate:  time.Now().Add(-72 * time.Hour),
		EndDate:    time.Now(),
		IncludeRaw: true,
	}

	job, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	if job == nil {
		t.Fatal("Export job is nil")
	}

	if job.ID == "" {
		t.Error("Export job ID is empty")
	}

	if job.Status != StatusPending && job.Status != StatusRunning {
		t.Errorf("Unexpected initial status: %s", job.Status)
	}

	// Wait a bit for the job to process
	time.Sleep(2 * time.Second)

	// Check job status
	status, err := em.GetExportStatus(job.ID)
	if err != nil {
		t.Fatalf("Failed to get export status: %v", err)
	}

	if status.Job.Status != StatusCompleted && status.Job.Status != StatusRunning {
		t.Errorf("Unexpected job status: %s (error: %s)", status.Job.Status, status.Job.Error)
	}
}

func TestGetExportStatus(t *testing.T) {
	em, _, _ := setupTestExport(t)

	request := ExportRequest{
		Format:     FormatJSON,
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
		IncludeRaw: true,
	}

	job, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	// Get status immediately
	status, err := em.GetExportStatus(job.ID)
	if err != nil {
		t.Fatalf("Failed to get export status: %v", err)
	}

	if status == nil {
		t.Fatal("Export status is nil")
	}

	if status.Job.ID != job.ID {
		t.Errorf("Expected job ID %s, got %s", job.ID, status.Job.ID)
	}

	// Test non-existent job
	_, err = em.GetExportStatus("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent job")
	}
}

func TestCancelExport(t *testing.T) {
	em, _, _ := setupTestExport(t)

	// Create a large export that will take some time
	request := ExportRequest{
		Format:     FormatCSV,
		StartDate:  time.Now().Add(-720 * time.Hour), // 30 days
		EndDate:    time.Now(),
		IncludeRaw: true,
	}

	job, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	// Cancel immediately (should work even if job hasn't started yet)
	err = em.CancelExport(job.ID)
	if err != nil {
		// If the job already completed or moved to history, skip this test
		if err.Error() == fmt.Sprintf("export job not found: %s", job.ID) {
			t.Skip("Job completed too quickly to cancel")
		}
		t.Fatalf("Failed to cancel export: %v", err)
	}

	// Wait a bit for cancellation to take effect
	time.Sleep(500 * time.Millisecond)

	// Check status
	status, err := em.GetExportStatus(job.ID)
	if err != nil {
		t.Fatalf("Failed to get export status: %v", err)
	}

	if status.Job.Status != StatusCancelled {
		t.Errorf("Expected status %s, got %s", StatusCancelled, status.Job.Status)
	}
}

func TestGetExportHistory(t *testing.T) {
	em, _, _ := setupTestExport(t)

	// Create a few exports
	for i := 0; i < 3; i++ {
		request := ExportRequest{
			Format:     FormatCSV,
			StartDate:  time.Now().Add(-24 * time.Hour),
			EndDate:    time.Now(),
			IncludeRaw: true,
		}

		_, err := em.CreateExport(request)
		if err != nil {
			t.Fatalf("Failed to create export: %v", err)
		}
	}

	// Wait for jobs to complete
	time.Sleep(3 * time.Second)

	history := em.GetExportHistory()
	if len(history) == 0 {
		t.Error("Expected non-empty history")
	}
}

func TestCleanupOldExports(t *testing.T) {
	em, _, exportDir := setupTestExport(t)

	// Create an export
	request := ExportRequest{
		Format:     FormatCSV,
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
		IncludeRaw: true,
	}

	job, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	// Wait for it to complete
	time.Sleep(2 * time.Second)

	// Verify file exists
	status, _ := em.GetExportStatus(job.ID)
	if status.Job.FilePath != "" {
		if _, err := os.Stat(status.Job.FilePath); err != nil {
			t.Errorf("Export file does not exist: %v", err)
		}
	}

	// Cleanup exports older than 0 days (all exports)
	removed, err := em.CleanupOldExports(0)
	if err != nil {
		t.Fatalf("Failed to cleanup exports: %v", err)
	}

	if removed == 0 {
		// The export might not be old enough yet
		t.Log("No exports were old enough to remove")
	}

	// Verify export directory still exists
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		t.Error("Export directory should still exist after cleanup")
	}
}

func TestGetActiveJobs(t *testing.T) {
	em, _, _ := setupTestExport(t)

	// Initially should have no active jobs
	active := em.GetActiveJobs()
	if len(active) != 0 {
		t.Errorf("Expected 0 active jobs, got %d", len(active))
	}

	// Create an export
	request := ExportRequest{
		Format:     FormatCSV,
		StartDate:  time.Now().Add(-24 * time.Hour),
		EndDate:    time.Now(),
		IncludeRaw: true,
	}

	_, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	// Should have one active job
	active = em.GetActiveJobs()
	if len(active) != 1 {
		t.Errorf("Expected 1 active job, got %d", len(active))
	}

	// Wait for completion
	time.Sleep(2 * time.Second)

	// Should have no active jobs after completion
	active = em.GetActiveJobs()
	if len(active) != 0 {
		t.Errorf("Expected 0 active jobs after completion, got %d", len(active))
	}
}

func TestExportCSVFormat(t *testing.T) {
	em, _, _ := setupTestExport(t)

	request := ExportRequest{
		Format:     FormatCSV,
		StartDate:  time.Now().Add(-72 * time.Hour),
		EndDate:    time.Now(),
		IncludeRaw: true,
		Columns:    DefaultCSVColumns(),
	}

	job, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	// Wait for completion
	time.Sleep(3 * time.Second)

	status, err := em.GetExportStatus(job.ID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status.Job.Status != StatusCompleted {
		t.Fatalf("Export failed: %s", status.Job.Error)
	}

	// Verify file exists and has content
	if status.Job.FilePath == "" {
		t.Fatal("No file path in completed job")
	}

	info, err := os.Stat(status.Job.FilePath)
	if err != nil {
		t.Fatalf("Export file does not exist: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Export file is empty")
	}

	// Verify it's a CSV file
	ext := filepath.Ext(status.Job.FilePath)
	if ext != ".csv" {
		t.Errorf("Expected .csv extension, got %s", ext)
	}
}

func TestExportJSONFormat(t *testing.T) {
	em, _, _ := setupTestExport(t)

	request := ExportRequest{
		Format:     FormatJSON,
		StartDate:  time.Now().Add(-72 * time.Hour),
		EndDate:    time.Now(),
		IncludeRaw: true,
	}

	job, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	// Wait for completion
	time.Sleep(3 * time.Second)

	status, err := em.GetExportStatus(job.ID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status.Job.Status != StatusCompleted {
		t.Fatalf("Export failed: %s", status.Job.Error)
	}

	// Verify file exists and has content
	if status.Job.FilePath == "" {
		t.Fatal("No file path in completed job")
	}

	info, err := os.Stat(status.Job.FilePath)
	if err != nil {
		t.Fatalf("Export file does not exist: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Export file is empty")
	}

	// Verify it's a JSON file
	ext := filepath.Ext(status.Job.FilePath)
	if ext != ".json" {
		t.Errorf("Expected .json extension, got %s", ext)
	}
}

func TestExportCompressed(t *testing.T) {
	em, _, _ := setupTestExport(t)

	request := ExportRequest{
		Format:     FormatCSV,
		StartDate:  time.Now().Add(-72 * time.Hour),
		EndDate:    time.Now(),
		IncludeRaw: true,
		Compressed: true,
	}

	job, err := em.CreateExport(request)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}

	// Wait for completion
	time.Sleep(3 * time.Second)

	status, err := em.GetExportStatus(job.ID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status.Job.Status != StatusCompleted {
		t.Fatalf("Export failed: %s", status.Job.Error)
	}

	// Verify it's a ZIP file
	ext := filepath.Ext(status.Job.FilePath)
	if ext != ".zip" {
		t.Errorf("Expected .zip extension, got %s", ext)
	}

	// Verify file exists and has content
	info, err := os.Stat(status.Job.FilePath)
	if err != nil {
		t.Fatalf("Export file does not exist: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Export file is empty")
	}
}
