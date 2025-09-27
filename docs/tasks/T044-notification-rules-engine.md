# T044: Notification Rules Engine

## Overview
Implement a flexible rules engine for notification routing and filtering, allowing users to define complex conditions for when, how, and to whom notifications should be sent.

## Context
Different users and teams need different notification preferences based on time, severity, endpoint importance, and personal schedules. A rules engine provides the flexibility to handle complex notification scenarios without hardcoding logic.

## Task Description
Create a comprehensive notification rules engine with conditional logic, multiple actions, time-based rules, escalation chains, and integration with all notification channels (system, email, external).

## Acceptance Criteria
- [ ] Rule-based notification routing with conditional logic
- [ ] Time-based notification rules (business hours, weekends, holidays)
- [ ] User and team-based notification preferences
- [ ] Escalation chains with multiple notification channels
- [ ] Rule testing and simulation capabilities
- [ ] Rule templates and presets for common scenarios
- [ ] Performance optimization for large rule sets
- [ ] Rule conflict resolution and priority handling
- [ ] Integration with all notification channels

## Rules Engine Architecture
```go
package rules

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type NotificationRulesEngine struct {
    rules           map[string]*NotificationRule
    ruleGroups      map[string]*RuleGroup
    evaluator       *RuleEvaluator
    actionExecutor  *ActionExecutor
    scheduler       *RuleScheduler
    cache           *RuleCache
    metrics         *RuleMetrics
    mutex           sync.RWMutex
    logger          Logger
}

type NotificationRule struct {
    ID              string                `json:"id"`
    Name            string                `json:"name"`
    Description     string                `json:"description"`
    Enabled         bool                  `json:"enabled"`
    Priority        int                   `json:"priority"` // Higher number = higher priority
    Conditions      []RuleCondition       `json:"conditions"`
    Actions         []RuleAction          `json:"actions"`
    TimeConstraints *TimeConstraints      `json:"timeConstraints,omitempty"`
    UserConstraints *UserConstraints      `json:"userConstraints,omitempty"`
    Cooldown        *CooldownConfig       `json:"cooldown,omitempty"`
    Tags            []string              `json:"tags,omitempty"`
    CreatedAt       time.Time             `json:"createdAt"`
    UpdatedAt       time.Time             `json:"updatedAt"`
    CreatedBy       string                `json:"createdBy"`
    LastTriggered   *time.Time            `json:"lastTriggered,omitempty"`
    TriggerCount    int64                 `json:"triggerCount"`
}

type RuleCondition struct {
    Type        ConditionType         `json:"type"`
    Field       string                `json:"field"`
    Operator    ComparisonOperator    `json:"operator"`
    Value       interface{}           `json:"value"`
    Values      []interface{}         `json:"values,omitempty"` // For IN/NOT_IN operators
    CaseSensitive bool                `json:"caseSensitive"`
    Regex       string                `json:"regex,omitempty"`
    Function    string                `json:"function,omitempty"` // Custom function name
}

type ConditionType string

const (
    ConditionAlert      ConditionType = "alert"
    ConditionEndpoint   ConditionType = "endpoint"
    ConditionTime       ConditionType = "time"
    ConditionUser       ConditionType = "user"
    ConditionFrequency  ConditionType = "frequency"
    ConditionCustom     ConditionType = "custom"
)

type ComparisonOperator string

const (
    OpEquals        ComparisonOperator = "equals"
    OpNotEquals     ComparisonOperator = "not_equals"
    OpContains      ComparisonOperator = "contains"
    OpNotContains   ComparisonOperator = "not_contains"
    OpStartsWith    ComparisonOperator = "starts_with"
    OpEndsWith      ComparisonOperator = "ends_with"
    OpGreaterThan   ComparisonOperator = "greater_than"
    OpLessThan      ComparisonOperator = "less_than"
    OpGreaterEqual  ComparisonOperator = "greater_equal"
    OpLessEqual     ComparisonOperator = "less_equal"
    OpIn            ComparisonOperator = "in"
    OpNotIn         ComparisonOperator = "not_in"
    OpMatches       ComparisonOperator = "matches" // Regex
    OpExists        ComparisonOperator = "exists"
    OpNotExists     ComparisonOperator = "not_exists"
)

type RuleAction struct {
    Type        ActionType            `json:"type"`
    Channel     NotificationChannel   `json:"channel"`
    Recipients  []string              `json:"recipients"`
    Template    string                `json:"template,omitempty"`
    Priority    ActionPriority        `json:"priority"`
    Delay       time.Duration         `json:"delay,omitempty"`
    Retries     int                   `json:"retries"`
    Condition   *ActionCondition      `json:"condition,omitempty"` // Additional condition for this action
    Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type ActionType string

const (
    ActionNotify       ActionType = "notify"
    ActionEscalate     ActionType = "escalate"
    ActionSuppress     ActionType = "suppress"
    ActionDelay        ActionType = "delay"
    ActionWebhook      ActionType = "webhook"
    ActionScript       ActionType = "script"
    ActionCreateTicket ActionType = "create_ticket"
)

type NotificationChannel string

const (
    ChannelSystem    NotificationChannel = "system"
    ChannelEmail     NotificationChannel = "email"
    ChannelSlack     NotificationChannel = "slack"
    ChannelTeams     NotificationChannel = "teams"
    ChannelWebhook   NotificationChannel = "webhook"
    ChannelSMS       NotificationChannel = "sms"
    ChannelVoice     NotificationChannel = "voice"
)

type ActionPriority string

const (
    ActionPriorityLow      ActionPriority = "low"
    ActionPriorityNormal   ActionPriority = "normal"
    ActionPriorityHigh     ActionPriority = "high"
    ActionPriorityCritical ActionPriority = "critical"
)

type TimeConstraints struct {
    BusinessHoursOnly bool              `json:"businessHoursOnly"`
    BusinessHours     *BusinessHours    `json:"businessHours,omitempty"`
    Timezone          string            `json:"timezone"`
    ExcludeWeekends   bool              `json:"excludeWeekends"`
    ExcludeHolidays   bool              `json:"excludeHolidays"`
    SpecificDays      []time.Weekday    `json:"specificDays,omitempty"`
    TimeRanges        []TimeRange       `json:"timeRanges,omitempty"`
}

type BusinessHours struct {
    Start string `json:"start"` // "09:00"
    End   string `json:"end"`   // "17:00"
}

type TimeRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type UserConstraints struct {
    IncludeUsers    []string `json:"includeUsers,omitempty"`
    ExcludeUsers    []string `json:"excludeUsers,omitempty"`
    RequireRole     string   `json:"requireRole,omitempty"`
    UserProperties  map[string]string `json:"userProperties,omitempty"`
}

type CooldownConfig struct {
    Duration    time.Duration `json:"duration"`
    Scope       CooldownScope `json:"scope"`
    CountLimit  int           `json:"countLimit,omitempty"` // Max notifications per duration
}

type CooldownScope string

const (
    CooldownPerRule     CooldownScope = "per_rule"
    CooldownPerEndpoint CooldownScope = "per_endpoint"
    CooldownPerUser     CooldownScope = "per_user"
    CooldownGlobal      CooldownScope = "global"
)

func NewNotificationRulesEngine(logger Logger) *NotificationRulesEngine {
    return &NotificationRulesEngine{
        rules:          make(map[string]*NotificationRule),
        ruleGroups:     make(map[string]*RuleGroup),
        evaluator:      NewRuleEvaluator(),
        actionExecutor: NewActionExecutor(),
        scheduler:      NewRuleScheduler(),
        cache:          NewRuleCache(),
        metrics:        NewRuleMetrics(),
        logger:         logger,
    }
}

func (nre *NotificationRulesEngine) ProcessEvent(ctx context.Context, event *NotificationEvent) error {
    // Get applicable rules based on event type and filters
    applicableRules := nre.getApplicableRules(event)

    // Sort rules by priority (highest first)
    sort.Slice(applicableRules, func(i, j int) bool {
        return applicableRules[i].Priority > applicableRules[j].Priority
    })

    var executedActions []string
    suppressRemaining := false

    for _, rule := range applicableRules {
        if suppressRemaining {
            nre.logger.Debug("Skipping rule due to suppression",
                Field{Key: "rule_id", Value: rule.ID})
            continue
        }

        // Check cooldown
        if nre.isRuleInCooldown(rule, event) {
            nre.logger.Debug("Rule in cooldown, skipping",
                Field{Key: "rule_id", Value: rule.ID})
            continue
        }

        // Evaluate rule conditions
        matches, err := nre.evaluator.EvaluateRule(rule, event)
        if err != nil {
            nre.logger.Error("Failed to evaluate rule",
                Field{Key: "rule_id", Value: rule.ID},
                Field{Key: "error", Value: err.Error()})
            continue
        }

        if !matches {
            continue
        }

        nre.logger.Info("Rule matched, executing actions",
            Field{Key: "rule_id", Value: rule.ID},
            Field{Key: "event_type", Value: event.Type})

        // Execute actions
        for _, action := range rule.Actions {
            actionID := fmt.Sprintf("%s_%s", rule.ID, action.Type)

            // Check if this action type has already been executed
            if contains(executedActions, string(action.Type)) && nre.shouldDeduplicateAction(action) {
                nre.logger.Debug("Skipping duplicate action",
                    Field{Key: "action_type", Value: string(action.Type)})
                continue
            }

            if err := nre.actionExecutor.ExecuteAction(ctx, action, event, rule); err != nil {
                nre.logger.Error("Failed to execute action",
                    Field{Key: "rule_id", Value: rule.ID},
                    Field{Key: "action_type", Value: string(action.Type)},
                    Field{Key: "error", Value: err.Error()})
                continue
            }

            executedActions = append(executedActions, string(action.Type))

            // Check if this action suppresses remaining rules
            if action.Type == ActionSuppress {
                suppressRemaining = true
                nre.logger.Debug("Suppression action executed, skipping remaining rules")
            }
        }

        // Update rule metrics
        nre.updateRuleMetrics(rule, event)

        // Update cooldown
        nre.updateRuleCooldown(rule, event)
    }

    return nil
}

func (nre *NotificationRulesEngine) getApplicableRules(event *NotificationEvent) []*NotificationRule {
    nre.mutex.RLock()
    defer nre.mutex.RUnlock()

    var applicable []*NotificationRule

    for _, rule := range nre.rules {
        if !rule.Enabled {
            continue
        }

        // Check time constraints
        if rule.TimeConstraints != nil && !nre.checkTimeConstraints(rule.TimeConstraints, time.Now()) {
            continue
        }

        // Check user constraints
        if rule.UserConstraints != nil && !nre.checkUserConstraints(rule.UserConstraints, event.User) {
            continue
        }

        // Quick filter based on event type and basic conditions
        if nre.quickFilterRule(rule, event) {
            applicable = append(applicable, rule)
        }
    }

    return applicable
}

// Rule Evaluator
type RuleEvaluator struct {
    functions map[string]EvaluationFunction
}

type EvaluationFunction func(event *NotificationEvent, condition *RuleCondition) (bool, error)

func NewRuleEvaluator() *RuleEvaluator {
    evaluator := &RuleEvaluator{
        functions: make(map[string]EvaluationFunction),
    }

    // Register built-in functions
    evaluator.RegisterFunction("endpoint_down_duration", evaluator.endpointDownDuration)
    evaluator.RegisterFunction("consecutive_failures", evaluator.consecutiveFailures)
    evaluator.RegisterFunction("business_hours", evaluator.businessHours)
    evaluator.RegisterFunction("user_on_call", evaluator.userOnCall)

    return evaluator
}

func (re *RuleEvaluator) EvaluateRule(rule *NotificationRule, event *NotificationEvent) (bool, error) {
    if len(rule.Conditions) == 0 {
        return true, nil // No conditions means always match
    }

    // Evaluate all conditions (AND logic by default)
    for _, condition := range rule.Conditions {
        matches, err := re.evaluateCondition(&condition, event)
        if err != nil {
            return false, fmt.Errorf("failed to evaluate condition: %w", err)
        }

        if !matches {
            return false, nil
        }
    }

    return true, nil
}

func (re *RuleEvaluator) evaluateCondition(condition *RuleCondition, event *NotificationEvent) (bool, error) {
    // Handle custom functions
    if condition.Function != "" {
        if fn, exists := re.functions[condition.Function]; exists {
            return fn(event, condition)
        }
        return false, fmt.Errorf("unknown function: %s", condition.Function)
    }

    // Get the value to compare
    actualValue, err := re.extractValue(condition.Field, event)
    if err != nil {
        return false, err
    }

    // Handle existence checks
    if condition.Operator == OpExists {
        return actualValue != nil, nil
    }
    if condition.Operator == OpNotExists {
        return actualValue == nil, nil
    }

    if actualValue == nil {
        return false, nil
    }

    // Perform comparison
    return re.compareValues(actualValue, condition.Value, condition.Operator, condition.CaseSensitive)
}

func (re *RuleEvaluator) extractValue(field string, event *NotificationEvent) (interface{}, error) {
    parts := strings.Split(field, ".")

    switch parts[0] {
    case "alert":
        return re.extractAlertValue(parts[1:], event.Alert)
    case "endpoint":
        return re.extractEndpointValue(parts[1:], event.Endpoint)
    case "event":
        return re.extractEventValue(parts[1:], event)
    case "time":
        return time.Now(), nil
    default:
        return nil, fmt.Errorf("unknown field prefix: %s", parts[0])
    }
}

// Action Executor
type ActionExecutor struct {
    notificationManager NotificationManager
    emailManager        EmailManager
    webhookClient       WebhookClient
    logger              Logger
}

func NewActionExecutor() *ActionExecutor {
    return &ActionExecutor{
        logger: logger,
    }
}

func (ae *ActionExecutor) ExecuteAction(ctx context.Context, action *RuleAction, event *NotificationEvent, rule *NotificationRule) error {
    // Apply delay if specified
    if action.Delay > 0 {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(action.Delay):
        }
    }

    // Check additional action condition
    if action.Condition != nil {
        matches, err := ae.evaluateActionCondition(action.Condition, event)
        if err != nil {
            return fmt.Errorf("failed to evaluate action condition: %w", err)
        }
        if !matches {
            ae.logger.Debug("Action condition not met, skipping action")
            return nil
        }
    }

    switch action.Type {
    case ActionNotify:
        return ae.executeNotifyAction(ctx, action, event, rule)
    case ActionEscalate:
        return ae.executeEscalateAction(ctx, action, event, rule)
    case ActionWebhook:
        return ae.executeWebhookAction(ctx, action, event, rule)
    case ActionSuppress:
        return ae.executeSuppressAction(ctx, action, event, rule)
    default:
        return fmt.Errorf("unknown action type: %s", action.Type)
    }
}

func (ae *ActionExecutor) executeNotifyAction(ctx context.Context, action *RuleAction, event *NotificationEvent, rule *NotificationRule) error {
    switch action.Channel {
    case ChannelSystem:
        return ae.sendSystemNotification(ctx, action, event, rule)
    case ChannelEmail:
        return ae.sendEmailNotification(ctx, action, event, rule)
    case ChannelWebhook:
        return ae.sendWebhookNotification(ctx, action, event, rule)
    default:
        return fmt.Errorf("unsupported notification channel: %s", action.Channel)
    }
}

// Rule Templates and Presets
type RuleTemplate struct {
    ID          string                `json:"id"`
    Name        string                `json:"name"`
    Description string                `json:"description"`
    Category    string                `json:"category"`
    Template    *NotificationRule     `json:"template"`
    Variables   []TemplateVariable    `json:"variables"`
}

type TemplateVariable struct {
    Name         string      `json:"name"`
    Type         string      `json:"type"`
    Description  string      `json:"description"`
    DefaultValue interface{} `json:"defaultValue"`
    Required     bool        `json:"required"`
    Options      []string    `json:"options,omitempty"`
}

func (nre *NotificationRulesEngine) GetRuleTemplates() ([]*RuleTemplate, error) {
    templates := []*RuleTemplate{
        {
            ID:          "endpoint_down_email",
            Name:        "Endpoint Down - Email Alert",
            Description: "Send email notification when an endpoint goes down",
            Category:    "basic",
            Template: &NotificationRule{
                Name:        "{{.endpoint_name}} Down Alert",
                Description: "Email notification for endpoint failures",
                Enabled:     true,
                Priority:    100,
                Conditions: []RuleCondition{
                    {
                        Type:     ConditionAlert,
                        Field:    "alert.status",
                        Operator: OpEquals,
                        Value:    "firing",
                    },
                    {
                        Type:     ConditionAlert,
                        Field:    "alert.severity",
                        Operator: OpIn,
                        Values:   []interface{}{"critical", "emergency"},
                    },
                },
                Actions: []RuleAction{
                    {
                        Type:       ActionNotify,
                        Channel:    ChannelEmail,
                        Recipients: []string{"{{.email_recipients}}"},
                        Template:   "endpoint_failure",
                        Priority:   ActionPriorityHigh,
                    },
                },
            },
            Variables: []TemplateVariable{
                {
                    Name:        "endpoint_name",
                    Type:        "string",
                    Description: "Name of the endpoint to monitor",
                    Required:    true,
                },
                {
                    Name:        "email_recipients",
                    Type:        "email_list",
                    Description: "Comma-separated list of email addresses",
                    Required:    true,
                },
            },
        },
        {
            ID:          "business_hours_only",
            Name:        "Business Hours Only Notifications",
            Description: "Only send notifications during business hours",
            Category:    "time_based",
            Template: &NotificationRule{
                Name:        "Business Hours Notifications",
                Description: "Notifications only during business hours (9 AM - 5 PM, weekdays)",
                Enabled:     true,
                Priority:    50,
                TimeConstraints: &TimeConstraints{
                    BusinessHoursOnly: true,
                    BusinessHours: &BusinessHours{
                        Start: "09:00",
                        End:   "17:00",
                    },
                    ExcludeWeekends: true,
                    Timezone:        "America/New_York",
                },
                Conditions: []RuleCondition{
                    {
                        Type:     ConditionAlert,
                        Field:    "alert.severity",
                        Operator: OpIn,
                        Values:   []interface{}{"warning", "critical"},
                    },
                },
                Actions: []RuleAction{
                    {
                        Type:       ActionNotify,
                        Channel:    ChannelSystem,
                        Recipients: []string{"{{.recipients}}"},
                        Priority:   ActionPriorityNormal,
                    },
                },
            },
            Variables: []TemplateVariable{
                {
                    Name:        "recipients",
                    Type:        "user_list",
                    Description: "List of users to notify",
                    Required:    true,
                },
            },
        },
    }

    return templates, nil
}

func (nre *NotificationRulesEngine) CreateRuleFromTemplate(templateID string, variables map[string]interface{}) (*NotificationRule, error) {
    templates, err := nre.GetRuleTemplates()
    if err != nil {
        return nil, err
    }

    var template *RuleTemplate
    for _, t := range templates {
        if t.ID == templateID {
            template = t
            break
        }
    }

    if template == nil {
        return nil, fmt.Errorf("template not found: %s", templateID)
    }

    // Validate required variables
    for _, variable := range template.Variables {
        if variable.Required {
            if _, exists := variables[variable.Name]; !exists {
                return nil, fmt.Errorf("required variable missing: %s", variable.Name)
            }
        }
    }

    // Create rule from template
    rule := *template.Template // Copy template
    rule.ID = generateRuleID()
    rule.CreatedAt = time.Now()
    rule.UpdatedAt = time.Now()

    // Replace template variables
    if err := nre.interpolateRuleTemplate(&rule, variables); err != nil {
        return nil, fmt.Errorf("failed to interpolate template: %w", err)
    }

    return &rule, nil
}

// API Methods
func (nre *NotificationRulesEngine) CreateRule(rule *NotificationRule) error {
    rule.ID = generateRuleID()
    rule.CreatedAt = time.Now()
    rule.UpdatedAt = time.Now()

    if err := nre.validateRule(rule); err != nil {
        return fmt.Errorf("rule validation failed: %w", err)
    }

    nre.mutex.Lock()
    nre.rules[rule.ID] = rule
    nre.mutex.Unlock()

    nre.logger.Info("Notification rule created",
        Field{Key: "rule_id", Value: rule.ID},
        Field{Key: "rule_name", Value: rule.Name})

    return nil
}

func (nre *NotificationRulesEngine) TestRule(ruleID string, testEvent *NotificationEvent) (*RuleTestResult, error) {
    nre.mutex.RLock()
    rule, exists := nre.rules[ruleID]
    nre.mutex.RUnlock()

    if !exists {
        return nil, fmt.Errorf("rule not found: %s", ruleID)
    }

    result := &RuleTestResult{
        RuleID:    ruleID,
        Matched:   false,
        Actions:   []string{},
        Errors:    []string{},
        Timestamp: time.Now(),
    }

    // Test rule evaluation
    matches, err := nre.evaluator.EvaluateRule(rule, testEvent)
    if err != nil {
        result.Errors = append(result.Errors, fmt.Sprintf("Evaluation error: %v", err))
        return result, nil
    }

    result.Matched = matches

    if matches {
        // Test actions (dry run)
        for _, action := range rule.Actions {
            actionResult := fmt.Sprintf("Would execute %s action on %s channel", action.Type, action.Channel)
            result.Actions = append(result.Actions, actionResult)
        }
    }

    return result, nil
}

type RuleTestResult struct {
    RuleID    string    `json:"ruleId"`
    Matched   bool      `json:"matched"`
    Actions   []string  `json:"actions"`
    Errors    []string  `json:"errors"`
    Timestamp time.Time `json:"timestamp"`
}
```

