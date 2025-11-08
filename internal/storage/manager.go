package storage

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

// Manager handles data storage operations
type Manager struct {
	dataDir string
	ctx     context.Context
	mutex   sync.RWMutex
}

// TestResult represents a network test result for storage
type TestResult struct {
	Timestamp  time.Time     `json:"timestamp"`
	EndpointID string        `json:"endpoint_id"`
	Protocol   string        `json:"protocol"`
	Latency    time.Duration `json:"-"` // Don't marshal directly
	Status     string        `json:"status"`
	Error      string        `json:"error,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for TestResult
func (tr *TestResult) MarshalJSON() ([]byte, error) {
	type Alias TestResult
	return json.Marshal(&struct {
		*Alias
		LatencyInMs float64 `json:"latencyInMs"`
	}{
		Alias:       (*Alias)(tr),
		LatencyInMs: float64(tr.Latency.Nanoseconds()) / 1_000_000.0,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for TestResult
func (tr *TestResult) UnmarshalJSON(data []byte) error {
	type Alias TestResult
	aux := &struct {
		*Alias
		LatencyInMs float64 `json:"latencyInMs"`
	}{
		Alias: (*Alias)(tr),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	tr.Latency = time.Duration(aux.LatencyInMs * 1_000_000.0)
	return nil
}

// DailyDataFile represents a daily data file structure
type DailyDataFile struct {
	Date     string        `json:"date"`
	Results  []*TestResult `json:"results"`
	Metadata *FileMetadata `json:"metadata"`
}

// FileMetadata contains metadata about the data file
type FileMetadata struct {
	Version      string    `json:"version"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
	ResultCount  int       `json:"resultCount"`
}

// NewManager creates a new storage manager
func NewManager(ctx context.Context, dataDir string) (*Manager, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Manager{
		dataDir: dataDir,
		ctx:     ctx,
	}, nil
}

// StoreTestResult stores a test result
func (m *Manager) StoreTestResult(result *TestResult) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	date := result.Timestamp.Format("2006-01-02")
	filename := fmt.Sprintf("%s.json", date)
	filepath := filepath.Join(m.dataDir, filename)

	// Load existing data file or create new one
	dataFile, err := m.loadDailyFile(filepath, date)
	if err != nil {
		return fmt.Errorf("failed to load daily file: %w", err)
	}

	// Add result to file
	dataFile.Results = append(dataFile.Results, result)
	dataFile.Metadata.LastModified = time.Now()
	dataFile.Metadata.ResultCount = len(dataFile.Results)

	// Save updated file
	if err := m.saveDailyFile(filepath, dataFile); err != nil {
		return fmt.Errorf("failed to save daily file: %w", err)
	}

	log.Ctx(m.ctx).Debug().
		Str("endpoint_id", result.EndpointID).
		Str("date", date).
		Int("total_results", dataFile.Metadata.ResultCount).
		Msg("Test result stored")

	return nil
}

// GetResults retrieves test results for a specific date
func (m *Manager) GetResults(date time.Time) ([]*TestResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	dateStr := date.Format("2006-01-02")
	filename := fmt.Sprintf("%s.json", dateStr)
	filepath := filepath.Join(m.dataDir, filename)

	dataFile, err := m.loadDailyFile(filepath, dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to load daily file: %w", err)
	}

	return dataFile.Results, nil
}

// GetResultsRange retrieves test results for a date range (inclusive)
func (m *Manager) GetResultsRange(start, end time.Time) ([]*TestResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var allResults []*TestResult

	// Iterate through each day in the range
	current := start
	for !current.After(end) {
		dateStr := current.Format("2006-01-02")
		filename := fmt.Sprintf("%s.json", dateStr)
		filepath := filepath.Join(m.dataDir, filename)

		// Try to load the file for this day
		dataFile, err := m.loadDailyFile(filepath, dateStr)
		if err != nil {
			// Skip days that don't have data files
			log.Ctx(m.ctx).Debug().
				Str("date", dateStr).
				Err(err).
				Msg("Skipping date with no data")
			current = current.AddDate(0, 0, 1)
			continue
		}

		// Append results from this day
		allResults = append(allResults, dataFile.Results...)

		// Move to next day
		current = current.AddDate(0, 0, 1)
	}

	log.Ctx(m.ctx).Debug().
		Str("start", start.Format("2006-01-02")).
		Str("end", end.Format("2006-01-02")).
		Int("total_results", len(allResults)).
		Msg("Retrieved results for date range")

	return allResults, nil
}

// CleanupOldFiles removes data files older than the specified retention period
func (m *Manager) CleanupOldFiles(retentionDays int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Parse date from filename (YYYY-MM-DD.json)
		filename := entry.Name()
		dateStr := filename[:len(filename)-5] // Remove .json extension

		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Ctx(m.ctx).Warn().
				Str("filename", filename).
				Err(err).
				Msg("Failed to parse date from filename, skipping")
			continue
		}

		// Delete if older than cutoff
		if fileDate.Before(cutoffDate) {
			filepath := filepath.Join(m.dataDir, filename)
			if err := os.Remove(filepath); err != nil {
				log.Ctx(m.ctx).Error().
					Str("filename", filename).
					Err(err).
					Msg("Failed to delete old file")
				continue
			}

			deletedCount++
			log.Ctx(m.ctx).Info().
				Str("filename", filename).
				Str("file_date", dateStr).
				Msg("Deleted old data file")
		}
	}

	log.Ctx(m.ctx).Info().
		Int("deleted_count", deletedCount).
		Int("retention_days", retentionDays).
		Msg("Cleanup completed")

	return nil
}

