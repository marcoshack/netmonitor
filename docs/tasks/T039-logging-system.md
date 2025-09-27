# T039: Comprehensive Logging System

## Overview
Implement a robust logging system with structured logging, multiple output formats, log rotation, and configurable log levels for debugging and monitoring purposes.

## Context
NetMonitor needs comprehensive logging for debugging issues, monitoring system health, and auditing network monitoring activities. The logging system should be performant, configurable, and provide useful information for troubleshooting.

## Task Description
Create a comprehensive logging framework with structured logging, multiple appenders, log rotation, filtering, and integration with the monitoring system for operational visibility.

## Acceptance Criteria
- [ ] Structured logging with JSON and text formats
- [ ] Multiple log levels (Debug, Info, Warn, Error, Fatal)
- [ ] Log rotation based on size and time
- [ ] Configurable log output destinations (file, console, syslog)
- [ ] Performance monitoring and metrics logging
- [ ] Security-aware logging (no sensitive data)
- [ ] Cross-platform compatibility
- [ ] Integration with system monitoring tools
- [ ] Log analysis and debugging tools

## Logging Architecture
```go
package logging

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/sirupsen/logrus"
    "gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    WithFields(fields ...Field) Logger
    WithContext(ctx context.Context) Logger
}

type Field struct {
    Key   string
    Value interface{}
}

type LogConfig struct {
    Level      string `json:"level"`      // debug, info, warn, error, fatal
    Format     string `json:"format"`     // json, text
    Output     string `json:"output"`     // file, console, both
    FilePath   string `json:"filePath"`
    MaxSize    int    `json:"maxSize"`    // MB
    MaxBackups int    `json:"maxBackups"`
    MaxAge     int    `json:"maxAge"`     // days
    Compress   bool   `json:"compress"`
}

type NetMonitorLogger struct {
    logger *logrus.Logger
    config *LogConfig
    fields logrus.Fields
}

func NewLogger(config *LogConfig) (*NetMonitorLogger, error) {
    logger := logrus.New()

    // Set log level
    level, err := logrus.ParseLevel(config.Level)
    if err != nil {
        return nil, fmt.Errorf("invalid log level: %v", err)
    }
    logger.SetLevel(level)

    // Set formatter
    switch config.Format {
    case "json":
        logger.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: time.RFC3339,
            FieldMap: logrus.FieldMap{
                logrus.FieldKeyTime:  "timestamp",
                logrus.FieldKeyLevel: "level",
                logrus.FieldKeyMsg:   "message",
            },
        })
    case "text":
        logger.SetFormatter(&logrus.TextFormatter{
            TimestampFormat: "2006-01-02 15:04:05",
            FullTimestamp:   true,
        })
    default:
        return nil, fmt.Errorf("unsupported log format: %s", config.Format)
    }

    // Set output
    if err := setupOutput(logger, config); err != nil {
        return nil, fmt.Errorf("failed to setup log output: %v", err)
    }

    return &NetMonitorLogger{
        logger: logger,
        config: config,
        fields: make(logrus.Fields),
    }, nil
}

func setupOutput(logger *logrus.Logger, config *LogConfig) error {
    var writers []io.Writer

    // Console output
    if config.Output == "console" || config.Output == "both" {
        writers = append(writers, os.Stdout)
    }

    // File output
    if config.Output == "file" || config.Output == "both" {
        if config.FilePath == "" {
            return fmt.Errorf("file path is required for file output")
        }

        // Create directory if it doesn't exist
        dir := filepath.Dir(config.FilePath)
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("failed to create log directory: %v", err)
        }

        // Setup log rotation
        rotator := &lumberjack.Logger{
            Filename:   config.FilePath,
            MaxSize:    config.MaxSize,
            MaxBackups: config.MaxBackups,
            MaxAge:     config.MaxAge,
            Compress:   config.Compress,
        }

        writers = append(writers, rotator)
    }

    if len(writers) == 0 {
        return fmt.Errorf("no output destination configured")
    }

    // Set output to multi-writer if multiple destinations
    if len(writers) == 1 {
        logger.SetOutput(writers[0])
    } else {
        logger.SetOutput(io.MultiWriter(writers...))
    }

    return nil
}

func (l *NetMonitorLogger) Debug(msg string, fields ...Field) {
    l.log(logrus.DebugLevel, msg, fields...)
}

func (l *NetMonitorLogger) Info(msg string, fields ...Field) {
    l.log(logrus.InfoLevel, msg, fields...)
}

func (l *NetMonitorLogger) Warn(msg string, fields ...Field) {
    l.log(logrus.WarnLevel, msg, fields...)
}

func (l *NetMonitorLogger) Error(msg string, fields ...Field) {
    l.log(logrus.ErrorLevel, msg, fields...)
}

func (l *NetMonitorLogger) Fatal(msg string, fields ...Field) {
    l.log(logrus.FatalLevel, msg, fields...)
}

func (l *NetMonitorLogger) log(level logrus.Level, msg string, fields ...Field) {
    entry := l.logger.WithFields(l.fields)

    for _, field := range fields {
        entry = entry.WithField(field.Key, field.Value)
    }

    entry.Log(level, msg)
}

func (l *NetMonitorLogger) WithFields(fields ...Field) Logger {
    newFields := make(logrus.Fields)
    for k, v := range l.fields {
        newFields[k] = v
    }

    for _, field := range fields {
        newFields[field.Key] = field.Value
    }

    return &NetMonitorLogger{
        logger: l.logger,
        config: l.config,
        fields: newFields,
    }
}

func (l *NetMonitorLogger) WithContext(ctx context.Context) Logger {
    // Extract context values for logging
    fields := []Field{}

    if reqID := ctx.Value("request_id"); reqID != nil {
        fields = append(fields, Field{Key: "request_id", Value: reqID})
    }

    if userID := ctx.Value("user_id"); userID != nil {
        fields = append(fields, Field{Key: "user_id", Value: userID})
    }

    return l.WithFields(fields...)
}

// Helper functions for creating fields
func String(key, value string) Field {
    return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
    return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
    return Field{Key: key, Value: value.String()}
}

func Error(err error) Field {
    return Field{Key: "error", Value: err.Error()}
}

func Endpoint(id, name string) Field {
    return Field{Key: "endpoint", Value: map[string]string{
        "id":   id,
        "name": name,
    }}
}
```

