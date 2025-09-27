# T042: Alert Management System

## Overview
Implement a comprehensive alert management system that handles alert rules, escalation policies, alert suppression, and integration with notification channels for intelligent monitoring alerting.

## Context
NetMonitor needs sophisticated alert management beyond basic notifications. This includes configurable alert rules, escalation policies, alert correlation, suppression during maintenance, and integration with multiple notification channels.

## Task Description
Create an advanced alert management framework with rule-based alerting, escalation policies, alert correlation, suppression mechanisms, and comprehensive alert lifecycle management.

## Acceptance Criteria
- [ ] Configurable alert rules with conditions and thresholds
- [ ] Alert escalation policies with multiple notification channels
- [ ] Alert suppression and maintenance windows
- [ ] Alert correlation and deduplication
- [ ] Alert lifecycle management (open, acknowledged, resolved)
- [ ] Historical alert tracking and analysis
- [ ] Integration with external alerting systems
- [ ] Alert dashboard and management interface
- [ ] Performance impact monitoring and optimization

## Alert Management Architecture
```go
package alerts

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type AlertManager struct {
    rules           map[string]*AlertRule
    alerts          map[string]*Alert
    escalations     map[string]*EscalationPolicy
    suppressions    map[string]*SuppressionRule
    correlations    map[string]*CorrelationRule
    notificationMgr NotificationManager
    storage         AlertStorage
    mutex           sync.RWMutex
    logger          Logger
}

type AlertRule struct {
    ID              string               `json:"id"`
    Name            string               `json:"name"`
    Description     string               `json:"description"`
    Enabled         bool                 `json:"enabled"`
    Conditions      []AlertCondition     `json:"conditions"`
    Severity        AlertSeverity        `json:"severity"`
    EvaluationInterval time.Duration     `json:"evaluationInterval"`
    ForDuration     time.Duration        `json:"forDuration"` // Alert must persist for this duration
    Labels          map[string]string    `json:"labels"`
    Annotations     map[string]string    `json:"annotations"`
    EscalationPolicy string              `json:"escalationPolicy"`
    AutoResolve     bool                 `json:"autoResolve"`
    CreatedAt       time.Time            `json:"createdAt"`
    UpdatedAt       time.Time            `json:"updatedAt"`
}

type AlertCondition struct {
    Type        ConditionType        `json:"type"`
    Metric      string               `json:"metric"`
    Operator    ComparisonOperator   `json:"operator"`
    Value       float64              `json:"value"`
    TimeWindow  time.Duration        `json:"timeWindow"`
    Aggregation AggregationType      `json:"aggregation"`
    Filters     map[string]string    `json:"filters"`
}

type ConditionType string

const (
    ConditionThreshold      ConditionType = "threshold"
    ConditionAnomaly        ConditionType = "anomaly"
    ConditionRateOfChange   ConditionType = "rate_of_change"
    ConditionMissingData    ConditionType = "missing_data"
    ConditionServiceHealth  ConditionType = "service_health"
)

type ComparisonOperator string

const (
    OperatorGreaterThan    ComparisonOperator = "gt"
    OperatorLessThan       ComparisonOperator = "lt"
    OperatorEquals         ComparisonOperator = "eq"
    OperatorNotEquals      ComparisonOperator = "ne"
    OperatorGreaterOrEqual ComparisonOperator = "gte"
    OperatorLessOrEqual    ComparisonOperator = "lte"
)

type AggregationType string

const (
    AggregationAverage AggregationType = "avg"
    AggregationSum     AggregationType = "sum"
    AggregationMin     AggregationType = "min"
    AggregationMax     AggregationType = "max"
    AggregationCount   AggregationType = "count"
    AggregationP95     AggregationType = "p95"
    AggregationP99     AggregationType = "p99"
)

type Alert struct {
    ID             string                 `json:"id"`
    RuleID         string                 `json:"ruleId"`
    RuleName       string                 `json:"ruleName"`
    Status         AlertStatus            `json:"status"`
    Severity       AlertSeverity          `json:"severity"`
    Message        string                 `json:"message"`
    Description    string                 `json:"description"`
    Labels         map[string]string      `json:"labels"`
    Annotations    map[string]string      `json:"annotations"`
    Source         AlertSource            `json:"source"`
    StartsAt       time.Time              `json:"startsAt"`
    EndsAt         *time.Time             `json:"endsAt,omitempty"`
    UpdatedAt      time.Time              `json:"updatedAt"`
    AcknowledgedAt *time.Time             `json:"acknowledgedAt,omitempty"`
    AcknowledgedBy string                 `json:"acknowledgedBy,omitempty"`
    ResolvedAt     *time.Time             `json:"resolvedAt,omitempty"`
    ResolvedBy     string                 `json:"resolvedBy,omitempty"`
    EscalationLevel int                   `json:"escalationLevel"`
    NotificationsSent []NotificationRecord `json:"notificationsSent"`
    Fingerprint    string                 `json:"fingerprint"`
}

type AlertStatus string

const (
    AlertStatusFiring      AlertStatus = "firing"
    AlertStatusAcknowledged AlertStatus = "acknowledged"
    AlertStatusResolved    AlertStatus = "resolved"
    AlertStatusSuppressed  AlertStatus = "suppressed"
)

type AlertSeverity string

const (
    SeverityInfo     AlertSeverity = "info"
    SeverityWarning  AlertSeverity = "warning"
    SeverityCritical AlertSeverity = "critical"
    SeverityEmergency AlertSeverity = "emergency"
)

type AlertSource struct {
    Type       string            `json:"type"`
    EndpointID string            `json:"endpointId,omitempty"`
    Region     string            `json:"region,omitempty"`
    Component  string            `json:"component,omitempty"`
    Additional map[string]string `json:"additional,omitempty"`
}

func NewAlertManager(notificationMgr NotificationManager, storage AlertStorage, logger Logger) *AlertManager {
    return &AlertManager{
        rules:           make(map[string]*AlertRule),
        alerts:          make(map[string]*Alert),
        escalations:     make(map[string]*EscalationPolicy),
        suppressions:    make(map[string]*SuppressionRule),
        correlations:    make(map[string]*CorrelationRule),
        notificationMgr: notificationMgr,
        storage:         storage,
        logger:          logger,
    }
}

func (am *AlertManager) EvaluateRules(ctx context.Context, data *MonitoringData) error {
    am.mutex.RLock()
    rules := make([]*AlertRule, 0, len(am.rules))
    for _, rule := range am.rules {
        if rule.Enabled {
            rules = append(rules, rule)
        }
    }
    am.mutex.RUnlock()

    for _, rule := range rules {
        if err := am.evaluateRule(ctx, rule, data); err != nil {
            am.logger.Error("Failed to evaluate alert rule",
                Field{Key: "rule_id", Value: rule.ID},
                Field{Key: "error", Value: err.Error()})
        }
    }

    return nil
}

func (am *AlertManager) evaluateRule(ctx context.Context, rule *AlertRule, data *MonitoringData) error {
    // Evaluate all conditions for the rule
    conditionResults := make([]bool, len(rule.Conditions))

    for i, condition := range rule.Conditions {
        result, err := am.evaluateCondition(condition, data)
        if err != nil {
            return fmt.Errorf("failed to evaluate condition %d: %w", i, err)
        }
        conditionResults[i] = result
    }

    // Check if all conditions are met (AND logic)
    allConditionsMet := true
    for _, result := range conditionResults {
        if !result {
            allConditionsMet = false
            break
        }
    }

    alertFingerprint := am.generateAlertFingerprint(rule, data)
    existingAlert := am.findAlertByFingerprint(alertFingerprint)

    if allConditionsMet {
        if existingAlert == nil {
            // Create new alert
            alert := &Alert{
                ID:           generateAlertID(),
                RuleID:       rule.ID,
                RuleName:     rule.Name,
                Status:       AlertStatusFiring,
                Severity:     rule.Severity,
                Message:      am.generateAlertMessage(rule, data),
                Description:  rule.Description,
                Labels:       am.mergeLabels(rule.Labels, data),
                Annotations:  rule.Annotations,
                Source:       am.extractAlertSource(data),
                StartsAt:     time.Now(),
                UpdatedAt:    time.Now(),
                Fingerprint:  alertFingerprint,
            }

            if err := am.createAlert(ctx, alert); err != nil {
                return fmt.Errorf("failed to create alert: %w", err)
            }
        } else if existingAlert.Status == AlertStatusResolved {
            // Reopen resolved alert
            existingAlert.Status = AlertStatusFiring
            existingAlert.UpdatedAt = time.Now()
            existingAlert.EndsAt = nil
            existingAlert.ResolvedAt = nil
            existingAlert.ResolvedBy = ""

            if err := am.updateAlert(ctx, existingAlert); err != nil {
                return fmt.Errorf("failed to reopen alert: %w", err)
            }
        }
    } else {
        // Conditions not met - auto-resolve if enabled
        if existingAlert != nil && existingAlert.Status == AlertStatusFiring && rule.AutoResolve {
            if err := am.resolveAlert(ctx, existingAlert.ID, "auto-resolved"); err != nil {
                return fmt.Errorf("failed to auto-resolve alert: %w", err)
            }
        }
    }

    return nil
}

func (am *AlertManager) evaluateCondition(condition AlertCondition, data *MonitoringData) (bool, error) {
    switch condition.Type {
    case ConditionThreshold:
        return am.evaluateThresholdCondition(condition, data)
    case ConditionAnomaly:
        return am.evaluateAnomalyCondition(condition, data)
    case ConditionRateOfChange:
        return am.evaluateRateOfChangeCondition(condition, data)
    case ConditionMissingData:
        return am.evaluateMissingDataCondition(condition, data)
    case ConditionServiceHealth:
        return am.evaluateServiceHealthCondition(condition, data)
    default:
        return false, fmt.Errorf("unknown condition type: %s", condition.Type)
    }
}

func (am *AlertManager) evaluateThresholdCondition(condition AlertCondition, data *MonitoringData) (bool, error) {
    // Get metric value based on condition configuration
    value, err := am.getMetricValue(condition.Metric, condition.Filters, condition.TimeWindow, condition.Aggregation, data)
    if err != nil {
        return false, err
    }

    // Compare against threshold
    switch condition.Operator {
    case OperatorGreaterThan:
        return value > condition.Value, nil
    case OperatorLessThan:
        return value < condition.Value, nil
    case OperatorEquals:
        return value == condition.Value, nil
    case OperatorNotEquals:
        return value != condition.Value, nil
    case OperatorGreaterOrEqual:
        return value >= condition.Value, nil
    case OperatorLessOrEqual:
        return value <= condition.Value, nil
    default:
        return false, fmt.Errorf("unknown operator: %s", condition.Operator)
    }
}

func (am *AlertManager) createAlert(ctx context.Context, alert *Alert) error {
    am.mutex.Lock()
    am.alerts[alert.ID] = alert
    am.mutex.Unlock()

    // Store in persistent storage
    if err := am.storage.SaveAlert(alert); err != nil {
        am.logger.Error("Failed to save alert to storage", Field{Key: "alert_id", Value: alert.ID}, Error(err))
    }

    // Check if alert should be suppressed
    if am.isAlertSuppressed(alert) {
        alert.Status = AlertStatusSuppressed
        am.logger.Info("Alert suppressed", Field{Key: "alert_id", Value: alert.ID})
        return nil
    }

    // Start escalation process
    if err := am.startEscalation(ctx, alert); err != nil {
        am.logger.Error("Failed to start escalation", Field{Key: "alert_id", Value: alert.ID}, Error(err))
    }

    am.logger.Info("Alert created",
        Field{Key: "alert_id", Value: alert.ID},
        Field{Key: "rule_id", Value: alert.RuleID},
        Field{Key: "severity", Value: string(alert.Severity)})

    return nil
}

// Escalation Policy Management
type EscalationPolicy struct {
    ID          string             `json:"id"`
    Name        string             `json:"name"`
    Description string             `json:"description"`
    Steps       []EscalationStep   `json:"steps"`
    RepeatCount int                `json:"repeatCount"` // How many times to repeat the policy
    CreatedAt   time.Time          `json:"createdAt"`
    UpdatedAt   time.Time          `json:"updatedAt"`
}

type EscalationStep struct {
    StepNumber      int               `json:"stepNumber"`
    WaitDuration    time.Duration     `json:"waitDuration"`
    NotificationChannels []string     `json:"notificationChannels"`
    Conditions      []StepCondition   `json:"conditions"`
}

type StepCondition struct {
    Type  string `json:"type"`  // "severity", "duration", "acknowledge_timeout"
    Value string `json:"value"`
}

func (am *AlertManager) startEscalation(ctx context.Context, alert *Alert) error {
    am.mutex.RLock()
    policy, exists := am.escalations[alert.RuleID]
    am.mutex.RUnlock()

    if !exists {
        // Use default escalation
        policy = am.getDefaultEscalationPolicy()
    }

    go am.runEscalation(ctx, alert, policy)
    return nil
}

func (am *AlertManager) runEscalation(ctx context.Context, alert *Alert, policy *EscalationPolicy) {
    for _, step := range policy.Steps {
        // Wait for step duration
        select {
        case <-ctx.Done():
            return
        case <-time.After(step.WaitDuration):
        }

        // Check if alert is still active
        am.mutex.RLock()
        currentAlert, exists := am.alerts[alert.ID]
        am.mutex.RUnlock()

        if !exists || currentAlert.Status != AlertStatusFiring {
            return // Alert resolved or acknowledged
        }

        // Send notifications for this step
        for _, channel := range step.NotificationChannels {
            if err := am.sendEscalationNotification(ctx, currentAlert, channel, step.StepNumber); err != nil {
                am.logger.Error("Failed to send escalation notification",
                    Field{Key: "alert_id", Value: alert.ID},
                    Field{Key: "channel", Value: channel},
                    Field{Key: "step", Value: step.StepNumber},
                    Error(err))
            }
        }

        // Record escalation level
        currentAlert.EscalationLevel = step.StepNumber
        currentAlert.UpdatedAt = time.Now()
    }
}

// Alert Correlation and Deduplication
type CorrelationRule struct {
    ID              string            `json:"id"`
    Name            string            `json:"name"`
    GroupByLabels   []string          `json:"groupByLabels"`
    TimeWindow      time.Duration     `json:"timeWindow"`
    MaxAlerts       int               `json:"maxAlerts"`
    SuppressMode    SuppressionMode   `json:"suppressMode"`
}

type SuppressionMode string

const (
    SuppressionModeDeduplication SuppressionMode = "deduplication"
    SuppressionModeGrouping      SuppressionMode = "grouping"
    SuppressionModeRateLimiting  SuppressionMode = "rate_limiting"
)

func (am *AlertManager) correlateAlerts(alert *Alert) {
    am.mutex.RLock()
    correlationRules := am.correlations
    am.mutex.RUnlock()

    for _, rule := range correlationRules {
        if am.matchesCorrelationRule(alert, rule) {
            am.applyCorrelation(alert, rule)
        }
    }
}

// Suppression and Maintenance Windows
type SuppressionRule struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Type        SuppressionType   `json:"type"`
    Filters     map[string]string `json:"filters"`
    StartTime   time.Time         `json:"startTime"`
    EndTime     time.Time         `json:"endTime"`
    Recurring   *RecurrenceRule   `json:"recurring,omitempty"`
    Reason      string            `json:"reason"`
    CreatedBy   string            `json:"createdBy"`
    CreatedAt   time.Time         `json:"createdAt"`
}

type SuppressionType string

const (
    SuppressionMaintenance SuppressionType = "maintenance"
    SuppressionTesting     SuppressionType = "testing"
    SuppressionManual      SuppressionType = "manual"
)

type RecurrenceRule struct {
    Pattern   string `json:"pattern"`   // "daily", "weekly", "monthly"
    Interval  int    `json:"interval"`  // Every N periods
    DaysOfWeek []int `json:"daysOfWeek,omitempty"` // For weekly recurrence
}

func (am *AlertManager) isAlertSuppressed(alert *Alert) bool {
    am.mutex.RLock()
    suppressions := am.suppressions
    am.mutex.RUnlock()

    now := time.Now()

    for _, suppression := range suppressions {
        if am.isSuppressionActive(suppression, now) && am.alertMatchesSuppression(alert, suppression) {
            am.logger.Debug("Alert suppressed by rule",
                Field{Key: "alert_id", Value: alert.ID},
                Field{Key: "suppression_rule", Value: suppression.ID})
            return true
        }
    }

    return false
}

func (am *AlertManager) isSuppressionActive(suppression *SuppressionRule, now time.Time) bool {
    if suppression.Recurring != nil {
        return am.isRecurringSuppressionActive(suppression, now)
    }

    return now.After(suppression.StartTime) && now.Before(suppression.EndTime)
}

// API Methods for Alert Management
func (am *AlertManager) CreateAlertRule(rule *AlertRule) error {
    rule.ID = generateRuleID()
    rule.CreatedAt = time.Now()
    rule.UpdatedAt = time.Now()

    am.mutex.Lock()
    am.rules[rule.ID] = rule
    am.mutex.Unlock()

    return am.storage.SaveAlertRule(rule)
}

func (am *AlertManager) UpdateAlertRule(ruleID string, updates *AlertRule) error {
    am.mutex.Lock()
    rule, exists := am.rules[ruleID]
    if !exists {
        am.mutex.Unlock()
        return fmt.Errorf("alert rule not found: %s", ruleID)
    }

    updates.ID = ruleID
    updates.CreatedAt = rule.CreatedAt
    updates.UpdatedAt = time.Now()
    am.rules[ruleID] = updates
    am.mutex.Unlock()

    return am.storage.SaveAlertRule(updates)
}

func (am *AlertManager) AcknowledgeAlert(alertID, acknowledgedBy string) error {
    am.mutex.Lock()
    alert, exists := am.alerts[alertID]
    if !exists {
        am.mutex.Unlock()
        return fmt.Errorf("alert not found: %s", alertID)
    }

    now := time.Now()
    alert.Status = AlertStatusAcknowledged
    alert.AcknowledgedAt = &now
    alert.AcknowledgedBy = acknowledgedBy
    alert.UpdatedAt = now
    am.mutex.Unlock()

    am.logger.Info("Alert acknowledged",
        Field{Key: "alert_id", Value: alertID},
        Field{Key: "acknowledged_by", Value: acknowledgedBy})

    return am.storage.SaveAlert(alert)
}

func (am *AlertManager) ResolveAlert(alertID, resolvedBy string) error {
    return am.resolveAlert(context.Background(), alertID, resolvedBy)
}

func (am *AlertManager) resolveAlert(ctx context.Context, alertID, resolvedBy string) error {
    am.mutex.Lock()
    alert, exists := am.alerts[alertID]
    if !exists {
        am.mutex.Unlock()
        return fmt.Errorf("alert not found: %s", alertID)
    }

    now := time.Now()
    alert.Status = AlertStatusResolved
    alert.ResolvedAt = &now
    alert.ResolvedBy = resolvedBy
    alert.EndsAt = &now
    alert.UpdatedAt = now
    am.mutex.Unlock()

    am.logger.Info("Alert resolved",
        Field{Key: "alert_id", Value: alertID},
        Field{Key: "resolved_by", Value: resolvedBy})

    return am.storage.SaveAlert(alert)
}

func (am *AlertManager) GetActiveAlerts() ([]*Alert, error) {
    am.mutex.RLock()
    defer am.mutex.RUnlock()

    var activeAlerts []*Alert
    for _, alert := range am.alerts {
        if alert.Status == AlertStatusFiring || alert.Status == AlertStatusAcknowledged {
            activeAlerts = append(activeAlerts, alert)
        }
    }

    return activeAlerts, nil
}

func (am *AlertManager) GetAlertHistory(filters *AlertFilters) ([]*Alert, error) {
    return am.storage.GetAlerts(filters)
}
```

