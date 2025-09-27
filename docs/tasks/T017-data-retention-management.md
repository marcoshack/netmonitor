# T017: Data Retention Management

## Overview
Implement configurable data retention system that automatically manages historical data cleanup based on user-defined retention policies.

## Context
NetMonitor stores historical data indefinitely by default, but users need control over data retention to manage disk space. The system should automatically clean up old data files based on configurable retention periods.

## Task Description
Create a comprehensive data retention management system that handles automatic cleanup of old data files, configurable retention policies, and efficient storage space management.

## Acceptance Criteria
- [ ] Configurable retention period (7-365 days, default 90 days)
- [ ] Automatic daily cleanup process
- [ ] Safe deletion with backup options
- [ ] Storage space monitoring and reporting
- [ ] Retention policy validation
- [ ] Manual cleanup trigger capability
- [ ] Selective retention (keep aggregated data longer)
- [ ] Cleanup logging and reporting
- [ ] Unit tests for retention logic

## Retention Configuration
```go
type RetentionPolicy struct {
    RawDataDays        int  `json:"rawDataDays"`        // Raw test results
    AggregatedDataDays int  `json:"aggregatedDataDays"` // Hourly/daily summaries
    ConfigBackupDays   int  `json:"configBackupDays"`   // Configuration history
    AutoCleanupEnabled bool `json:"autoCleanupEnabled"`
    CleanupTime        string `json:"cleanupTime"`       // "02:00" for 2 AM
}
```

## Implementation Components
- **Retention Manager**: Core logic for applying retention policies
- **Cleanup Scheduler**: Daily scheduled cleanup operations
- **Storage Monitor**: Track disk usage and file counts
- **Backup Manager**: Create backups before deletion
- **Cleanup Logger**: Record all cleanup operations

## API Methods
```go
func (a *App) UpdateRetentionPolicy(policy RetentionPolicy) error
func (a *App) GetRetentionPolicy() (*RetentionPolicy, error)
func (a *App) TriggerManualCleanup() (*CleanupReport, error)
func (a *App) GetStorageStats() (*StorageStats, error)
func (a *App) GetCleanupHistory() ([]*CleanupOperation, error)
```

## Data Structures
```go
type CleanupReport struct {
    StartTime       time.Time `json:"startTime"`
    EndTime         time.Time `json:"endTime"`
    FilesDeleted    int       `json:"filesDeleted"`
    SpaceFreed      int64     `json:"spaceFreed"`     // bytes
    ErrorCount      int       `json:"errorCount"`
    Errors          []string  `json:"errors"`
}

type StorageStats struct {
    TotalFiles      int   `json:"totalFiles"`
    TotalSizeBytes  int64 `json:"totalSizeBytes"`
    OldestDataDate  time.Time `json:"oldestDataDate"`
    NewestDataDate  time.Time `json:"newestDataDate"`
    DaysOfData      int   `json:"daysOfData"`
}
```

## Cleanup Strategy
1. **Raw Data**: Delete files older than `rawDataDays`
2. **Aggregated Data**: Keep aggregated summaries longer than raw data
3. **Configuration**: Keep configuration backup history
4. **Safety**: Never delete current day's data
5. **Verification**: Validate files before deletion

## Verification Steps
1. Set retention to 30 days - should delete files older than 30 days
2. Test with different retention periods - should respect settings
3. Verify aggregated data retention - should keep longer than raw data
4. Test manual cleanup trigger - should clean up immediately
5. Verify storage stats accuracy - should report correct file counts and sizes
6. Test cleanup with active monitoring - should not interfere with data collection
7. Verify cleanup logging - should record all operations
8. Test edge cases - should handle missing files gracefully

## Dependencies
- T016: JSON Storage System
- T003: Configuration System

## Notes
- Run cleanup during low-activity periods (early morning)
- Implement safe deletion (move to trash, then permanent delete)
- Consider compression instead of deletion for very old data
- Provide storage usage warnings before disk space issues
- Plan for different retention policies per data type