## Specialized Loggers
```go
// Network monitoring specific loggers
type MonitoringLogger struct {
    logger Logger
}

func NewMonitoringLogger(baseLogger Logger) *MonitoringLogger {
    return &MonitoringLogger{
        logger: baseLogger.WithFields(String("component", "monitoring")),
    }
}

func (ml *MonitoringLogger) TestStarted(endpointID, testType string) {
    ml.logger.Info("Network test started",
        String("endpoint_id", endpointID),
        String("test_type", testType),
        String("event", "test_started"))
}

func (ml *MonitoringLogger) TestCompleted(endpointID string, result *TestResult) {
    fields := []Field{
        String("endpoint_id", endpointID),
        String("status", result.Status),
        Duration("latency", result.Latency),
        String("event", "test_completed"),
    }

    if result.Error != "" {
        fields = append(fields, String("error", result.Error))
    }

    if result.Status == "success" {
        ml.logger.Info("Network test completed successfully", fields...)
    } else {
        ml.logger.Warn("Network test failed", fields...)
    }
}

func (ml *MonitoringLogger) EndpointStatusChanged(endpointID, oldStatus, newStatus string) {
    ml.logger.Info("Endpoint status changed",
        String("endpoint_id", endpointID),
        String("old_status", oldStatus),
        String("new_status", newStatus),
        String("event", "status_changed"))
}

func (ml *MonitoringLogger) ThresholdExceeded(endpointID string, metric string, value, threshold float64) {
    ml.logger.Warn("Threshold exceeded",
        String("endpoint_id", endpointID),
        String("metric", metric),
        Field{Key: "value", Value: value},
        Field{Key: "threshold", Value: threshold},
        String("event", "threshold_exceeded"))
}

// Performance logger
type PerformanceLogger struct {
    logger Logger
}

func NewPerformanceLogger(baseLogger Logger) *PerformanceLogger {
    return &PerformanceLogger{
        logger: baseLogger.WithFields(String("component", "performance")),
    }
}

func (pl *PerformanceLogger) APICall(method string, duration time.Duration, err error) {
    fields := []Field{
        String("method", method),
        Duration("duration", duration),
        String("event", "api_call"),
    }

    if err != nil {
        fields = append(fields, Error(err))
        pl.logger.Error("API call failed", fields...)
    } else {
        pl.logger.Debug("API call completed", fields...)
    }
}

func (pl *PerformanceLogger) DatabaseOperation(operation string, duration time.Duration, recordCount int) {
    pl.logger.Debug("Database operation",
        String("operation", operation),
        Duration("duration", duration),
        Int("record_count", recordCount),
        String("event", "db_operation"))
}

func (pl *PerformanceLogger) MemoryUsage(heapSize, stackSize int64) {
    pl.logger.Debug("Memory usage",
        Field{Key: "heap_size", Value: heapSize},
        Field{Key: "stack_size", Value: stackSize},
        String("event", "memory_usage"))
}
```

