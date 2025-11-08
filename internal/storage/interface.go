package storage

import "time"

// Storage defines the interface for data persistence operations
type Storage interface {
	// Configuration methods
	SaveConfiguration(config interface{}) error
	LoadConfiguration(config interface{}) error

	// Test result methods
	StoreTestResult(result *TestResult) error
	GetResults(date time.Time) ([]*TestResult, error)
	GetResultsRange(start, end time.Time) ([]*TestResult, error)

	// File management methods
	CleanupOldFiles(retentionDays int) error
	GetStorageStats() (*StorageStats, error)

	// Data integrity methods
	ValidateDataFile(filepath string) error
	RecoverDataFile(filepath string) error

	// Lifecycle methods
	Close() error
}