## Application Integration
```go
// App integration with notification rules engine
func (a *App) initializeNotificationRules() error {
    a.rulesEngine = NewNotificationRulesEngine(a.logger)

    // Load existing rules
    if err := a.loadNotificationRules(); err != nil {
        return fmt.Errorf("failed to load notification rules: %w", err)
    }

    return nil
}

// API methods
func (a *App) CreateNotificationRule(rule *NotificationRule) error {
    return a.rulesEngine.CreateRule(rule)
}

func (a *App) GetNotificationRules() ([]*NotificationRule, error) {
    return a.rulesEngine.GetRules()
}

func (a *App) TestNotificationRule(ruleID string, testData map[string]interface{}) (*RuleTestResult, error) {
    testEvent := &NotificationEvent{
        Type:      "test",
        Timestamp: time.Now(),
        Data:      testData,
    }
    return a.rulesEngine.TestRule(ruleID, testEvent)
}

func (a *App) GetRuleTemplates() ([]*RuleTemplate, error) {
    return a.rulesEngine.GetRuleTemplates()
}

func (a *App) CreateRuleFromTemplate(templateID string, variables map[string]interface{}) (*NotificationRule, error) {
    return a.rulesEngine.CreateRuleFromTemplate(templateID, variables)
}
```

## Verification Steps
1. Test rule creation and validation - should create and validate rules correctly
2. Verify condition evaluation - should evaluate complex conditions accurately
3. Test action execution - should execute actions through correct channels
4. Verify time constraints - should respect business hours and time zones
5. Test rule templates - should create rules from templates with variable substitution
6. Verify cooldown functionality - should prevent notification spam
7. Test rule testing/simulation - should provide accurate test results
8. Verify integration with notification channels - should route to correct destinations

## Dependencies
- T041: System Notifications
- T042: Alert Management System
- T043: Email Notifications
- T039: Comprehensive Logging System

## Notes
- Design rules engine for extensibility and performance
- Consider implementing rule versioning and rollback
- Plan for rule analytics and optimization recommendations
- Implement comprehensive rule validation and testing
- Consider integration with external rule engines if needed
- Plan for future advanced features like machine learning-based routing