package retention

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Manager handles data retention and cleanup operations
type Manager struct {
	ctx              context.Context
	dataDir          string
	policy           *RetentionPolicy
	mutex            sync.RWMutex
	scheduler        *time.Ticker
	stopChan         chan struct{}
	cleanupHistory   []*CleanupOperation
	maxHistoryLength int
}

// NewManager creates a new retention manager
func NewManager(ctx context.Context, dataDir string, policy *RetentionPolicy) (*Manager, error) {
	if policy == nil {
		policy = DefaultRetentionPolicy()
	}

	// Validate policy
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid retention policy: %w", err)
	}

	m := &Manager{
		ctx:              ctx,
		dataDir:          dataDir,
		policy:           policy,
		stopChan:         make(chan struct{}),
		cleanupHistory:   make([]*CleanupOperation, 0),
		maxHistoryLength: 100, // Keep last 100 cleanup operations
	}

	// Load cleanup history from disk
	if err := m.loadCleanupHistory(); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to load cleanup history, starting fresh")
	}

	// Start automatic cleanup scheduler if enabled
	if policy.AutoCleanupEnabled {
		if err := m.startScheduler(); err != nil {
			return nil, fmt.Errorf("failed to start cleanup scheduler: %w", err)
		}
	}

	log.Ctx(ctx).Info().
		Int("raw_data_days", policy.RawDataDays).
		Int("aggregated_data_days", policy.AggregatedDataDays).
		Bool("auto_cleanup", policy.AutoCleanupEnabled).
		Str("cleanup_time", policy.CleanupTime).
		Msg("Retention manager initialized")

	return m, nil
}

// UpdatePolicy updates the retention policy
func (m *Manager) UpdatePolicy(policy *RetentionPolicy) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate new policy
	if err := policy.Validate(); err != nil {
		return fmt.Errorf("invalid retention policy: %w", err)
	}

	// Check if scheduler needs to be restarted
	restartScheduler := m.policy.AutoCleanupEnabled != policy.AutoCleanupEnabled ||
		m.policy.CleanupTime != policy.CleanupTime

	m.policy = policy

	log.Ctx(m.ctx).Info().
		Int("raw_data_days", policy.RawDataDays).
		Bool("auto_cleanup", policy.AutoCleanupEnabled).
		Msg("Retention policy updated")

	// Restart scheduler if needed
	if restartScheduler {
		m.stopScheduler()
		if policy.AutoCleanupEnabled {
			if err := m.startScheduler(); err != nil {
				return fmt.Errorf("failed to restart cleanup scheduler: %w", err)
			}
		}
	}

	return nil
}

// GetPolicy returns the current retention policy
func (m *Manager) GetPolicy() *RetentionPolicy {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy
	policyCopy := *m.policy
	return &policyCopy
}

// TriggerManualCleanup performs an immediate cleanup operation
func (m *Manager) TriggerManualCleanup() (*CleanupReport, error) {
	log.Ctx(m.ctx).Info().Msg("Manual cleanup triggered")

	report := m.performCleanup()

	// Add to history
	m.addToHistory(&CleanupOperation{
		Timestamp:    report.StartTime,
		FilesDeleted: report.FilesDeleted,
		SpaceFreed:   report.SpaceFreed,
		Duration:     report.EndTime.Sub(report.StartTime).Milliseconds(),
		Success:      report.ErrorCount == 0,
		ErrorMessage: m.formatErrorMessage(report.Errors),
	})

	return report, nil
}

