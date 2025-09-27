# T040: Error Handling and Recovery

## Overview
Implement comprehensive error handling and recovery mechanisms throughout NetMonitor to ensure graceful degradation, automatic recovery, and detailed error reporting.

## Context
NetMonitor needs robust error handling to maintain reliable operation in the face of network issues, system problems, and unexpected failures. The system should recover gracefully and provide useful error information for debugging.

## Task Description
Create a comprehensive error handling framework with automatic recovery mechanisms, error classification, retry logic, and user-friendly error reporting across all application components.

## Acceptance Criteria
- [ ] Structured error handling with error types and codes
- [ ] Automatic retry mechanisms with exponential backoff
- [ ] Circuit breaker pattern for external dependencies
- [ ] Graceful degradation when components fail
- [ ] Error recovery and self-healing capabilities
- [ ] Comprehensive error logging and monitoring
- [ ] User-friendly error messages and guidance
- [ ] Error aggregation and reporting
- [ ] Testing framework for error scenarios

## Error Framework Architecture
```go
package errors

import (
    "context"
    "fmt"
    "time"
)

// Error types and classification
type ErrorType string

const (
    ErrorTypeNetwork      ErrorType = "network"
    ErrorTypeConfiguration ErrorType = "configuration"
    ErrorTypeStorage      ErrorType = "storage"
    ErrorTypePermission   ErrorType = "permission"
    ErrorTypeTimeout      ErrorType = "timeout"
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypeSystem       ErrorType = "system"
    ErrorTypeInternal     ErrorType = "internal"
)

type ErrorSeverity string

const (
    SeverityLow      ErrorSeverity = "low"
    SeverityMedium   ErrorSeverity = "medium"
    SeverityHigh     ErrorSeverity = "high"
    SeverityCritical ErrorSeverity = "critical"
)

type NetMonitorError struct {
    Type        ErrorType     `json:"type"`
    Code        string        `json:"code"`
    Message     string        `json:"message"`
    Severity    ErrorSeverity `json:"severity"`
    Component   string        `json:"component"`
    Operation   string        `json:"operation"`
    Context     map[string]interface{} `json:"context,omitempty"`
    Cause       error         `json:"cause,omitempty"`
    Timestamp   time.Time     `json:"timestamp"`
    Recoverable bool          `json:"recoverable"`
    UserMessage string        `json:"userMessage,omitempty"`
}

func (e *NetMonitorError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *NetMonitorError) Unwrap() error {
    return e.Cause
}

func (e *NetMonitorError) Is(target error) bool {
    if t, ok := target.(*NetMonitorError); ok {
        return e.Type == t.Type && e.Code == t.Code
    }
    return false
}

func (e *NetMonitorError) WithContext(key string, value interface{}) *NetMonitorError {
    if e.Context == nil {
        e.Context = make(map[string]interface{})
    }
    e.Context[key] = value
    return e
}

func (e *NetMonitorError) WithUserMessage(message string) *NetMonitorError {
    e.UserMessage = message
    return e
}

// Error constructors
func NewNetworkError(code, message string, cause error) *NetMonitorError {
    return &NetMonitorError{
        Type:        ErrorTypeNetwork,
        Code:        code,
        Message:     message,
        Severity:    SeverityMedium,
        Cause:       cause,
        Timestamp:   time.Now(),
        Recoverable: true,
    }
}

func NewConfigurationError(code, message string) *NetMonitorError {
    return &NetMonitorError{
        Type:        ErrorTypeConfiguration,
        Code:        code,
        Message:     message,
        Severity:    SeverityHigh,
        Timestamp:   time.Now(),
        Recoverable: false,
        UserMessage: "Please check your configuration settings.",
    }
}

func NewStorageError(code, message string, cause error) *NetMonitorError {
    return &NetMonitorError{
        Type:        ErrorTypeStorage,
        Code:        code,
        Message:     message,
        Severity:    SeverityHigh,
        Cause:       cause,
        Timestamp:   time.Now(),
        Recoverable: true,
    }
}

func NewTimeoutError(operation string, timeout time.Duration) *NetMonitorError {
    return &NetMonitorError{
        Type:        ErrorTypeTimeout,
        Code:        "TIMEOUT",
        Message:     fmt.Sprintf("Operation %s timed out after %v", operation, timeout),
        Severity:    SeverityMedium,
        Operation:   operation,
        Timestamp:   time.Now(),
        Recoverable: true,
        Context:     map[string]interface{}{"timeout": timeout.String()},
    }
}

func NewValidationError(field, message string) *NetMonitorError {
    return &NetMonitorError{
        Type:        ErrorTypeValidation,
        Code:        "VALIDATION_FAILED",
        Message:     fmt.Sprintf("Validation failed for %s: %s", field, message),
        Severity:    SeverityLow,
        Timestamp:   time.Now(),
        Recoverable: false,
        Context:     map[string]interface{}{"field": field},
        UserMessage: fmt.Sprintf("Invalid %s: %s", field, message),
    }
}
```

