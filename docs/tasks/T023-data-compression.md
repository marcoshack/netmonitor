# T023: Data Compression

## Overview
Implement data compression for historical data files to reduce storage space while maintaining fast access to recent data and reasonable decompression performance.

## Context
As NetMonitor accumulates historical data over time, storage space becomes a concern. Implementing compression for older data files can significantly reduce disk usage while keeping recent data uncompressed for optimal performance.

## Task Description
Create a data compression system that automatically compresses older data files based on age policies while providing transparent access to compressed data for queries and exports.

## Acceptance Criteria
- [ ] Automatic compression of data files based on age
- [ ] Configurable compression policies (age thresholds, algorithms)
- [ ] Transparent decompression for data access
- [ ] Multiple compression algorithm support (gzip, zstd, lz4)
- [ ] Compression performance monitoring
- [ ] Space savings reporting
- [ ] Background compression processing
- [ ] Compression integrity verification
- [ ] Integration with existing storage operations

## Compression Configuration
```go
type CompressionConfig struct {
    Enabled               bool     `json:"enabled"`
    Algorithm             string   `json:"algorithm"`            // "gzip", "zstd", "lz4"
    CompressionLevel      int      `json:"compressionLevel"`     // Algorithm-specific level
    CompressAfterDays     int      `json:"compressAfterDays"`    // Compress files older than X days
    KeepUncompressedDays  int      `json:"keepUncompressedDays"` // Keep recent files uncompressed
    BackgroundCompression bool     `json:"backgroundCompression"`
    CompressionSchedule   string   `json:"compressionSchedule"`  // Cron-style schedule
}
```

## Implementation Components
```go
type CompressionManager struct {
    config     CompressionConfig
    compressor Compressor
    scheduler  *CompressionScheduler
    metrics    *CompressionMetrics
}

type Compressor interface {
    Compress(data []byte) ([]byte, error)
    Decompress(data []byte) ([]byte, error)
    Algorithm() string
    Level() int
}

type CompressionJob struct {
    ID           string    `json:"id"`
    FilePath     string    `json:"filePath"`
    OriginalSize int64     `json:"originalSize"`
    CompressedSize int64   `json:"compressedSize"`
    Algorithm    string    `json:"algorithm"`
    StartTime    time.Time `json:"startTime"`
    EndTime      *time.Time `json:"endTime"`
    Status       string    `json:"status"`       // "pending", "running", "completed", "failed"
    Error        string    `json:"error"`
}
```

## API Methods
```go
func (a *App) GetCompressionConfig() (*CompressionConfig, error)
func (a *App) UpdateCompressionConfig(config CompressionConfig) error
func (a *App) TriggerCompression() error
func (a *App) GetCompressionMetrics() (*CompressionMetrics, error)
func (a *App) GetCompressionJobs() ([]*CompressionJob, error)
func (a *App) DecompressFile(filePath string) error
```

## Compression Strategies
- **Age-Based**: Compress files older than configured threshold
- **Size-Based**: Compress files above certain size
- **Access-Based**: Compress files not accessed recently
- **Scheduled**: Compress during low-activity periods
- **Manual**: On-demand compression triggers

## Storage Integration
- **File Naming**: Add `.gz`, `.zst`, or `.lz4` extensions
- **Metadata**: Track compression status in file headers
- **Transparent Access**: Automatically decompress when reading
- **Query Optimization**: Consider compression status in query planning

## Compression Metrics
```go
type CompressionMetrics struct {
    TotalFiles          int     `json:"totalFiles"`
    CompressedFiles     int     `json:"compressedFiles"`
    UncompressedFiles   int     `json:"uncompressedFiles"`
    TotalSizeBytes      int64   `json:"totalSizeBytes"`
    CompressedSizeBytes int64   `json:"compressedSizeBytes"`
    SpaceSavedBytes     int64   `json:"spaceSavedBytes"`
    CompressionRatio    float64 `json:"compressionRatio"`
    LastCompressionRun  time.Time `json:"lastCompressionRun"`
}
```

## Verification Steps
1. Configure compression for files older than 7 days - should compress qualifying files
2. Test different compression algorithms - should use configured algorithm
3. Access compressed data - should transparently decompress
4. Verify compression ratios - should achieve expected space savings
5. Test background compression - should run during scheduled times
6. Verify data integrity - should validate compressed files
7. Test query performance - should handle mixed compressed/uncompressed data
8. Test compression with exports - should handle compressed source files

## Dependencies
- T016: JSON Storage System
- T021: Historical Data Queries
- T017: Data Retention Management

## Notes
- Choose compression algorithms based on read/write patterns
- gzip: Good general purpose, widely supported
- zstd: Better compression ratio and speed
- lz4: Fast compression/decompression, lower ratio
- Consider compression impact on query performance
- Monitor CPU usage during compression operations
- Plan for future decompression on different systems