## Application Integration
```go
// App integration with alert management
func (a *App) initializeAlertManagement() error {
    a.alertManager = NewAlertManager(a.notificationManager, a.alertStorage, a.logger)

    // Load existing rules and policies
    if err := a.loadAlertConfiguration(); err != nil {
        return fmt.Errorf("failed to load alert configuration: %w", err)
    }

    // Start alert evaluation loop
    go a.runAlertEvaluationLoop()

    return nil
}

func (a *App) runAlertEvaluationLoop() {
    ticker := time.NewTicker(30 * time.Second) // Evaluate every 30 seconds
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

            data, err := a.getMonitoringData()
            if err != nil {
                a.logger.Error("Failed to get monitoring data for alert evaluation", Error(err))
                cancel()
                continue
            }

            if err := a.alertManager.EvaluateRules(ctx, data); err != nil {
                a.logger.Error("Failed to evaluate alert rules", Error(err))
            }

            cancel()
        }
    }
}

// API methods exposed to frontend
func (a *App) CreateAlertRule(rule *AlertRule) error {
    return a.alertManager.CreateAlertRule(rule)
}

func (a *App) GetAlertRules() ([]*AlertRule, error) {
    return a.alertManager.GetAlertRules()
}

func (a *App) GetActiveAlerts() ([]*Alert, error) {
    return a.alertManager.GetActiveAlerts()
}

func (a *App) AcknowledgeAlert(alertID string) error {
    return a.alertManager.AcknowledgeAlert(alertID, "user")
}

func (a *App) ResolveAlert(alertID string) error {
    return a.alertManager.ResolveAlert(alertID, "user")
}

func (a *App) CreateSuppressionRule(rule *SuppressionRule) error {
    return a.alertManager.CreateSuppressionRule(rule)
}
```

## Verification Steps
1. Test alert rule creation and evaluation - should create and evaluate rules correctly
2. Verify escalation policies - should escalate alerts according to configured policies
3. Test alert correlation - should group related alerts appropriately
4. Verify suppression rules - should suppress alerts during maintenance windows
5. Test alert lifecycle - should handle acknowledgment and resolution correctly
6. Verify alert history - should track and retrieve alert history
7. Test integration with notifications - should send notifications through configured channels
8. Verify performance impact - should not significantly impact monitoring performance

## Dependencies
- T041: System Notifications
- T039: Comprehensive Logging System
- T015: Monitoring Status API
- T025: Storage API Integration

## Notes
- Implement comprehensive testing for alert rule evaluation
- Consider implementing alert templates for common scenarios
- Plan for integration with external alerting systems (PagerDuty, OpsGenie)
- Optimize alert evaluation performance for large numbers of rules
- Consider implementing machine learning for anomaly detection
- Plan for alert analytics and reporting features