## Retry Mechanism
```go
package retry

import (
    "context"
    "fmt"
    "math"
    "time"
    "math/rand"
)

type RetryConfig struct {
    MaxAttempts     int           `json:"maxAttempts"`
    InitialDelay    time.Duration `json:"initialDelay"`
    MaxDelay        time.Duration `json:"maxDelay"`
    BackoffFactor   float64       `json:"backoffFactor"`
    Jitter          bool          `json:"jitter"`
    RetryableErrors []string      `json:"retryableErrors"`
}

type RetryableFunc func() error

func (rc *RetryConfig) Execute(ctx context.Context, fn RetryableFunc) error {
    var lastErr error

    for attempt := 1; attempt <= rc.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil // Success
        }

        lastErr = err

        // Check if error is retryable
        if !rc.isRetryable(err) {
            return err
        }

        // Don't sleep after the last attempt
        if attempt == rc.MaxAttempts {
            break
        }

        // Calculate delay with exponential backoff
        delay := rc.calculateDelay(attempt)

        // Check if context is cancelled
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }

    return fmt.Errorf("max retry attempts (%d) exceeded, last error: %w",
                     rc.MaxAttempts, lastErr)
}

func (rc *RetryConfig) calculateDelay(attempt int) time.Duration {
    // Exponential backoff: initialDelay * (backoffFactor ^ (attempt - 1))
    delay := float64(rc.InitialDelay) * math.Pow(rc.BackoffFactor, float64(attempt-1))

    // Apply maximum delay limit
    if time.Duration(delay) > rc.MaxDelay {
        delay = float64(rc.MaxDelay)
    }

    // Add jitter to prevent thundering herd
    if rc.Jitter {
        jitter := delay * 0.1 * rand.Float64() // 10% jitter
        delay += jitter
    }

    return time.Duration(delay)
}

func (rc *RetryConfig) isRetryable(err error) bool {
    if len(rc.RetryableErrors) == 0 {
        // Default retryable conditions
        return isDefaultRetryable(err)
    }

    // Check against configured retryable error codes
    if nmErr, ok := err.(*NetMonitorError); ok {
        for _, code := range rc.RetryableErrors {
            if nmErr.Code == code {
                return true
            }
        }
    }

    return false
}

func isDefaultRetryable(err error) bool {
    if nmErr, ok := err.(*NetMonitorError); ok {
        switch nmErr.Type {
        case ErrorTypeNetwork, ErrorTypeTimeout:
            return true
        case ErrorTypeStorage:
            return nmErr.Recoverable
        default:
            return false
        }
    }

    // Check for common retryable errors
    return isNetworkError(err) || isTimeoutError(err)
}

// Default retry configurations
var (
    DefaultNetworkRetry = &RetryConfig{
        MaxAttempts:   3,
        InitialDelay:  1 * time.Second,
        MaxDelay:      30 * time.Second,
        BackoffFactor: 2.0,
        Jitter:        true,
    }

    DefaultStorageRetry = &RetryConfig{
        MaxAttempts:   5,
        InitialDelay:  500 * time.Millisecond,
        MaxDelay:      10 * time.Second,
        BackoffFactor: 1.5,
        Jitter:        true,
    }

    QuickRetry = &RetryConfig{
        MaxAttempts:   2,
        InitialDelay:  100 * time.Millisecond,
        MaxDelay:      1 * time.Second,
        BackoffFactor: 2.0,
        Jitter:        false,
    }
)
```

