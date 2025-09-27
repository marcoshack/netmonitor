# T018: Data Export Functionality

## Overview
Implement data export capabilities for CSV and JSON formats, allowing users to export test results for custom date ranges and analysis.

## Context
Users need to export network monitoring data for external analysis, reporting, or backup purposes. The system should support both CSV and JSON formats with flexible date range selection.

## Task Description
Create a comprehensive data export system that can generate CSV and JSON exports for specified date ranges, endpoints, and regions with progress reporting for large exports.

## Acceptance Criteria
- [ ] CSV export with customizable columns
- [ ] JSON export with full test result data
- [ ] Date range selection for exports
- [ ] Endpoint and region filtering options
- [ ] Progress reporting for large exports
- [ ] Export job queuing and background processing
- [ ] Compressed export files for large datasets
- [ ] Export history and download management
- [ ] Frontend integration for export requests

## Export Configuration
```go
type ExportRequest struct {
    Format      string    `json:"format"`      // "csv" or "json"
    StartDate   time.Time `json:"startDate"`
    EndDate     time.Time `json:"endDate"`
    Endpoints   []string  `json:"endpoints"`   // Empty = all endpoints
    Regions     []string  `json:"regions"`     // Empty = all regions
    Columns     []string  `json:"columns"`     // For CSV: which columns to include
    Compressed  bool      `json:"compressed"`  // ZIP compression
    IncludeRaw  bool      `json:"includeRaw"`  // Include raw test results
    IncludeAgg  bool      `json:"includeAgg"`  // Include aggregated data
}
```

## API Methods
```go
func (a *App) StartExport(request ExportRequest) (*ExportJob, error)
func (a *App) GetExportStatus(jobID string) (*ExportStatus, error)
func (a *App) CancelExport(jobID string) error
func (a *App) DownloadExport(jobID string) ([]byte, error)
func (a *App) GetExportHistory() ([]*ExportJob, error)
func (a *App) DeleteExport(jobID string) error
```

## Data Structures
```go
type ExportJob struct {
    ID          string        `json:"id"`
    Request     ExportRequest `json:"request"`
    Status      string        `json:"status"`      // "pending", "running", "completed", "failed"
    Progress    float64       `json:"progress"`    // 0.0 to 1.0
    StartTime   time.Time     `json:"startTime"`
    EndTime     *time.Time    `json:"endTime"`
    FilePath    string        `json:"filePath"`
    FileSize    int64         `json:"fileSize"`
    Error       string        `json:"error"`
}

type ExportStatus struct {
    Job             *ExportJob `json:"job"`
    RecordsProcessed int       `json:"recordsProcessed"`
    TotalRecords    int       `json:"totalRecords"`
    CurrentPhase    string    `json:"currentPhase"`
    EstimatedTimeLeft string  `json:"estimatedTimeLeft"`
}
```

## CSV Format Options
- **Basic**: Timestamp, Endpoint, Status, Latency
- **Detailed**: All test result fields
- **Summary**: Aggregated statistics only
- **Custom**: User-selected columns

## Export Features
- **Streaming**: Handle large datasets without memory issues
- **Compression**: ZIP files for large exports
- **Validation**: Verify data integrity before export
- **Resumption**: Resume interrupted exports
- **Cleanup**: Automatic cleanup of old export files

## Verification Steps
1. Export small date range to CSV - should complete quickly
2. Export large date range to JSON - should handle with progress reporting
3. Export with endpoint filtering - should only include specified endpoints
4. Export with region filtering - should only include endpoints from specified regions
5. Test compressed export - should create ZIP file
6. Cancel running export - should stop and cleanup
7. Download completed export - should return correct file
8. Test export history - should track all export operations

## Dependencies
- T016: JSON Storage System
- T005: Wails Frontend-Backend Integration
- T013: Test Result Aggregation

## Notes
- Use streaming writers for memory efficiency
- Implement proper escaping for CSV format
- Consider using goroutines for parallel processing
- Provide meaningful progress updates
- Implement proper error handling and cleanup
- Plan for future export formats (Excel, PDF)