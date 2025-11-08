# Data Retention Management

The retention package provides comprehensive data lifecycle management for NetMonitor, including configurable retention policies, automatic cleanup, and storage monitoring.

## Features

- **Configurable Retention Policies**: Set different retention periods for raw data, aggregated data, and configuration backups
- **Automatic Daily Cleanup**: Schedule automated cleanup operations at specified times
- **Manual Cleanup Triggers**: Perform immediate cleanup operations on demand
- **Storage Monitoring**: Track disk usage, file counts, and data age
- **Cleanup History**: Maintain a log of all cleanup operations
- **Safe Deletion**: Protects current day's data from accidental deletion
- **Validation**: Comprehensive policy validation with sensible limits

## Usage

### Basic Setup

```go
import (
    "context"
    "github.com/marcoshack/netmonitor/internal/retention"
)

// Create a retention manager with default policy
ctx := context.Background()
dataDir := "./data"
policy := retention.DefaultRetentionPolicy()

manager, err := retention.NewManager(ctx, dataDir, policy)
if err != nil {
    log.Fatal(err)
}
defer manager.Close()
```

### Custom Retention Policy

```go
// Create a custom retention policy
policy := &retention.RetentionPolicy{
    RawDataDays:        30,   // Keep raw test results for 30 days
    AggregatedDataDays: 180,  // Keep aggregated summaries for 6 months
    ConfigBackupDays:   14,   // Keep config backups for 2 weeks
    AutoCleanupEnabled: true, // Enable automatic daily cleanup
    CleanupTime:        "02:00", // Run cleanup at 2 AM
}

// Validate the policy
if err := policy.Validate(); err != nil {
    log.Fatal(err)
}

manager, err := retention.NewManager(ctx, dataDir, policy)
if err != nil {
    log.Fatal(err)
}
```

### Manual Cleanup

```go
// Trigger an immediate cleanup operation
report, err := manager.TriggerManualCleanup()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Deleted %d files, freed %d bytes\n",
    report.FilesDeleted, report.SpaceFreed)
fmt.Printf("Duration: %v\n", report.EndTime.Sub(report.StartTime))

if report.ErrorCount > 0 {
    for _, errMsg := range report.Errors {
        fmt.Println("Error:", errMsg)
    }
}
```

### Storage Statistics

```go
// Get current storage statistics
stats, err := manager.GetStorageStats()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total files: %d\n", stats.TotalFiles)
fmt.Printf("Total size: %d bytes (%.2f MB)\n",
    stats.TotalSizeBytes, float64(stats.TotalSizeBytes)/1024/1024)
fmt.Printf("Data range: %s to %s\n",
    stats.OldestDataDate.Format("2006-01-02"),
    stats.NewestDataDate.Format("2006-01-02"))
fmt.Printf("Days of data: %d\n", stats.DaysOfData)
```

### Update Retention Policy

```go
// Update the retention policy at runtime
newPolicy := &retention.RetentionPolicy{
    RawDataDays:        60,
    AggregatedDataDays: 365,
    ConfigBackupDays:   30,
    AutoCleanupEnabled: true,
    CleanupTime:        "03:00",
}

if err := manager.UpdatePolicy(newPolicy); err != nil {
    log.Fatal(err)
}
```

### Cleanup History

```go
// Retrieve cleanup operation history
history := manager.GetCleanupHistory()

for _, op := range history {
    fmt.Printf("Cleanup at %s:\n", op.Timestamp.Format("2006-01-02 15:04:05"))
    fmt.Printf("  Files deleted: %d\n", op.FilesDeleted)
    fmt.Printf("  Space freed: %d bytes\n", op.SpaceFreed)
    fmt.Printf("  Duration: %d ms\n", op.Duration)
    fmt.Printf("  Success: %v\n", op.Success)
    if op.ErrorMessage != "" {
        fmt.Printf("  Error: %s\n", op.ErrorMessage)
    }
}
```

## API Integration

The retention system is integrated into the main App with the following API methods:

```go
// Update retention policy
err := app.UpdateRetentionPolicy(policy)

// Get current policy
policy, err := app.GetRetentionPolicy()

// Trigger manual cleanup
report, err := app.TriggerManualCleanup()

// Get storage statistics
stats, err := app.GetStorageStats()

// Get cleanup history
history, err := app.GetCleanupHistory()
```

