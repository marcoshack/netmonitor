package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/marcoshack/netmonitor/internal/storage"
)

// Manager handles data export operations
type Manager struct {
	ctx           context.Context
	storage       *storage.Manager
	exportDir     string
	jobs          map[string]*ExportJob
	jobsMutex     sync.RWMutex
	history       []*ExportJob
	historyMutex  sync.RWMutex
	maxHistory    int
	cancelFuncs   map[string]context.CancelFunc
	cancelMutex   sync.RWMutex
}

// NewManager creates a new export manager
func NewManager(ctx context.Context, storage *storage.Manager, exportDir string) (*Manager, error) {
	// Ensure export directory exists
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create export directory: %w", err)
	}

	return &Manager{
		ctx:         ctx,
		storage:     storage,
		exportDir:   exportDir,
		jobs:        make(map[string]*ExportJob),
		history:     make([]*ExportJob, 0),
		maxHistory:  100,
		cancelFuncs: make(map[string]context.CancelFunc),
	}, nil
}

// CreateExport creates a new export job and starts it in the background
func (m *Manager) CreateExport(request ExportRequest) (*ExportJob, error) {
	// Validate request
	if err := m.validateRequest(&request); err != nil {
		return nil, fmt.Errorf("invalid export request: %w", err)
	}

	// Create job
	job := &ExportJob{
		ID:        uuid.New().String(),
		Request:   request,
		Status:    StatusPending,
		Progress:  0.0,
		StartTime: time.Now(),
	}

	// Store job
	m.jobsMutex.Lock()
	m.jobs[job.ID] = job
	m.jobsMutex.Unlock()

	// Start export in background
	jobCtx, cancel := context.WithCancel(m.ctx)
	m.cancelMutex.Lock()
	m.cancelFuncs[job.ID] = cancel
	m.cancelMutex.Unlock()

	go m.executeExport(jobCtx, job)

	log.Ctx(m.ctx).Info().
		Str("job_id", job.ID).
		Str("format", request.Format).
		Bool("compressed", request.Compressed).
		Msg("Export job created")

	return job, nil
}

// GetExportStatus returns the status of an export job
func (m *Manager) GetExportStatus(jobID string) (*ExportStatus, error) {
	m.jobsMutex.RLock()
	job, exists := m.jobs[jobID]
	m.jobsMutex.RUnlock()

	if !exists {
		// Check history
		m.historyMutex.RLock()
		for _, histJob := range m.history {
			if histJob.ID == jobID {
				job = histJob
				break
			}
		}
		m.historyMutex.RUnlock()

		if job == nil {
			return nil, fmt.Errorf("export job not found: %s", jobID)
		}
	}

	status := &ExportStatus{
		Job:          job,
		CurrentPhase: m.getCurrentPhase(job),
	}

	return status, nil
}

// CancelExport cancels a running export job
func (m *Manager) CancelExport(jobID string) error {
	m.jobsMutex.RLock()
	job, exists := m.jobs[jobID]
	m.jobsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("export job not found: %s", jobID)
	}

	if job.Status != StatusPending && job.Status != StatusRunning {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	// Cancel the job context
	m.cancelMutex.Lock()
	if cancel, ok := m.cancelFuncs[jobID]; ok {
		cancel()
		delete(m.cancelFuncs, jobID)
	}
	m.cancelMutex.Unlock()

	// Update job status
	m.jobsMutex.Lock()
	job.Status = StatusCancelled
	now := time.Now()
	job.EndTime = &now
	m.jobsMutex.Unlock()

	log.Ctx(m.ctx).Info().Str("job_id", jobID).Msg("Export job cancelled")

	return nil
}

// GetExportHistory returns the export job history
func (m *Manager) GetExportHistory() []*ExportJob {
	m.historyMutex.RLock()
	defer m.historyMutex.RUnlock()

	// Return a copy
	history := make([]*ExportJob, len(m.history))
	copy(history, m.history)
	return history
}

// CleanupOldExports removes export files older than the specified number of days
func (m *Manager) CleanupOldExports(retentionDays int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	removed := 0

	// Clean up history
	m.historyMutex.Lock()
	newHistory := make([]*ExportJob, 0)
	for _, job := range m.history {
		if job.EndTime != nil && job.EndTime.Before(cutoff) {
			// Remove file if it exists
			if job.FilePath != "" {
				if err := os.Remove(job.FilePath); err != nil && !os.IsNotExist(err) {
					log.Ctx(m.ctx).Warn().
						Err(err).
						Str("file", job.FilePath).
						Msg("Failed to remove export file")
				} else {
					removed++
				}
			}
		} else {
			newHistory = append(newHistory, job)
		}
	}
	m.history = newHistory
	m.historyMutex.Unlock()

	log.Ctx(m.ctx).Info().
		Int("removed", removed).
		Int("retention_days", retentionDays).
		Msg("Cleaned up old export files")

	return removed, nil
}