## Security-Aware Logging
```go
// Security logger with data sanitization
type SecurityLogger struct {
    logger Logger
}

func NewSecurityLogger(baseLogger Logger) *SecurityLogger {
    return &SecurityLogger{
        logger: baseLogger.WithFields(String("component", "security")),
    }
}

func (sl *SecurityLogger) sanitizeValue(value interface{}) interface{} {
    if str, ok := value.(string); ok {
        // Sanitize sensitive patterns
        patterns := []string{
            "password", "passwd", "secret", "token", "key",
            "authorization", "auth", "credential", "api_key",
        }

        lower := strings.ToLower(str)
        for _, pattern := range patterns {
            if strings.Contains(lower, pattern) {
                return "[REDACTED]"
            }
        }
    }
    return value
}

func (sl *SecurityLogger) sanitizeFields(fields []Field) []Field {
    sanitized := make([]Field, len(fields))
    for i, field := range fields {
        sanitized[i] = Field{
            Key:   field.Key,
            Value: sl.sanitizeValue(field.Value),
        }
    }
    return sanitized
}

func (sl *SecurityLogger) AuthenticationAttempt(username string, success bool, clientIP string) {
    status := "success"
    if !success {
        status = "failure"
    }

    sl.logger.Info("Authentication attempt",
        String("username", username),
        String("status", status),
        String("client_ip", clientIP),
        String("event", "auth_attempt"))
}

func (sl *SecurityLogger) ConfigurationChanged(user, component string, changes map[string]interface{}) {
    // Sanitize configuration changes
    sanitizedChanges := make(map[string]interface{})
    for k, v := range changes {
        sanitizedChanges[k] = sl.sanitizeValue(v)
    }

    sl.logger.Info("Configuration changed",
        String("user", user),
        String("component", component),
        Field{Key: "changes", Value: sanitizedChanges},
        String("event", "config_changed"))
}
```

## Log Analysis Tools
```go
// Log analysis and metrics
type LogAnalyzer struct {
    logger Logger
    metrics map[string]*LogMetric
    mu      sync.RWMutex
}

type LogMetric struct {
    Count     int64
    LastSeen  time.Time
    FirstSeen time.Time
}

func NewLogAnalyzer(logger Logger) *LogAnalyzer {
    return &LogAnalyzer{
        logger:  logger,
        metrics: make(map[string]*LogMetric),
    }
}

func (la *LogAnalyzer) RecordEvent(event string) {
    la.mu.Lock()
    defer la.mu.Unlock()

    if metric, exists := la.metrics[event]; exists {
        metric.Count++
        metric.LastSeen = time.Now()
    } else {
        la.metrics[event] = &LogMetric{
            Count:     1,
            FirstSeen: time.Now(),
            LastSeen:  time.Now(),
        }
    }
}

func (la *LogAnalyzer) GetMetrics() map[string]*LogMetric {
    la.mu.RLock()
    defer la.mu.RUnlock()

    result := make(map[string]*LogMetric)
    for k, v := range la.metrics {
        result[k] = &LogMetric{
            Count:     v.Count,
            FirstSeen: v.FirstSeen,
            LastSeen:  v.LastSeen,
        }
    }
    return result
}

func (la *LogAnalyzer) Reset() {
    la.mu.Lock()
    defer la.mu.Unlock()
    la.metrics = make(map[string]*LogMetric)
}
```