## Circuit Breaker Pattern
```go
package circuitbreaker

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    name           string
    config         *Config
    state          State
    failureCount   int
    successCount   int
    lastFailTime   time.Time
    lastStateChange time.Time
    mu             sync.RWMutex
}

type Config struct {
    MaxFailures     int           `json:"maxFailures"`
    Timeout         time.Duration `json:"timeout"`
    ResetTimeout    time.Duration `json:"resetTimeout"`
    SuccessThreshold int          `json:"successThreshold"`
}

type CallFunc func() (interface{}, error)

func NewCircuitBreaker(name string, config *Config) *CircuitBreaker {
    if config == nil {
        config = &Config{
            MaxFailures:      5,
            Timeout:          30 * time.Second,
            ResetTimeout:     60 * time.Second,
            SuccessThreshold: 2,
        }
    }

    return &CircuitBreaker{
        name:            name,
        config:          config,
        state:           StateClosed,
        lastStateChange: time.Now(),
    }
}

func (cb *CircuitBreaker) Call(ctx context.Context, fn CallFunc) (interface{}, error) {
    state := cb.getState()

    switch state {
    case StateOpen:
        return nil, NewCircuitBreakerError("circuit breaker is open", cb.name)
    case StateHalfOpen:
        return cb.callHalfOpen(ctx, fn)
    default: // StateClosed
        return cb.callClosed(ctx, fn)
    }
}

func (cb *CircuitBreaker) callClosed(ctx context.Context, fn CallFunc) (interface{}, error) {
    result, err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failureCount++
        cb.lastFailTime = time.Now()

        if cb.failureCount >= cb.config.MaxFailures {
            cb.state = StateOpen
            cb.lastStateChange = time.Now()
        }
        return nil, err
    }

    // Success - reset failure count
    cb.failureCount = 0
    return result, nil
}

func (cb *CircuitBreaker) callHalfOpen(ctx context.Context, fn CallFunc) (interface{}, error) {
    result, err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.state = StateOpen
        cb.lastStateChange = time.Now()
        cb.failureCount++
        return nil, err
    }

    cb.successCount++
    if cb.successCount >= cb.config.SuccessThreshold {
        cb.state = StateClosed
        cb.failureCount = 0
        cb.successCount = 0
        cb.lastStateChange = time.Now()
    }

    return result, nil
}

func (cb *CircuitBreaker) getState() State {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    if cb.state == StateOpen {
        if time.Since(cb.lastStateChange) > cb.config.ResetTimeout {
            cb.mu.RUnlock()
            cb.mu.Lock()
            if cb.state == StateOpen && time.Since(cb.lastStateChange) > cb.config.ResetTimeout {
                cb.state = StateHalfOpen
                cb.successCount = 0
                cb.lastStateChange = time.Now()
            }
            cb.mu.Unlock()
            cb.mu.RLock()
        }
    }

    return cb.state
}

func (cb *CircuitBreaker) GetStats() *Stats {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    return &Stats{
        Name:            cb.name,
        State:           cb.state,
        FailureCount:    cb.failureCount,
        SuccessCount:    cb.successCount,
        LastFailTime:    cb.lastFailTime,
        LastStateChange: cb.lastStateChange,
    }
}

type Stats struct {
    Name            string    `json:"name"`
    State           State     `json:"state"`
    FailureCount    int       `json:"failureCount"`
    SuccessCount    int       `json:"successCount"`
    LastFailTime    time.Time `json:"lastFailTime"`
    LastStateChange time.Time `json:"lastStateChange"`
}

type CircuitBreakerError struct {
    message string
    name    string
}

func (e *CircuitBreakerError) Error() string {
    return fmt.Sprintf("circuit breaker '%s': %s", e.name, e.message)
}

func NewCircuitBreakerError(message, name string) *CircuitBreakerError {
    return &CircuitBreakerError{
        message: message,
        name:    name,
    }
}
```