// validateRequest validates an export request
func (m *Manager) validateRequest(request *ExportRequest) error {
	// Validate format
	if request.Format != FormatCSV && request.Format != FormatJSON {
		return fmt.Errorf("invalid format: %s (must be 'csv' or 'json')", request.Format)
	}

	// Validate date range
	if request.StartDate.After(request.EndDate) {
		return fmt.Errorf("start date must be before or equal to end date")
	}

	// Validate at least one data type is requested
	if !request.IncludeRaw && !request.IncludeAgg {
		return fmt.Errorf("must include at least one of: raw data or aggregated data")
	}

	// Validate CSV columns if specified
	if request.Format == FormatCSV && len(request.Columns) > 0 {
		validColumns := make(map[string]bool)
		for _, col := range AllCSVColumns() {
			validColumns[col] = true
		}
		for _, col := range request.Columns {
			if !validColumns[col] {
				return fmt.Errorf("invalid CSV column: %s", col)
			}
		}
	}

	return nil
}

// executeExport performs the actual export operation
func (m *Manager) executeExport(ctx context.Context, job *ExportJob) {
	// Update status to running
	m.jobsMutex.Lock()
	job.Status = StatusRunning
	m.jobsMutex.Unlock()

	// Generate filename
	timestamp := time.Now().Format("20060102-150405")
	var filename string
	if job.Request.Compressed {
		filename = fmt.Sprintf("export-%s-%s.zip", timestamp, job.ID[:8])
	} else {
		ext := job.Request.Format
		filename = fmt.Sprintf("export-%s-%s.%s", timestamp, job.ID[:8], ext)
	}
	filePath := filepath.Join(m.exportDir, filename)

	// Perform export
	var err error
	switch job.Request.Format {
	case FormatCSV:
		err = m.exportCSV(ctx, job, filePath)
	case FormatJSON:
		err = m.exportJSON(ctx, job, filePath)
	default:
		err = fmt.Errorf("unsupported format: %s", job.Request.Format)
	}

	// Update job status
	m.jobsMutex.Lock()
	now := time.Now()
	job.EndTime = &now

	if err != nil {
		if ctx.Err() == context.Canceled {
			job.Status = StatusCancelled
		} else {
			job.Status = StatusFailed
			job.Error = err.Error()
		}
	} else {
		job.Status = StatusCompleted
		job.FilePath = filePath
		job.Progress = 1.0

		// Get file size
		if info, err := os.Stat(filePath); err == nil {
			job.FileSize = info.Size()
		}
	}
	m.jobsMutex.Unlock()

	// Move to history and remove from active jobs
	m.moveToHistory(job)

	// Clean up cancel function
	m.cancelMutex.Lock()
	delete(m.cancelFuncs, job.ID)
	m.cancelMutex.Unlock()

	log.Ctx(m.ctx).Info().
		Str("job_id", job.ID).
		Str("status", job.Status).
		Str("file", filePath).
		Msg("Export job completed")
}

// moveToHistory moves a job from active to history
func (m *Manager) moveToHistory(job *ExportJob) {
	// Remove from active jobs
	m.jobsMutex.Lock()
	delete(m.jobs, job.ID)
	m.jobsMutex.Unlock()

	// Add to history
	m.historyMutex.Lock()
	m.history = append(m.history, job)

	// Trim history if too long
	if len(m.history) > m.maxHistory {
		m.history = m.history[len(m.history)-m.maxHistory:]
	}
	m.historyMutex.Unlock()
}

// getCurrentPhase returns a human-readable description of the current phase
func (m *Manager) getCurrentPhase(job *ExportJob) string {
	switch job.Status {
	case StatusPending:
		return "Waiting to start"
	case StatusRunning:
		if job.Progress < 0.1 {
			return "Initializing export"
		} else if job.Progress < 0.5 {
			return "Reading data"
		} else if job.Progress < 0.9 {
			return "Writing export file"
		} else {
			return "Finalizing"
		}
	case StatusCompleted:
		return "Completed successfully"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// GetActiveJobs returns all currently active (pending or running) export jobs
func (m *Manager) GetActiveJobs() []*ExportJob {
	m.jobsMutex.RLock()
	defer m.jobsMutex.RUnlock()

	jobs := make([]*ExportJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// Close gracefully shuts down the export manager
func (m *Manager) Close() error {
	// Cancel all active jobs
	m.cancelMutex.Lock()
	for _, cancel := range m.cancelFuncs {
		cancel()
	}
	m.cancelFuncs = make(map[string]context.CancelFunc)
	m.cancelMutex.Unlock()

	log.Ctx(m.ctx).Info().Msg("Export manager closed")
	return nil
}
