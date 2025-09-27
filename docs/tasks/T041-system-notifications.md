# T041: System Notifications

## Overview
Implement cross-platform system notifications for alerting users about network monitoring events, failures, and threshold breaches with customizable notification preferences.

## Context
NetMonitor needs to alert users about important events like endpoint failures, threshold breaches, and system issues through native system notifications. These should be non-intrusive but informative, with user control over notification preferences.

## Task Description
Create a comprehensive system notification framework with cross-platform native notifications, customizable triggers, notification management, and integration with the monitoring system.

## Acceptance Criteria
- [ ] Cross-platform native system notifications (Windows, macOS, Linux)
- [ ] Configurable notification triggers and thresholds
- [ ] Notification templates and customization
- [ ] Notification rate limiting and grouping
- [ ] Action buttons in notifications (where supported)
- [ ] Notification history and management
- [ ] Do Not Disturb mode integration
- [ ] Sound and visual customization options
- [ ] Notification persistence and retry logic

## Notification System Architecture
```go
package notifications

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type NotificationManager struct {
    provider    NotificationProvider
    config      *NotificationConfig
    rateLimit   *RateLimiter
    history     *NotificationHistory
    templates   map[string]*NotificationTemplate
    mu          sync.RWMutex
    logger      Logger
}

type NotificationProvider interface {
    ShowNotification(notification *Notification) error
    SupportsActions() bool
    SupportsCustomSounds() bool
    SupportsCustomIcons() bool
    GetCapabilities() *ProviderCapabilities
}

type Notification struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Message     string                 `json:"message"`
    Icon        string                 `json:"icon"`
    Sound       string                 `json:"sound"`
    Priority    Priority               `json:"priority"`
    Category    Category               `json:"category"`
    Actions     []NotificationAction   `json:"actions"`
    Data        map[string]interface{} `json:"data"`
    Timestamp   time.Time              `json:"timestamp"`
    ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
    Persistent  bool                   `json:"persistent"`
}

type NotificationAction struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    Icon  string `json:"icon,omitempty"`
}

type Priority int

const (
    PriorityLow Priority = iota
    PriorityNormal
    PriorityHigh
    PriorityCritical
)

type Category string

const (
    CategoryEndpointFailure Category = "endpoint_failure"
    CategoryThresholdBreach Category = "threshold_breach"
    CategorySystemAlert     Category = "system_alert"
    CategoryTestComplete    Category = "test_complete"
    CategoryConfigChange    Category = "config_change"
)

type NotificationConfig struct {
    Enabled             bool                      `json:"enabled"`
    Categories          map[Category]*CategoryConfig `json:"categories"`
    RateLimit           *RateLimitConfig          `json:"rateLimit"`
    DoNotDisturbEnabled bool                      `json:"doNotDisturbEnabled"`
    DoNotDisturbStart   string                    `json:"doNotDisturbStart"` // "22:00"
    DoNotDisturbEnd     string                    `json:"doNotDisturbEnd"`   // "07:00"
    DefaultSound        string                    `json:"defaultSound"`
    DefaultIcon         string                    `json:"defaultIcon"`
    ShowInTray          bool                      `json:"showInTray"`
}

type CategoryConfig struct {
    Enabled         bool     `json:"enabled"`
    MinPriority     Priority `json:"minPriority"`
    Sound           string   `json:"sound"`
    ShowActions     bool     `json:"showActions"`
    Persistent      bool     `json:"persistent"`
    CooldownMinutes int      `json:"cooldownMinutes"`
}

type ProviderCapabilities struct {
    SupportsActions      bool `json:"supportsActions"`
    SupportsCustomSounds bool `json:"supportsCustomSounds"`
    SupportsCustomIcons  bool `json:"supportsCustomIcons"`
    SupportsPersistence  bool `json:"supportsPersistence"`
    MaxTitleLength       int  `json:"maxTitleLength"`
    MaxMessageLength     int  `json:"maxMessageLength"`
}

func NewNotificationManager(provider NotificationProvider, config *NotificationConfig, logger Logger) *NotificationManager {
    return &NotificationManager{
        provider:  provider,
        config:    config,
        rateLimit: NewRateLimiter(config.RateLimit),
        history:   NewNotificationHistory(),
        templates: make(map[string]*NotificationTemplate),
        logger:    logger,
    }
}

func (nm *NotificationManager) ShowNotification(ctx context.Context, notification *Notification) error {
    if !nm.config.Enabled {
        return nil // Notifications disabled
    }

    // Check Do Not Disturb mode
    if nm.isDoNotDisturbActive() {
        nm.logger.Debug("Notification suppressed due to Do Not Disturb mode",
            Field{Key: "notification_id", Value: notification.ID})
        return nil
    }

    // Check category configuration
    categoryConfig := nm.config.Categories[notification.Category]
    if categoryConfig == nil || !categoryConfig.Enabled {
        return nil // Category disabled
    }

    // Check minimum priority
    if notification.Priority < categoryConfig.MinPriority {
        return nil // Priority too low
    }

    // Check rate limiting
    if !nm.rateLimit.Allow(notification.Category) {
        nm.logger.Debug("Notification rate limited",
            Field{Key: "category", Value: string(notification.Category)})
        return nil
    }

    // Apply category defaults
    nm.applyDefaults(notification, categoryConfig)

    // Show notification
    if err := nm.provider.ShowNotification(notification); err != nil {
        nm.logger.Error("Failed to show notification",
            Field{Key: "notification_id", Value: notification.ID},
            Field{Key: "error", Value: err.Error()})
        return fmt.Errorf("failed to show notification: %w", err)
    }

    // Record in history
    nm.history.Add(notification)

    nm.logger.Info("Notification shown",
        Field{Key: "notification_id", Value: notification.ID},
        Field{Key: "category", Value: string(notification.Category)},
        Field{Key: "priority", Value: int(notification.Priority)})

    return nil
}

func (nm *NotificationManager) isDoNotDisturbActive() bool {
    if !nm.config.DoNotDisturbEnabled {
        return false
    }

    now := time.Now()
    currentTime := now.Format("15:04")

    start := nm.config.DoNotDisturbStart
    end := nm.config.DoNotDisturbEnd

    if start == "" || end == "" {
        return false
    }

    // Handle overnight DND periods (e.g., 22:00 to 07:00)
    if start > end {
        return currentTime >= start || currentTime <= end
    }

    return currentTime >= start && currentTime <= end
}

func (nm *NotificationManager) applyDefaults(notification *Notification, categoryConfig *CategoryConfig) {
    // Apply category-specific defaults
    if notification.Sound == "" && categoryConfig.Sound != "" {
        notification.Sound = categoryConfig.Sound
    }

    if notification.Sound == "" && nm.config.DefaultSound != "" {
        notification.Sound = nm.config.DefaultSound
    }

    if notification.Icon == "" && nm.config.DefaultIcon != "" {
        notification.Icon = nm.config.DefaultIcon
    }

    if categoryConfig.Persistent {
        notification.Persistent = true
    }

    // Remove actions if not supported or disabled
    if !nm.provider.SupportsActions() || !categoryConfig.ShowActions {
        notification.Actions = nil
    }
}

// Notification templates for different events
type NotificationTemplate struct {
    Title    string            `json:"title"`
    Message  string            `json:"message"`
    Icon     string            `json:"icon"`
    Sound    string            `json:"sound"`
    Priority Priority          `json:"priority"`
    Category Category          `json:"category"`
    Actions  []NotificationAction `json:"actions"`
}

func (nm *NotificationManager) RegisterTemplate(name string, template *NotificationTemplate) {
    nm.mu.Lock()
    defer nm.mu.Unlock()
    nm.templates[name] = template
}

func (nm *NotificationManager) CreateFromTemplate(templateName string, data map[string]interface{}) (*Notification, error) {
    nm.mu.RLock()
    template, exists := nm.templates[templateName]
    nm.mu.RUnlock()

    if !exists {
        return nil, fmt.Errorf("template not found: %s", templateName)
    }

    notification := &Notification{
        ID:        generateNotificationID(),
        Title:     nm.interpolateString(template.Title, data),
        Message:   nm.interpolateString(template.Message, data),
        Icon:      template.Icon,
        Sound:     template.Sound,
        Priority:  template.Priority,
        Category:  template.Category,
        Actions:   template.Actions,
        Data:      data,
        Timestamp: time.Now(),
    }

    return notification, nil
}

func (nm *NotificationManager) interpolateString(template string, data map[string]interface{}) string {
    result := template
    for key, value := range data {
        placeholder := fmt.Sprintf("{{.%s}}", key)
        result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
    }
    return result
}

// Rate limiting for notifications
type RateLimiter struct {
    config    *RateLimitConfig
    buckets   map[Category]*TokenBucket
    mu        sync.RWMutex
}

type RateLimitConfig struct {
    MaxPerMinute   int `json:"maxPerMinute"`
    MaxPerHour     int `json:"maxPerHour"`
    BurstSize      int `json:"burstSize"`
    CooldownPeriod int `json:"cooldownPeriod"` // minutes
}

type TokenBucket struct {
    tokens     int
    maxTokens  int
    refillRate time.Duration
    lastRefill time.Time
    mu         sync.Mutex
}

func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
    if config == nil {
        config = &RateLimitConfig{
            MaxPerMinute:   10,
            MaxPerHour:     100,
            BurstSize:      5,
            CooldownPeriod: 5,
        }
    }

    return &RateLimiter{
        config:  config,
        buckets: make(map[Category]*TokenBucket),
    }
}

func (rl *RateLimiter) Allow(category Category) bool {
    rl.mu.Lock()
    bucket, exists := rl.buckets[category]
    if !exists {
        bucket = &TokenBucket{
            tokens:     rl.config.BurstSize,
            maxTokens:  rl.config.BurstSize,
            refillRate: time.Minute / time.Duration(rl.config.MaxPerMinute),
            lastRefill: time.Now(),
        }
        rl.buckets[category] = bucket
    }
    rl.mu.Unlock()

    return bucket.Take()
}

func (tb *TokenBucket) Take() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    now := time.Now()
    elapsed := now.Sub(tb.lastRefill)

    // Refill tokens based on elapsed time
    tokensToAdd := int(elapsed / tb.refillRate)
    if tokensToAdd > 0 {
        tb.tokens = min(tb.maxTokens, tb.tokens+tokensToAdd)
        tb.lastRefill = now
    }

    if tb.tokens > 0 {
        tb.tokens--
        return true
    }

    return false
}
```