## Error Recovery Manager
```go
package recovery

import (
    "context"
    "sync"
    "time"
)

type RecoveryManager struct {
    recoveryStrategies map[string]RecoveryStrategy
    circuitBreakers    map[string]*CircuitBreaker
    mu                 sync.RWMutex
    logger             Logger
}

type RecoveryStrategy interface {
    CanRecover(err error) bool
    Recover(ctx context.Context, err error) error
    Priority() int
}

func NewRecoveryManager(logger Logger) *RecoveryManager {
    rm := &RecoveryManager{
        recoveryStrategies: make(map[string]RecoveryStrategy),
        circuitBreakers:    make(map[string]*CircuitBreaker),
        logger:             logger,
    }

    // Register default recovery strategies
    rm.RegisterStrategy("network", &NetworkRecoveryStrategy{logger: logger})
    rm.RegisterStrategy("storage", &StorageRecoveryStrategy{logger: logger})
    rm.RegisterStrategy("configuration", &ConfigurationRecoveryStrategy{logger: logger})

    return rm
}

func (rm *RecoveryManager) RegisterStrategy(name string, strategy RecoveryStrategy) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    rm.recoveryStrategies[name] = strategy
}

func (rm *RecoveryManager) HandleError(ctx context.Context, err error) error {
    rm.logger.Error("Handling error for recovery",
        Field{Key: "error", Value: err.Error()})

    // Try recovery strategies in priority order
    strategies := rm.getOrderedStrategies()

    for _, strategy := range strategies {
        if strategy.CanRecover(err) {
            rm.logger.Info("Attempting recovery",
                Field{Key: "strategy", Value: fmt.Sprintf("%T", strategy)})

            if recoveryErr := strategy.Recover(ctx, err); recoveryErr == nil {
                rm.logger.Info("Recovery successful")
                return nil
            } else {
                rm.logger.Warn("Recovery failed",
                    Field{Key: "recovery_error", Value: recoveryErr.Error()})
            }
        }
    }

    rm.logger.Error("All recovery strategies failed")
    return err
}

func (rm *RecoveryManager) getOrderedStrategies() []RecoveryStrategy {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    strategies := make([]RecoveryStrategy, 0, len(rm.recoveryStrategies))
    for _, strategy := range rm.recoveryStrategies {
        strategies = append(strategies, strategy)
    }

    // Sort by priority (higher priority first)
    sort.Slice(strategies, func(i, j int) bool {
        return strategies[i].Priority() > strategies[j].Priority()
    })

    return strategies
}

// Network Recovery Strategy
type NetworkRecoveryStrategy struct {
    logger Logger
}

func (nrs *NetworkRecoveryStrategy) CanRecover(err error) bool {
    if nmErr, ok := err.(*NetMonitorError); ok {
        return nmErr.Type == ErrorTypeNetwork && nmErr.Recoverable
    }
    return isNetworkError(err)
}

func (nrs *NetworkRecoveryStrategy) Recover(ctx context.Context, err error) error {
    nrs.logger.Info("Attempting network recovery")

    // Wait a moment for network issues to resolve
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(2 * time.Second):
    }

    // Test basic connectivity
    if err := testConnectivity(ctx); err != nil {
        return fmt.Errorf("network connectivity test failed: %w", err)
    }

    return nil
}

func (nrs *NetworkRecoveryStrategy) Priority() int {
    return 10
}

// Storage Recovery Strategy
type StorageRecoveryStrategy struct {
    logger Logger
}

func (srs *StorageRecoveryStrategy) CanRecover(err error) bool {
    if nmErr, ok := err.(*NetMonitorError); ok {
        return nmErr.Type == ErrorTypeStorage && nmErr.Recoverable
    }
    return false
}

func (srs *StorageRecoveryStrategy) Recover(ctx context.Context, err error) error {
    srs.logger.Info("Attempting storage recovery")

    // Try to recreate directories, repair indexes, etc.
    if err := repairStorage(ctx); err != nil {
        return fmt.Errorf("storage repair failed: %w", err)
    }

    return nil
}

func (srs *StorageRecoveryStrategy) Priority() int {
    return 8
}
```

