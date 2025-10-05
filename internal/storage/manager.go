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

// saveDailyFile saves a daily data file
func (m *Manager) saveDailyFile(filepath string, dataFile *DailyDataFile) error {
	data, err := json.MarshalIndent(dataFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data file: %w", err)
	}

	return os.WriteFile(filepath, data, 0644)
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