// GetStorageStats returns current storage statistics
func (m *Manager) GetStorageStats() (*StorageStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &StorageStats{
		TotalFiles:     0,
		TotalSizeBytes: 0,
		OldestDataDate: time.Now(),
		NewestDataDate: time.Time{},
		DaysOfData:     0,
	}

	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil // Return empty stats if directory doesn't exist
		}
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var oldestDate, newestDate time.Time
	firstFile := true

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		stats.TotalFiles++

		// Get file size
		if info, err := entry.Info(); err == nil {
			stats.TotalSizeBytes += info.Size()
		}

		// Parse date from filename (YYYY-MM-DD.json)
		filename := entry.Name()
		if len(filename) < 15 { // YYYY-MM-DD.json = 15 chars
			continue
		}

		dateStr := filename[:10] // Extract YYYY-MM-DD
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if firstFile {
			oldestDate = fileDate
			newestDate = fileDate
			firstFile = false
		} else {
			if fileDate.Before(oldestDate) {
				oldestDate = fileDate
			}
			if fileDate.After(newestDate) {
				newestDate = fileDate
			}
		}
	}

	if !firstFile {
		stats.OldestDataDate = oldestDate
		stats.NewestDataDate = newestDate
		stats.DaysOfData = int(newestDate.Sub(oldestDate).Hours()/24) + 1
	}

	return stats, nil
}

// GetCleanupHistory returns the cleanup operation history
func (m *Manager) GetCleanupHistory() []*CleanupOperation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy
	history := make([]*CleanupOperation, len(m.cleanupHistory))
	copy(history, m.cleanupHistory)

	return history
}

// Close stops the retention manager
func (m *Manager) Close() error {
	log.Ctx(m.ctx).Info().Msg("Retention manager closing")

	m.stopScheduler()

	// Save cleanup history
	if err := m.saveCleanupHistory(); err != nil {
		log.Ctx(m.ctx).Error().Err(err).Msg("Failed to save cleanup history")
	}

	return nil
}

// performCleanup executes the cleanup operation
func (m *Manager) performCleanup() *CleanupReport {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	report := &CleanupReport{
		StartTime:    time.Now(),
		FilesDeleted: 0,
		SpaceFreed:   0,
		ErrorCount:   0,
		Errors:       make([]string, 0),
	}

	// Calculate cutoff dates
	now := time.Now()
	rawDataCutoff := now.AddDate(0, 0, -m.policy.RawDataDays)

	// Never delete current day's data
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if rawDataCutoff.After(today.AddDate(0, 0, -1)) {
		rawDataCutoff = today.AddDate(0, 0, -1)
	}

	log.Ctx(m.ctx).Info().
		Time("cutoff_date", rawDataCutoff).
		Int("retention_days", m.policy.RawDataDays).
		Msg("Starting cleanup operation")

	// Read data directory
	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		report.ErrorCount++
		report.Errors = append(report.Errors, fmt.Sprintf("Failed to read data directory: %v", err))
		report.EndTime = time.Now()
		return report
	}

	// Process each file
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filename := entry.Name()

		// Parse date from filename (YYYY-MM-DD.json)
		if len(filename) < 15 {
			continue
		}

		dateStr := filename[:10]
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Ctx(m.ctx).Warn().
				Str("filename", filename).
				Err(err).
				Msg("Failed to parse date from filename, skipping")
			continue
		}

		// Delete if older than cutoff
		if fileDate.Before(rawDataCutoff) {
			filePath := filepath.Join(m.dataDir, filename)

			// Get file size before deletion
			var fileSize int64
			if info, err := entry.Info(); err == nil {
				fileSize = info.Size()
			}

			// Delete file
			if err := os.Remove(filePath); err != nil {
				report.ErrorCount++
				report.Errors = append(report.Errors,
					fmt.Sprintf("Failed to delete %s: %v", filename, err))
				log.Ctx(m.ctx).Error().
					Str("filename", filename).
					Err(err).
					Msg("Failed to delete old file")
				continue
			}

			report.FilesDeleted++
			report.SpaceFreed += fileSize

			log.Ctx(m.ctx).Info().
				Str("filename", filename).
				Str("file_date", dateStr).
				Int64("size_bytes", fileSize).
				Msg("Deleted old data file")
		}
	}

	report.EndTime = time.Now()

	log.Ctx(m.ctx).Info().
		Int("files_deleted", report.FilesDeleted).
		Int64("space_freed_mb", report.SpaceFreed/1024/1024).
		Int("errors", report.ErrorCount).
		Dur("duration", report.EndTime.Sub(report.StartTime)).
		Msg("Cleanup operation completed")

	return report
}