## Application Integration
```go
// Application logger setup
type App struct {
    logger            Logger
    monitoringLogger  *MonitoringLogger
    performanceLogger *PerformanceLogger
    securityLogger    *SecurityLogger
    analyzer          *LogAnalyzer
}

func (a *App) initializeLogging(config *LogConfig) error {
    // Create base logger
    baseLogger, err := NewLogger(config)
    if err != nil {
        return fmt.Errorf("failed to create logger: %v", err)
    }

    a.logger = baseLogger

    // Create specialized loggers
    a.monitoringLogger = NewMonitoringLogger(baseLogger)
    a.performanceLogger = NewPerformanceLogger(baseLogger)
    a.securityLogger = NewSecurityLogger(baseLogger)
    a.analyzer = NewLogAnalyzer(baseLogger)

    // Log application startup
    a.logger.Info("NetMonitor application starting",
        String("version", a.GetVersion()),
        String("platform", runtime.GOOS),
        String("architecture", runtime.GOARCH))

    return nil
}

// Example usage in monitoring
func (a *App) RunNetworkTest(endpointID string) (*TestResult, error) {
    start := time.Now()

    a.monitoringLogger.TestStarted(endpointID, "manual")
    a.analyzer.RecordEvent("test_started")

    result, err := a.executeTest(endpointID)

    duration := time.Since(start)
    a.performanceLogger.APICall("RunNetworkTest", duration, err)

    if err != nil {
        a.logger.Error("Network test failed",
            String("endpoint_id", endpointID),
            Duration("duration", duration),
            Error(err))
        return nil, err
    }

    a.monitoringLogger.TestCompleted(endpointID, result)
    a.analyzer.RecordEvent("test_completed")

    return result, nil
}

// Log configuration methods
func (a *App) UpdateLogLevel(level string) error {
    a.securityLogger.ConfigurationChanged("system", "logging",
        map[string]interface{}{"level": level})

    // Update actual log level
    return a.updateLoggerLevel(level)
}

func (a *App) GetLogMetrics() (map[string]*LogMetric, error) {
    return a.analyzer.GetMetrics(), nil
}

func (a *App) ExportLogs(startTime, endTime time.Time, format string) ([]byte, error) {
    a.logger.Info("Log export requested",
        Field{Key: "start_time", Value: startTime},
        Field{Key: "end_time", Value: endTime},
        String("format", format))

    // Implementation for log export
    return nil, nil
}
```

## Configuration Management
```go
// Log configuration with validation
func (a *App) UpdateLogConfig(config *LogConfig) error {
    // Validate configuration
    if err := validateLogConfig(config); err != nil {
        return fmt.Errorf("invalid log configuration: %v", err)
    }

    // Log the configuration change
    a.securityLogger.ConfigurationChanged("user", "logging", map[string]interface{}{
        "level":       config.Level,
        "format":      config.Format,
        "output":      config.Output,
        "max_size":    config.MaxSize,
        "max_backups": config.MaxBackups,
    })

    // Apply new configuration
    newLogger, err := NewLogger(config)
    if err != nil {
        return fmt.Errorf("failed to create new logger: %v", err)
    }

    // Replace current logger
    a.logger = newLogger
    a.monitoringLogger = NewMonitoringLogger(newLogger)
    a.performanceLogger = NewPerformanceLogger(newLogger)
    a.securityLogger = NewSecurityLogger(newLogger)

    a.logger.Info("Log configuration updated successfully")

    return nil
}

func validateLogConfig(config *LogConfig) error {
    validLevels := []string{"debug", "info", "warn", "error", "fatal"}
    validFormats := []string{"json", "text"}
    validOutputs := []string{"file", "console", "both"}

    if !contains(validLevels, config.Level) {
        return fmt.Errorf("invalid log level: %s", config.Level)
    }

    if !contains(validFormats, config.Format) {
        return fmt.Errorf("invalid log format: %s", config.Format)
    }

    if !contains(validOutputs, config.Output) {
        return fmt.Errorf("invalid log output: %s", config.Output)
    }

    if (config.Output == "file" || config.Output == "both") && config.FilePath == "" {
        return fmt.Errorf("file path is required for file output")
    }

    if config.MaxSize <= 0 {
        return fmt.Errorf("max size must be positive")
    }

    return nil
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

## Verification Steps
1. Test log output formats - should write in JSON and text formats correctly
2. Verify log rotation - should rotate logs based on size and time
3. Test different log levels - should filter messages based on configured level
4. Verify structured logging - should include all specified fields
5. Test performance impact - should not significantly impact application performance
6. Verify security sanitization - should redact sensitive information
7. Test cross-platform functionality - should work on all supported platforms
8. Verify log analysis tools - should provide useful metrics and insights

## Dependencies
- T002: Basic Application Structure
- T003: Configuration System
- T038: Cross-Platform Compatibility
- logrus or similar logging library
- lumberjack for log rotation

## Notes
- Consider implementing log streaming for remote monitoring
- Plan for future integration with centralized logging systems
- Implement proper log retention policies
- Consider adding structured query capabilities for logs
- Plan for log compression and archival
- Ensure logs provide enough information for debugging without being too verbose