## Graceful Degradation
```go
package degradation

import (
    "context"
    "sync"
)

type DegradationManager struct {
    features map[string]*FeatureState
    mu       sync.RWMutex
    logger   Logger
}

type FeatureState struct {
    Name        string    `json:"name"`
    Enabled     bool      `json:"enabled"`
    Degraded    bool      `json:"degraded"`
    LastError   error     `json:"-"`
    ErrorCount  int       `json:"errorCount"`
    LastAttempt time.Time `json:"lastAttempt"`
}

func NewDegradationManager(logger Logger) *DegradationManager {
    return &DegradationManager{
        features: make(map[string]*FeatureState),
        logger:   logger,
    }
}

func (dm *DegradationManager) RegisterFeature(name string) {
    dm.mu.Lock()
    defer dm.mu.Unlock()

    dm.features[name] = &FeatureState{
        Name:    name,
        Enabled: true,
        Degraded: false,
    }
}

func (dm *DegradationManager) RecordError(featureName string, err error) {
    dm.mu.Lock()
    defer dm.mu.Unlock()

    feature, exists := dm.features[featureName]
    if !exists {
        return
    }

    feature.ErrorCount++
    feature.LastError = err
    feature.LastAttempt = time.Now()

    // Determine if feature should be degraded
    if feature.ErrorCount >= 3 {
        if !feature.Degraded {
            dm.logger.Warn("Feature degraded due to repeated errors",
                Field{Key: "feature", Value: featureName},
                Field{Key: "error_count", Value: feature.ErrorCount})
        }
        feature.Degraded = true
    }
}

func (dm *DegradationManager) RecordSuccess(featureName string) {
    dm.mu.Lock()
    defer dm.mu.Unlock()

    feature, exists := dm.features[featureName]
    if !exists {
        return
    }

    // Reset error count on success
    feature.ErrorCount = 0
    feature.LastError = nil

    if feature.Degraded {
        dm.logger.Info("Feature recovered from degraded state",
            Field{Key: "feature", Value: featureName})
        feature.Degraded = false
    }
}

func (dm *DegradationManager) IsFeatureAvailable(featureName string) bool {
    dm.mu.RLock()
    defer dm.mu.RUnlock()

    feature, exists := dm.features[featureName]
    if !exists {
        return false
    }

    return feature.Enabled && !feature.Degraded
}

func (dm *DegradationManager) GetFeatureState(featureName string) *FeatureState {
    dm.mu.RLock()
    defer dm.mu.RUnlock()

    if feature, exists := dm.features[featureName]; exists {
        return &FeatureState{
            Name:        feature.Name,
            Enabled:     feature.Enabled,
            Degraded:    feature.Degraded,
            ErrorCount:  feature.ErrorCount,
            LastAttempt: feature.LastAttempt,
        }
    }

    return nil
}

func (dm *DegradationManager) GetAllFeatures() map[string]*FeatureState {
    dm.mu.RLock()
    defer dm.mu.RUnlock()

    result := make(map[string]*FeatureState)
    for name, feature := range dm.features {
        result[name] = &FeatureState{
            Name:        feature.Name,
            Enabled:     feature.Enabled,
            Degraded:    feature.Degraded,
            ErrorCount:  feature.ErrorCount,
            LastAttempt: feature.LastAttempt,
        }
    }

    return result
}
```