## Retention Policy Validation

The retention policy has the following constraints:

- **RawDataDays**: 7-365 days
- **AggregatedDataDays**: 7-730 days (must be ≥ RawDataDays)
- **ConfigBackupDays**: 1-365 days
- **CleanupTime**: Valid time format "HH:MM" (24-hour)

## Cleanup Strategy

1. **Raw Data Files**: Deleted when older than `RawDataDays`
2. **Current Day Protection**: Today's data is never deleted
3. **Safe Cutoff**: If retention period would delete yesterday's data, cutoff is adjusted to protect it
4. **File Validation**: Only valid date-formatted JSON files are considered for cleanup
5. **Error Handling**: Failed deletions are logged and reported but don't stop the cleanup process

## Automatic Cleanup Scheduler

When `AutoCleanupEnabled` is true, the manager automatically:

1. Calculates the next cleanup time based on `CleanupTime`
2. Schedules cleanup for the configured time each day
3. Executes cleanup and logs results
4. Records each operation in the cleanup history
5. Reschedules for the next day

The scheduler runs in a background goroutine and is automatically stopped when the manager is closed.

## Storage Statistics

The `StorageStats` structure provides:

- **TotalFiles**: Count of data files
- **TotalSizeBytes**: Total disk space used
- **OldestDataDate**: Date of oldest data file
- **NewestDataDate**: Date of newest data file
- **DaysOfData**: Number of days spanned by data

## Cleanup Reports

Each cleanup operation generates a `CleanupReport` with:

- **StartTime**: When cleanup began
- **EndTime**: When cleanup completed
- **FilesDeleted**: Number of files removed
- **SpaceFreed**: Bytes of disk space freed
- **ErrorCount**: Number of errors encountered
- **Errors**: List of error messages

## Cleanup History

The manager maintains up to 100 recent cleanup operations in memory and persists them to disk. History is automatically:

- Loaded on manager startup
- Updated after each cleanup
- Saved to `cleanup_history.json`
- Trimmed to maximum length

## Demo

Run the retention demo to see the system in action:

```bash
go run ./cmd/retention-demo/main.go
```

This demonstrates:
- Creating test data files
- Getting storage statistics
- Triggering cleanup operations
- Updating retention policies
- Viewing cleanup history

## Testing

The package includes comprehensive unit tests:

```bash
# Run all retention tests
go test ./internal/retention/... -v

# Run with coverage
go test ./internal/retention/... -cover

# Run specific test
go test ./internal/retention/... -run TestTriggerManualCleanup
```

## Thread Safety

All operations are thread-safe using read-write mutexes. Multiple goroutines can safely:
- Read policies and statistics concurrently
- Update policies (exclusive access)
- Trigger cleanups (exclusive access)
- Access cleanup history (concurrent reads)

## Performance Considerations

- Cleanup operations acquire an exclusive lock on the data directory
- Large directories may take longer to scan
- Consider scheduling cleanup during low-activity periods
- File operations are batched for efficiency
- Statistics are calculated on-demand (not cached)

## Error Handling

- Invalid policies are rejected at creation and update time
- File operation errors are logged but don't stop cleanup
- All errors are included in cleanup reports
- Missing directories are handled gracefully
- Corrupted history files are ignored and recreated

## Best Practices

1. **Retention Periods**: Set aggregated data retention longer than raw data
2. **Cleanup Time**: Schedule during off-peak hours (2-4 AM recommended)
3. **Testing**: Test retention policies in non-production environments first
4. **Monitoring**: Review cleanup history regularly for errors
5. **Validation**: Always validate policies before applying
6. **Backups**: Consider external backups for critical data
7. **Gradual Changes**: Reduce retention periods gradually, not dramatically

## Integration with Storage Manager

The retention manager works alongside the storage manager:

- **Storage Manager**: Handles day-to-day data operations
- **Retention Manager**: Handles lifecycle and cleanup
- Both operate on the same data directory
- No conflicts as operations are properly synchronized

## Future Enhancements

Potential future improvements:
- Compression instead of deletion for very old data
- Selective retention per endpoint or region
- Archive to cloud storage before deletion
- Storage quota limits with automatic cleanup
- Email notifications for cleanup operations
- Metrics export for monitoring systems