## Platform-Specific Implementations

### Windows Notifications
```go
//go:build windows

package notifications

import (
    "fmt"
    "os/exec"
    "syscall"
    "unsafe"

    "golang.org/x/sys/windows"
)

type WindowsNotificationProvider struct {
    appID string
}

func NewWindowsNotificationProvider(appID string) *WindowsNotificationProvider {
    return &WindowsNotificationProvider{
        appID: appID,
    }
}

func (wnp *WindowsNotificationProvider) ShowNotification(notification *Notification) error {
    // Use Windows Toast Notifications
    return wnp.showToastNotification(notification)
}

func (wnp *WindowsNotificationProvider) showToastNotification(notification *Notification) error {
    // PowerShell command to show toast notification
    psScript := fmt.Sprintf(`
        [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null
        [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] > $null

        $template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)

        $textElements = $template.GetElementsByTagName("text")
        $textElements[0].AppendChild($template.CreateTextNode("%s")) > $null
        $textElements[1].AppendChild($template.CreateTextNode("%s")) > $null

        $toast = [Windows.UI.Notifications.ToastNotification]::new($template)
        $notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("%s")
        $notifier.Show($toast)
    `, notification.Title, notification.Message, wnp.appID)

    cmd := exec.Command("powershell", "-Command", psScript)
    cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    return cmd.Run()
}

func (wnp *WindowsNotificationProvider) SupportsActions() bool {
    return true
}

func (wnp *WindowsNotificationProvider) SupportsCustomSounds() bool {
    return true
}

func (wnp *WindowsNotificationProvider) SupportsCustomIcons() bool {
    return true
}

func (wnp *WindowsNotificationProvider) GetCapabilities() *ProviderCapabilities {
    return &ProviderCapabilities{
        SupportsActions:      true,
        SupportsCustomSounds: true,
        SupportsCustomIcons:  true,
        SupportsPersistence:  true,
        MaxTitleLength:       100,
        MaxMessageLength:     200,
    }
}
```