## Application Integration
```go
// App integration with error handling
func (a *App) initializeErrorHandling() {
    a.recoveryManager = NewRecoveryManager(a.logger)
    a.degradationManager = NewDegradationManager(a.logger)

    // Register features for degradation management
    a.degradationManager.RegisterFeature("network_monitoring")
    a.degradationManager.RegisterFeature("data_storage")
    a.degradationManager.RegisterFeature("notifications")
    a.degradationManager.RegisterFeature("system_tray")

    // Setup circuit breakers for external dependencies
    a.circuitBreakers = map[string]*CircuitBreaker{
        "network_tests": NewCircuitBreaker("network_tests", &Config{
            MaxFailures:  5,
            Timeout:      30 * time.Second,
            ResetTimeout: 60 * time.Second,
        }),
        "data_storage": NewCircuitBreaker("data_storage", &Config{
            MaxFailures:  3,
            Timeout:      10 * time.Second,
            ResetTimeout: 30 * time.Second,
        }),
    }
}

// Example of using error handling in network tests
func (a *App) executeNetworkTest(endpointID string) (*TestResult, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Use circuit breaker for network tests
    result, err := a.circuitBreakers["network_tests"].Call(ctx, func() (interface{}, error) {
        return a.performNetworkTest(ctx, endpointID)
    })

    if err != nil {
        // Record error for feature degradation
        a.degradationManager.RecordError("network_monitoring", err)

        // Try recovery
        if recoveryErr := a.recoveryManager.HandleError(ctx, err); recoveryErr == nil {
            // Retry after successful recovery
            return a.performNetworkTest(ctx, endpointID)
        }

        return nil, err
    }

    // Record success
    a.degradationManager.RecordSuccess("network_monitoring")
    return result.(*TestResult), nil
}

// Error reporting to frontend
func (a *App) GetErrorStatus() (*ErrorStatus, error) {
    features := a.degradationManager.GetAllFeatures()
    circuitStats := make(map[string]*Stats)

    for name, cb := range a.circuitBreakers {
        circuitStats[name] = cb.GetStats()
    }

    return &ErrorStatus{
        Features:        features,
        CircuitBreakers: circuitStats,
        Timestamp:       time.Now(),
    }, nil
}

type ErrorStatus struct {
    Features        map[string]*FeatureState `json:"features"`
    CircuitBreakers map[string]*Stats        `json:"circuitBreakers"`
    Timestamp       time.Time                `json:"timestamp"`
}
```

## Testing Framework for Error Scenarios
```go
// Error injection for testing
type ErrorInjector struct {
    rules map[string]*InjectionRule
    mu    sync.RWMutex
}

type InjectionRule struct {
    ErrorType   ErrorType
    Probability float64
    Duration    time.Duration
    Active      bool
}

func (ei *ErrorInjector) InjectError(operation string) error {
    ei.mu.RLock()
    rule, exists := ei.rules[operation]
    ei.mu.RUnlock()

    if !exists || !rule.Active {
        return nil
    }

    if rand.Float64() < rule.Probability {
        switch rule.ErrorType {
        case ErrorTypeNetwork:
            return NewNetworkError("INJECTED_ERROR", "Injected network error", nil)
        case ErrorTypeTimeout:
            return NewTimeoutError(operation, rule.Duration)
        default:
            return fmt.Errorf("injected error for %s", operation)
        }
    }

    return nil
}
```

## Verification Steps
1. Test error classification - should categorize errors correctly
2. Verify retry mechanisms - should retry with exponential backoff
3. Test circuit breaker functionality - should open/close based on failure rates
4. Verify graceful degradation - should disable features when failing
5. Test recovery mechanisms - should attempt automatic recovery
6. Verify error logging - should log errors with appropriate detail
7. Test user error messages - should provide helpful guidance
8. Verify error aggregation - should collect and report error patterns

## Dependencies
- T039: Comprehensive Logging System
- T002: Basic Application Structure
- T015: Monitoring Status API

## Notes
- Implement comprehensive error testing scenarios
- Consider implementing error budgets for SLA management
- Plan for error alerting and monitoring integration
- Ensure error handling doesn't impact performance significantly
- Consider implementing chaos engineering practices for testing
- Plan for error analytics and trend analysis