// SaveConfiguration saves application configuration to disk
func (m *Manager) SaveConfiguration(config interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	configPath := filepath.Join(m.dataDir, "..", "config.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Use atomic write pattern
	tempFile := configPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	if err := os.Rename(tempFile, configPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp config file: %w", err)
	}

	log.Ctx(m.ctx).Info().
		Str("path", configPath).
		Msg("Configuration saved")

	return nil
}

// LoadConfiguration loads application configuration from disk
func (m *Manager) LoadConfiguration(config interface{}) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	configPath := filepath.Join(m.dataDir, "..", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	log.Ctx(m.ctx).Info().
		Str("path", configPath).
		Msg("Configuration loaded")

	return nil
}

// ValidateDataFile validates the structure and integrity of a data file
func (m *Manager) ValidateDataFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var dataFile DailyDataFile
	if err := json.Unmarshal(data, &dataFile); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	// Validate metadata
	if dataFile.Metadata == nil {
		return fmt.Errorf("missing metadata")
	}

	if dataFile.Metadata.Version == "" {
		return fmt.Errorf("missing version in metadata")
	}

	// Validate result count matches
	if dataFile.Metadata.ResultCount != len(dataFile.Results) {
		return fmt.Errorf("metadata result count (%d) does not match actual count (%d)",
			dataFile.Metadata.ResultCount, len(dataFile.Results))
	}

	return nil
}

// RecoverDataFile attempts to recover a corrupted data file
func (m *Manager) RecoverDataFile(filepath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Try to read the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read corrupted file: %w", err)
	}

	// Try to parse and recover what we can
	var dataFile DailyDataFile
	if err := json.Unmarshal(data, &dataFile); err != nil {
		// If JSON is completely corrupted, create backup and new file
		backupPath := filepath + ".corrupted"
		if err := os.Rename(filepath, backupPath); err != nil {
			return fmt.Errorf("failed to backup corrupted file: %w", err)
		}

		log.Ctx(m.ctx).Warn().
			Str("original", filepath).
			Str("backup", backupPath).
			Msg("Corrupted file backed up")

		return fmt.Errorf("file is completely corrupted, backup created at %s", backupPath)
	}

	// If we got here, JSON is valid but might have issues
	// Fix metadata if needed
	if dataFile.Metadata == nil {
		dataFile.Metadata = &FileMetadata{
			Version:      "1.0.0",
			CreatedAt:    time.Now(),
			LastModified: time.Now(),
			ResultCount:  len(dataFile.Results),
		}
	} else {
		// Fix result count if mismatched
		dataFile.Metadata.ResultCount = len(dataFile.Results)
		dataFile.Metadata.LastModified = time.Now()
	}

	// Save recovered file
	if err := m.saveDailyFile(filepath, &dataFile); err != nil {
		return fmt.Errorf("failed to save recovered file: %w", err)
	}

	log.Ctx(m.ctx).Info().
		Str("filepath", filepath).
		Int("recovered_results", len(dataFile.Results)).
		Msg("Data file recovered successfully")

	return nil
}

// Close gracefully closes the storage manager
func (m *Manager) Close() error {
	log.Ctx(m.ctx).Info().Msg("Storage manager closing")
	return nil
}

// loadDailyFile loads a daily data file
func (m *Manager) loadDailyFile(filepath, date string) (*DailyDataFile, error) {
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// Create new file
		return &DailyDataFile{
			Date:    date,
			Results: []*TestResult{},
			Metadata: &FileMetadata{
				Version:      "1.0.0",
				CreatedAt:    time.Now(),
				LastModified: time.Now(),
				ResultCount:  0,
			},
		}, nil
	}

	// Load existing file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dataFile DailyDataFile
	if err := json.Unmarshal(data, &dataFile); err != nil {
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	return &dataFile, nil
}

// saveDailyFile saves a daily data file atomically
// Uses atomic write pattern: write to temp file, then rename
func (m *Manager) saveDailyFile(filepath string, dataFile *DailyDataFile) error {
	data, err := json.MarshalIndent(dataFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data file: %w", err)
	}

	// Write to temporary file first (atomic write pattern)
	tempFile := filepath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomically rename temp file to final location
	if err := os.Rename(tempFile, filepath); err != nil {
		// Clean up temp file on error
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// GetStorageStats returns storage statistics
func (m *Manager) GetStorageStats() (*StorageStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &StorageStats{
		TotalFiles:     0,
		TotalSizeBytes: 0,
		DataDirectory:  m.dataDir,
	}

	// Count files and calculate size
	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		return stats, nil // Return empty stats if directory doesn't exist
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			stats.TotalFiles++
			
			if info, err := entry.Info(); err == nil {
				stats.TotalSizeBytes += info.Size()
			}
		}
	}

	return stats, nil
}

// StorageStats contains storage statistics
type StorageStats struct {
	TotalFiles     int    `json:"totalFiles"`
	TotalSizeBytes int64  `json:"totalSizeBytes"`
	DataDirectory  string `json:"dataDirectory"`
}