### macOS Notifications
```go
//go:build darwin

package notifications

import (
    "fmt"
    "os/exec"
)

type DarwinNotificationProvider struct {
    bundleID string
}

func NewDarwinNotificationProvider(bundleID string) *DarwinNotificationProvider {
    return &DarwinNotificationProvider{
        bundleID: bundleID,
    }
}

func (dnp *DarwinNotificationProvider) ShowNotification(notification *Notification) error {
    // Use osascript to show notification
    script := fmt.Sprintf(`display notification "%s" with title "%s"`,
        notification.Message, notification.Title)

    if notification.Sound != "" {
        script += fmt.Sprintf(` sound name "%s"`, notification.Sound)
    }

    cmd := exec.Command("osascript", "-e", script)
    return cmd.Run()
}

func (dnp *DarwinNotificationProvider) SupportsActions() bool {
    return false // Basic osascript doesn't support actions
}

func (dnp *DarwinNotificationProvider) SupportsCustomSounds() bool {
    return true
}

func (dnp *DarwinNotificationProvider) SupportsCustomIcons() bool {
    return false // Basic osascript doesn't support custom icons
}

func (dnp *DarwinNotificationProvider) GetCapabilities() *ProviderCapabilities {
    return &ProviderCapabilities{
        SupportsActions:      false,
        SupportsCustomSounds: true,
        SupportsCustomIcons:  false,
        SupportsPersistence:  false,
        MaxTitleLength:       100,
        MaxMessageLength:     200,
    }
}
```

