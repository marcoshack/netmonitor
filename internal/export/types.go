package export

import "time"

// ExportRequest represents a request to export data
type ExportRequest struct {
	Format      string    `json:"format"`      // "csv" or "json"
	StartDate   time.Time `json:"startDate"`   // Start of date range
	EndDate     time.Time `json:"endDate"`     // End of date range
	Endpoints   []string  `json:"endpoints"`   // Empty = all endpoints
	Regions     []string  `json:"regions"`     // Empty = all regions
	Columns     []string  `json:"columns"`     // For CSV: which columns to include
	Compressed  bool      `json:"compressed"`  // ZIP compression
	IncludeRaw  bool      `json:"includeRaw"`  // Include raw test results
	IncludeAgg  bool      `json:"includeAgg"`  // Include aggregated data
}

// ExportJob represents an export job
type ExportJob struct {
	ID          string         `json:"id"`
	Request     ExportRequest  `json:"request"`
	Status      string         `json:"status"`      // "pending", "running", "completed", "failed", "cancelled"
	Progress    float64        `json:"progress"`    // 0.0 to 1.0
	StartTime   time.Time      `json:"startTime"`
	EndTime     *time.Time     `json:"endTime,omitempty"`
	FilePath    string         `json:"filePath,omitempty"`
	FileSize    int64          `json:"fileSize"`
	Error       string         `json:"error,omitempty"`
}

// ExportStatus represents the status of an export job
type ExportStatus struct {
	Job               *ExportJob `json:"job"`
	RecordsProcessed  int        `json:"recordsProcessed"`
	TotalRecords      int        `json:"totalRecords"`
	CurrentPhase      string     `json:"currentPhase"`
	EstimatedTimeLeft string     `json:"estimatedTimeLeft"`
}

// Export job status constants
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

// Export format constants
const (
	FormatCSV  = "csv"
	FormatJSON = "json"
)

// CSV column names
const (
	ColumnTimestamp  = "timestamp"
	ColumnEndpointID = "endpoint_id"
	ColumnRegion     = "region"
	ColumnProtocol   = "protocol"
	ColumnLatency    = "latency_ms"
	ColumnStatus     = "status"
	ColumnError      = "error"
)

// DefaultCSVColumns returns the default columns for CSV export
func DefaultCSVColumns() []string {
	return []string{
		ColumnTimestamp,
		ColumnEndpointID,
		ColumnProtocol,
		ColumnStatus,
		ColumnLatency,
	}
}

// AllCSVColumns returns all available columns for CSV export
func AllCSVColumns() []string {
	return []string{
		ColumnTimestamp,
		ColumnEndpointID,
		ColumnRegion,
		ColumnProtocol,
		ColumnStatus,
		ColumnLatency,
		ColumnError,
	}
}
