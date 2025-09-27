# T016: JSON Storage System

## Overview
Implement the core JSON-based storage system for persisting test results, configuration data, and historical information using daily organized files.

## Context
NetMonitor stores all data locally in JSON files organized by date. The storage system needs to efficiently handle daily data files, configuration persistence, and provide fast access for dashboard queries.

## Task Description
Create a robust JSON storage system that manages daily data files, configuration persistence, and provides efficient read/write operations for the monitoring system.

## Acceptance Criteria
- [ ] Daily JSON file organization (`data/YYYY-MM-DD.json`)
- [ ] Configuration file storage and management
- [ ] Efficient append operations for new test results
- [ ] Fast query operations for dashboard data
- [ ] Atomic file operations to prevent corruption
- [ ] File rotation and cleanup based on retention policy
- [ ] Concurrent read/write safety
- [ ] Error handling and data recovery
- [ ] Unit tests for all storage operations

## Storage Structure
```
data/
├── config.json              # Main configuration
├── 2025-09-27.json          # Daily test results
├── 2025-09-26.json
├── ...
└── aggregated/              # Pre-computed aggregations
    ├── 2025-09-27-hourly.json
    └── 2025-09-27-daily.json
```

## Data File Format
```go
type DailyDataFile struct {
    Date         string        `json:"date"`
    Results      []TestResult  `json:"results"`
    Metadata     FileMetadata  `json:"metadata"`
}

type FileMetadata struct {
    Version      string    `json:"version"`
    CreatedAt    time.Time `json:"createdAt"`
    LastModified time.Time `json:"lastModified"`
    ResultCount  int       `json:"resultCount"`
}
```

## Storage Interface
```go
type Storage interface {
    // Configuration
    SaveConfiguration(config *Config) error
    LoadConfiguration() (*Config, error)

    // Test results
    AppendResult(result *TestResult) error
    GetResults(date time.Time) ([]TestResult, error)
    GetResultsRange(start, end time.Time) ([]TestResult, error)

    // File management
    CleanupOldFiles(retentionDays int) error
    GetStorageStats() (*StorageStats, error)
}
```

## Implementation Requirements
- Use atomic file operations (write to temp, then rename)
- Implement file locking for concurrent access
- Handle large files efficiently (streaming reads/writes)
- Validate JSON structure on read operations
- Implement backup and recovery mechanisms

## Verification Steps
1. Save configuration - should create/update config.json
2. Append test result - should add to daily file
3. Query results by date - should return correct data
4. Query date range - should aggregate across multiple files
5. Test concurrent access - should handle multiple writers safely
6. Test large file handling - should efficiently handle thousands of results
7. Test cleanup operation - should remove old files correctly
8. Test data recovery - should handle corrupted files gracefully

## Dependencies
- T002: Basic Application Structure
- T006: Network Test Interfaces

## Notes
- Consider using JSON streaming for large files
- Implement proper file locking mechanism
- Plan for future database migration if needed
- Consider compression for old data files
- Implement data validation and schema versioning