### Linux Notifications
```go
//go:build linux

package notifications

import (
    "fmt"
    "os/exec"
    "strconv"
    "strings"
)

type LinuxNotificationProvider struct {
    appName string
}

func NewLinuxNotificationProvider(appName string) *LinuxNotificationProvider {
    return &LinuxNotificationProvider{
        appName: appName,
    }
}

func (lnp *LinuxNotificationProvider) ShowNotification(notification *Notification) error {
    // Use notify-send (libnotify)
    args := []string{
        "notify-send",
        "--app-name", lnp.appName,
        "--urgency", lnp.mapPriorityToUrgency(notification.Priority),
    }

    if notification.Icon != "" {
        args = append(args, "--icon", notification.Icon)
    }

    if notification.Sound != "" {
        // Note: notify-send doesn't directly support sounds
        // Could use separate sound command
    }

    args = append(args, notification.Title, notification.Message)

    cmd := exec.Command(args[0], args[1:]...)
    return cmd.Run()
}

func (lnp *LinuxNotificationProvider) mapPriorityToUrgency(priority Priority) string {
    switch priority {
    case PriorityLow:
        return "low"
    case PriorityNormal:
        return "normal"
    case PriorityHigh, PriorityCritical:
        return "critical"
    default:
        return "normal"
    }
}

func (lnp *LinuxNotificationProvider) SupportsActions() bool {
    // Check if desktop environment supports actions
    return lnp.hasActionSupport()
}

func (lnp *LinuxNotificationProvider) hasActionSupport() bool {
    // Check desktop environment
    desktop := os.Getenv("XDG_CURRENT_DESKTOP")
    supportedDesktops := []string{"GNOME", "KDE", "XFCE"}

    for _, supported := range supportedDesktops {
        if strings.Contains(strings.ToUpper(desktop), supported) {
            return true
        }
    }

    return false
}

func (lnp *LinuxNotificationProvider) SupportsCustomSounds() bool {
    return false // notify-send doesn't support sounds directly
}

func (lnp *LinuxNotificationProvider) SupportsCustomIcons() bool {
    return true
}

func (lnp *LinuxNotificationProvider) GetCapabilities() *ProviderCapabilities {
    return &ProviderCapabilities{
        SupportsActions:      lnp.hasActionSupport(),
        SupportsCustomSounds: false,
        SupportsCustomIcons:  true,
        SupportsPersistence:  false,
        MaxTitleLength:       100,
        MaxMessageLength:     500,
    }
}
```

