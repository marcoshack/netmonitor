# T022: Data Migration System

## Overview
Implement data migration system to handle schema changes, version upgrades, and data format transitions while preserving historical data integrity.

## Context
As NetMonitor evolves, the data storage format may need updates for new features or optimizations. A migration system ensures existing user data remains accessible and is converted to new formats seamlessly.

## Task Description
Create a robust data migration framework that can handle schema version changes, data format conversions, and backward compatibility while maintaining data integrity throughout the upgrade process.

## Acceptance Criteria
- [ ] Schema versioning for data files
- [ ] Automatic migration detection on startup
- [ ] Backward compatibility for older data formats
- [ ] Migration validation and rollback capabilities
- [ ] Progress reporting for large migrations
- [ ] Data integrity verification pre/post migration
- [ ] Backup creation before migrations
- [ ] Migration logging and error reporting
- [ ] Support for incremental migrations

## Schema Versioning
```go
type SchemaVersion struct {
    Version     string    `json:"version"`     // "1.0.0"
    CreatedAt   time.Time `json:"createdAt"`
    Description string    `json:"description"`
    Changes     []string  `json:"changes"`
}

type DataFileHeader struct {
    FormatVersion string        `json:"formatVersion"`
    SchemaVersion string        `json:"schemaVersion"`
    CreatedWith   string        `json:"createdWith"`    // Application version
    Migrations    []Migration   `json:"migrations"`
}
```

## Migration Framework
```go
type Migration interface {
    Version() string
    Description() string
    Migrate(data interface{}) (interface{}, error)
    Validate(data interface{}) error
    Rollback(data interface{}) (interface{}, error)
}

type MigrationManager struct {
    migrations     []Migration
    currentVersion string
    backupEnabled  bool
}
```

## API Methods
```go
func (a *App) CheckMigrationsNeeded() (*MigrationStatus, error)
func (a *App) StartMigration() (*MigrationJob, error)
func (a *App) GetMigrationStatus() (*MigrationProgress, error)
func (a *App) RollbackMigration(migrationID string) error
func (a *App) ValidateDataIntegrity() (*ValidationReport, error)
```

## Migration Types
- **Schema Updates**: Changes to data structure
- **Format Changes**: JSON to binary, compression changes
- **Field Additions**: New fields with default values
- **Field Removal**: Deprecated field cleanup
- **Data Type Changes**: String to numeric conversions
- **Index Rebuilding**: Performance optimization migrations

## Data Structures
```go
type MigrationJob struct {
    ID              string    `json:"id"`
    FromVersion     string    `json:"fromVersion"`
    ToVersion       string    `json:"toVersion"`
    Status          string    `json:"status"`         // "pending", "running", "completed", "failed"
    StartTime       time.Time `json:"startTime"`
    EndTime         *time.Time `json:"endTime"`
    FilesProcessed  int       `json:"filesProcessed"`
    TotalFiles      int       `json:"totalFiles"`
    BackupPath      string    `json:"backupPath"`
    Error           string    `json:"error"`
}

type ValidationReport struct {
    Valid           bool      `json:"valid"`
    FilesChecked    int       `json:"filesChecked"`
    ErrorsFound     int       `json:"errorsFound"`
    WarningsFound   int       `json:"warningsFound"`
    Issues          []Issue   `json:"issues"`
}
```

## Migration Process
1. **Pre-Migration**: Backup existing data
2. **Validation**: Verify data integrity before migration
3. **Migration**: Apply transformations to data files
4. **Verification**: Validate migrated data
5. **Cleanup**: Remove temporary files and update version markers

## Common Migration Scenarios
- **V1.0 to V1.1**: Add endpoint metadata fields
- **V1.1 to V1.2**: Change timestamp format
- **V1.2 to V2.0**: Restructure configuration format
- **V2.0 to V2.1**: Add aggregation data

## Verification Steps
1. Detect migration needed - should identify version differences
2. Run migration with backup - should create backup before migration
3. Validate migrated data - should verify data integrity
4. Test rollback capability - should restore from backup
5. Test incremental migration - should handle partial migrations
6. Verify version tracking - should update version markers correctly
7. Test large dataset migration - should handle with progress reporting
8. Test migration failure recovery - should cleanup and restore

## Dependencies
- T016: JSON Storage System
- T019: Data Backup and Recovery
- T003: Configuration System

## Notes
- Always create backups before migrations
- Test migrations thoroughly with sample data
- Provide clear user communication during migrations
- Plan for future schema changes from the start
- Consider using database-style migration patterns
- Implement proper error handling and recovery mechanisms