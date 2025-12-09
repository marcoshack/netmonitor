package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/marcoshack/netmonitor/internal/models"
)

type Storage struct {
	DataDir string
	mu      sync.Mutex
}

func NewStorage(dataDir string) *Storage {
	_ = os.MkdirAll(dataDir, 0755)
	return &Storage{
		DataDir: dataDir,
	}
}

// GetDailyFilePath returns the file path for a specific day
func (s *Storage) GetDailyFilePath(date time.Time) string {
	filename := fmt.Sprintf("%s.json", date.Format("2006-01-02"))
	return filepath.Join(s.DataDir, filename)
}

// SaveResult appends a test result to the daily log file
func (s *Storage) SaveResult(result models.TestResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filepath := s.GetDailyFilePath(result.Timestamp)

	// In a real app with high volume, we'd want to buffer or use a more efficient storage.
	// For this app, reading/appending/writing or appending to a file stream is fine.
	// JSON lines or a JSON array?
	// The requirement says "Historical Data Structure: data/YYYY-MM-DD.json ... Data Points: Each test result includes..."
	// Appending to a JSON array in a file requires reading the whole file, decoding, appending, encoding.
	// Appending JSON lines is better for performance, but JSON array is standard for "Valid JSON file".
	// Let's go with Reading/Writing Array for correctness with the spec "stored in daily JSON files".
	// To minimize I/O issues, we'll try to append. But standard JSON needs array wrapper [ ... ].
	// Let's implementation: Read existing, Append, Write. Max file size won't be huge (3 months limit elsewhere, but daily file size depends on interval).
	// Interval 5 mins * 12 endpoints * 24 hours * 12 checks/hour = 3456 entries. Tiny.

	var results []models.TestResult

	// Read existing
	if _, err := os.Stat(filepath); err == nil {
		data, err := os.ReadFile(filepath)
		if err == nil {
			_ = json.Unmarshal(data, &results)
		}
	}

	results = append(results, result)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// GetResultsForDay retrieves all results for a specific day
func (s *Storage) GetResultsForDay(date time.Time) ([]models.TestResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filepath := s.GetDailyFilePath(date)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return []models.TestResult{}, nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var results []models.TestResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetResultsForRange retrieves results between start and end time
func (s *Storage) GetResultsForRange(start, end time.Time) ([]models.TestResult, error) {
	// Identify all days in range
	var allResults []models.TestResult

	current := start
	// Normalize to start of day
	current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())

	for !current.After(end) {
		dayResults, _ := s.GetResultsForDay(current)
		for _, r := range dayResults {
			if (r.Timestamp.Equal(start) || r.Timestamp.After(start)) && (r.Timestamp.Equal(end) || r.Timestamp.Before(end)) {
				allResults = append(allResults, r)
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return allResults, nil
}