## Application Integration
```go
// App integration with notification system
func (a *App) initializeNotifications() error {
    // Create platform-specific provider
    var provider NotificationProvider

    switch runtime.GOOS {
    case "windows":
        provider = NewWindowsNotificationProvider("NetMonitor")
    case "darwin":
        provider = NewDarwinNotificationProvider("com.netmonitor.app")
    case "linux":
        provider = NewLinuxNotificationProvider("NetMonitor")
    default:
        return fmt.Errorf("notifications not supported on %s", runtime.GOOS)
    }

    // Create notification manager
    a.notificationManager = NewNotificationManager(provider, a.config.Notifications, a.logger)

    // Register notification templates
    a.registerNotificationTemplates()

    return nil
}

func (a *App) registerNotificationTemplates() {
    templates := map[string]*NotificationTemplate{
        "endpoint_failure": {
            Title:    "Endpoint Failed: {{.EndpointName}}",
            Message:  "{{.EndpointName}} in {{.Region}} is not responding. Last seen: {{.LastSeen}}",
            Priority: PriorityHigh,
            Category: CategoryEndpointFailure,
            Actions: []NotificationAction{
                {ID: "test_now", Title: "Test Now"},
                {ID: "view_details", Title: "View Details"},
            },
        },
        "threshold_breach": {
            Title:    "Threshold Exceeded: {{.EndpointName}}",
            Message:  "{{.Metric}} is {{.Value}} (threshold: {{.Threshold}}) for {{.EndpointName}}",
            Priority: PriorityNormal,
            Category: CategoryThresholdBreach,
            Actions: []NotificationAction{
                {ID: "view_graph", Title: "View Graph"},
                {ID: "adjust_threshold", Title: "Adjust Threshold"},
            },
        },
        "system_alert": {
            Title:    "NetMonitor Alert",
            Message:  "{{.Message}}",
            Priority: PriorityCritical,
            Category: CategorySystemAlert,
        },
        "test_complete": {
            Title:    "Manual Test Complete",
            Message:  "Tested {{.EndpointCount}} endpoints. {{.SuccessCount}} successful, {{.FailureCount}} failed.",
            Priority: PriorityLow,
            Category: CategoryTestComplete,
        },
    }

    for name, template := range templates {
        a.notificationManager.RegisterTemplate(name, template)
    }
}

// Methods for showing notifications
func (a *App) NotifyEndpointFailure(endpoint *Endpoint, lastSeen time.Time) {
    notification, err := a.notificationManager.CreateFromTemplate("endpoint_failure", map[string]interface{}{
        "EndpointName": endpoint.Name,
        "Region":       endpoint.Region,
        "LastSeen":     lastSeen.Format("15:04"),
    })

    if err != nil {
        a.logger.Error("Failed to create endpoint failure notification", Error(err))
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := a.notificationManager.ShowNotification(ctx, notification); err != nil {
        a.logger.Error("Failed to show endpoint failure notification", Error(err))
    }
}

func (a *App) NotifyThresholdBreach(endpoint *Endpoint, metric string, value, threshold float64) {
    notification, err := a.notificationManager.CreateFromTemplate("threshold_breach", map[string]interface{}{
        "EndpointName": endpoint.Name,
        "Metric":       metric,
        "Value":        fmt.Sprintf("%.2f", value),
        "Threshold":    fmt.Sprintf("%.2f", threshold),
    })

    if err != nil {
        a.logger.Error("Failed to create threshold breach notification", Error(err))
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := a.notificationManager.ShowNotification(ctx, notification); err != nil {
        a.logger.Error("Failed to show threshold breach notification", Error(err))
    }
}

// Configuration methods
func (a *App) UpdateNotificationConfig(config *NotificationConfig) error {
    a.config.Notifications = config
    a.notificationManager.UpdateConfig(config)

    // Save configuration
    return a.SaveConfiguration()
}

func (a *App) GetNotificationConfig() (*NotificationConfig, error) {
    return a.config.Notifications, nil
}

func (a *App) GetNotificationHistory() ([]*Notification, error) {
    return a.notificationManager.GetHistory(), nil
}

func (a *App) TestNotification() error {
    notification := &Notification{
        ID:        "test",
        Title:     "NetMonitor Test",
        Message:   "This is a test notification to verify your settings.",
        Priority:  PriorityNormal,
        Category:  CategorySystemAlert,
        Timestamp: time.Now(),
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return a.notificationManager.ShowNotification(ctx, notification)
}
```