// startScheduler starts the automatic cleanup scheduler
func (m *Manager) startScheduler() error {
	nextCleanup, err := m.policy.GetNextCleanupTime()
	if err != nil {
		return fmt.Errorf("failed to calculate next cleanup time: %w", err)
	}

	log.Ctx(m.ctx).Info().
		Time("next_cleanup", nextCleanup).
		Msg("Cleanup scheduler started")

	go m.schedulerLoop()

	return nil
}

// stopScheduler stops the automatic cleanup scheduler
func (m *Manager) stopScheduler() {
	if m.scheduler != nil {
		m.scheduler.Stop()
		m.scheduler = nil
	}

	select {
	case m.stopChan <- struct{}{}:
	default:
	}
}

// schedulerLoop runs the cleanup scheduler
func (m *Manager) schedulerLoop() {
	for {
		nextCleanup, err := m.policy.GetNextCleanupTime()
		if err != nil {
			log.Ctx(m.ctx).Error().Err(err).Msg("Failed to calculate next cleanup time")
			time.Sleep(1 * time.Hour) // Retry in 1 hour
			continue
		}

		now := time.Now()
		waitDuration := nextCleanup.Sub(now)

		log.Ctx(m.ctx).Debug().
			Time("next_cleanup", nextCleanup).
			Dur("wait_duration", waitDuration).
			Msg("Waiting for next cleanup")

		timer := time.NewTimer(waitDuration)

		select {
		case <-timer.C:
			// Execute cleanup
			log.Ctx(m.ctx).Info().Msg("Automatic cleanup started")
			report := m.performCleanup()

			// Add to history
			m.addToHistory(&CleanupOperation{
				Timestamp:    report.StartTime,
				FilesDeleted: report.FilesDeleted,
				SpaceFreed:   report.SpaceFreed,
				Duration:     report.EndTime.Sub(report.StartTime).Milliseconds(),
				Success:      report.ErrorCount == 0,
				ErrorMessage: m.formatErrorMessage(report.Errors),
			})

		case <-m.stopChan:
			timer.Stop()
			log.Ctx(m.ctx).Info().Msg("Cleanup scheduler stopped")
			return
		}
	}
}

// addToHistory adds a cleanup operation to the history
func (m *Manager) addToHistory(operation *CleanupOperation) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.cleanupHistory = append(m.cleanupHistory, operation)

	// Trim history if too long
	if len(m.cleanupHistory) > m.maxHistoryLength {
		m.cleanupHistory = m.cleanupHistory[len(m.cleanupHistory)-m.maxHistoryLength:]
	}

	// Save to disk
	if err := m.saveCleanupHistory(); err != nil {
		log.Ctx(m.ctx).Error().Err(err).Msg("Failed to save cleanup history")
	}
}

// loadCleanupHistory loads cleanup history from disk
func (m *Manager) loadCleanupHistory() error {
	historyPath := filepath.Join(m.dataDir, "..", "cleanup_history.json")

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No history file yet
		}
		return fmt.Errorf("failed to read history file: %w", err)
	}

	var history []*CleanupOperation
	if err := json.Unmarshal(data, &history); err != nil {
		return fmt.Errorf("failed to parse history file: %w", err)
	}

	m.cleanupHistory = history

	log.Ctx(m.ctx).Info().
		Int("entries", len(history)).
		Msg("Cleanup history loaded")

	return nil
}

// saveCleanupHistory saves cleanup history to disk
func (m *Manager) saveCleanupHistory() error {
	historyPath := filepath.Join(m.dataDir, "..", "cleanup_history.json")

	data, err := json.MarshalIndent(m.cleanupHistory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Use atomic write pattern
	tempFile := historyPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp history file: %w", err)
	}

	if err := os.Rename(tempFile, historyPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp history file: %w", err)
	}

	return nil
}

// formatErrorMessage formats error list into a single string
func (m *Manager) formatErrorMessage(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	if len(errors) == 1 {
		return errors[0]
	}
	return fmt.Sprintf("%d errors occurred (first: %s)", len(errors), errors[0])
}
