package retention

import (
	"fmt"
	"time"
)

// RetentionPolicy defines data retention configuration
type RetentionPolicy struct {
	RawDataDays        int    `json:"rawDataDays"`        // Raw test results retention
	AggregatedDataDays int    `json:"aggregatedDataDays"` // Hourly/daily summaries retention
	ConfigBackupDays   int    `json:"configBackupDays"`   // Configuration history retention
	AutoCleanupEnabled bool   `json:"autoCleanupEnabled"` // Enable automatic daily cleanup
	CleanupTime        string `json:"cleanupTime"`        // Time of day for cleanup (HH:MM)
}

// CleanupReport contains the results of a cleanup operation
type CleanupReport struct {
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	FilesDeleted int       `json:"filesDeleted"`
	SpaceFreed   int64     `json:"spaceFreed"` // bytes
	ErrorCount   int       `json:"errorCount"`
	Errors       []string  `json:"errors"`
}

// StorageStats contains storage statistics
type StorageStats struct {
	TotalFiles     int       `json:"totalFiles"`
	TotalSizeBytes int64     `json:"totalSizeBytes"`
	OldestDataDate time.Time `json:"oldestDataDate"`
	NewestDataDate time.Time `json:"newestDataDate"`
	DaysOfData     int       `json:"daysOfData"`
}

// CleanupOperation represents a single cleanup operation log entry
type CleanupOperation struct {
	Timestamp    time.Time `json:"timestamp"`
	FilesDeleted int       `json:"filesDeleted"`
	SpaceFreed   int64     `json:"spaceFreed"`
	Duration     int64     `json:"duration"` // milliseconds
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// DefaultRetentionPolicy returns the default retention policy
func DefaultRetentionPolicy() *RetentionPolicy {
	return &RetentionPolicy{
		RawDataDays:        90,  // 90 days default
		AggregatedDataDays: 365, // Keep aggregated data for 1 year
		ConfigBackupDays:   30,  // Keep config backups for 30 days
		AutoCleanupEnabled: true,
		CleanupTime:        "02:00", // 2 AM
	}
}

// Validate validates the retention policy
func (p *RetentionPolicy) Validate() error {
	// Validate raw data retention
	if p.RawDataDays < 7 || p.RawDataDays > 365 {
		return fmt.Errorf("rawDataDays must be between 7 and 365, got %d", p.RawDataDays)
	}

	// Validate aggregated data retention
	if p.AggregatedDataDays < 7 || p.AggregatedDataDays > 730 {
		return fmt.Errorf("aggregatedDataDays must be between 7 and 730, got %d", p.AggregatedDataDays)
	}

	// Validate config backup retention
	if p.ConfigBackupDays < 1 || p.ConfigBackupDays > 365 {
		return fmt.Errorf("configBackupDays must be between 1 and 365, got %d", p.ConfigBackupDays)
	}

	// Validate cleanup time format
	if _, err := time.Parse("15:04", p.CleanupTime); err != nil {
		return fmt.Errorf("invalid cleanupTime format, must be HH:MM (e.g., '02:00'): %w", err)
	}

	// Logical validation: aggregated data should be kept longer than raw data
	if p.AggregatedDataDays < p.RawDataDays {
		return fmt.Errorf("aggregatedDataDays (%d) should be >= rawDataDays (%d)",
			p.AggregatedDataDays, p.RawDataDays)
	}

	return nil
}

// GetCleanupTimeToday returns the next cleanup time for today
func (p *RetentionPolicy) GetCleanupTimeToday() (time.Time, error) {
	now := time.Now()
	cleanupTime, err := time.Parse("15:04", p.CleanupTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cleanup time: %w", err)
	}

	// Combine today's date with the cleanup time
	cleanupDateTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		cleanupTime.Hour(), cleanupTime.Minute(), 0, 0,
		now.Location(),
	)

	return cleanupDateTime, nil
}

// GetNextCleanupTime returns the next scheduled cleanup time
func (p *RetentionPolicy) GetNextCleanupTime() (time.Time, error) {
	now := time.Now()
	todayCleanup, err := p.GetCleanupTimeToday()
	if err != nil {
		return time.Time{}, err
	}

	// If cleanup time for today has passed, schedule for tomorrow
	if now.After(todayCleanup) {
		return todayCleanup.AddDate(0, 0, 1), nil
	}

	return todayCleanup, nil
}