## Notification History Management
```go
type NotificationHistory struct {
    notifications []*Notification
    maxSize       int
    mu            sync.RWMutex
}

func NewNotificationHistory() *NotificationHistory {
    return &NotificationHistory{
        notifications: make([]*Notification, 0),
        maxSize:       1000, // Keep last 1000 notifications
    }
}

func (nh *NotificationHistory) Add(notification *Notification) {
    nh.mu.Lock()
    defer nh.mu.Unlock()

    nh.notifications = append(nh.notifications, notification)

    // Trim to max size
    if len(nh.notifications) > nh.maxSize {
        nh.notifications = nh.notifications[len(nh.notifications)-nh.maxSize:]
    }
}

func (nh *NotificationHistory) GetRecent(limit int) []*Notification {
    nh.mu.RLock()
    defer nh.mu.RUnlock()

    if limit <= 0 || limit > len(nh.notifications) {
        limit = len(nh.notifications)
    }

    start := len(nh.notifications) - limit
    result := make([]*Notification, limit)
    copy(result, nh.notifications[start:])

    return result
}

func (nh *NotificationHistory) Clear() {
    nh.mu.Lock()
    defer nh.mu.Unlock()

    nh.notifications = nh.notifications[:0]
}
```

## Verification Steps
1. Test cross-platform notifications - should show native notifications on all platforms
2. Verify notification templates - should interpolate data correctly
3. Test rate limiting - should prevent notification spam
4. Verify Do Not Disturb mode - should suppress notifications during configured hours
5. Test notification actions - should handle action callbacks where supported
6. Verify configuration changes - should apply new settings immediately
7. Test notification history - should track and retrieve past notifications
8. Verify priority handling - should respect minimum priority settings

## Dependencies
- T036: System Tray Integration
- T039: Comprehensive Logging System
- T015: Monitoring Status API
- Platform-specific notification libraries

## Notes
- Test notification appearance and behavior on each target platform
- Consider implementing notification sound customization
- Plan for future rich notification features (images, progress bars)
- Ensure notifications comply with platform guidelines
- Consider implementing notification analytics for optimization
- Plan for future email/SMS notification extensions