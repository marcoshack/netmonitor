# T043: Email Notifications

## Overview
Implement email notification capabilities for NetMonitor alerts and reports, including SMTP configuration, email templates, delivery tracking, and integration with the alert management system.

## Context
While system notifications provide immediate alerts, email notifications are essential for situations where users are away from their workstation, for team notifications, and for detailed reporting. This extends NetMonitor's alerting capabilities beyond local notifications.

## Task Description
Create a comprehensive email notification system with SMTP support, HTML/text email templates, delivery tracking, attachment support, and seamless integration with the existing alert and notification framework.

## Acceptance Criteria
- [ ] SMTP client configuration and connection management
- [ ] HTML and plain text email templates
- [ ] Email delivery tracking and retry logic
- [ ] Attachment support for reports and logs
- [ ] Integration with alert management system
- [ ] Email notification preferences and filtering
- [ ] Bulk email handling and rate limiting
- [ ] Email authentication (OAuth2, App Passwords)
- [ ] Delivery status reporting and error handling

## Email System Architecture
```go
package email

import (
    "bytes"
    "context"
    "crypto/tls"
    "fmt"
    "html/template"
    "io"
    "mime"
    "mime/multipart"
    "net/smtp"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

type EmailManager struct {
    config       *EmailConfig
    smtpClient   SMTPClient
    templates    map[string]*EmailTemplate
    deliveryLog  *DeliveryLog
    rateLimiter  *EmailRateLimiter
    retryQueue   *RetryQueue
    mutex        sync.RWMutex
    logger       Logger
}

type EmailConfig struct {
    Enabled         bool              `json:"enabled"`
    SMTPHost        string            `json:"smtpHost"`
    SMTPPort        int               `json:"smtpPort"`
    Username        string            `json:"username"`
    Password        string            `json:"password"`
    FromAddress     string            `json:"fromAddress"`
    FromName        string            `json:"fromName"`
    Security        SecurityType      `json:"security"`
    Authentication  AuthenticationType `json:"authentication"`
    Timeout         time.Duration     `json:"timeout"`
    MaxRetries      int               `json:"maxRetries"`
    RetryDelay      time.Duration     `json:"retryDelay"`
    RateLimit       *RateLimitConfig  `json:"rateLimit"`
    TestRecipient   string            `json:"testRecipient"`
}

type SecurityType string

const (
    SecurityNone     SecurityType = "none"
    SecuritySTARTTLS SecurityType = "starttls"
    SecuritySSL      SecurityType = "ssl"
)

type AuthenticationType string

const (
    AuthNone       AuthenticationType = "none"
    AuthPlain      AuthenticationType = "plain"
    AuthLogin      AuthenticationType = "login"
    AuthCRAMMD5    AuthenticationType = "cram-md5"
    AuthOAuth2     AuthenticationType = "oauth2"
)

type Email struct {
    ID          string            `json:"id"`
    To          []string          `json:"to"`
    CC          []string          `json:"cc,omitempty"`
    BCC         []string          `json:"bcc,omitempty"`
    Subject     string            `json:"subject"`
    Body        string            `json:"body"`
    HTMLBody    string            `json:"htmlBody,omitempty"`
    Attachments []EmailAttachment `json:"attachments,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
    Priority    EmailPriority     `json:"priority"`
    Category    string            `json:"category"`
    Tags        []string          `json:"tags,omitempty"`
    CreatedAt   time.Time         `json:"createdAt"`
    ScheduledAt *time.Time        `json:"scheduledAt,omitempty"`
}

type EmailAttachment struct {
    Filename    string `json:"filename"`
    ContentType string `json:"contentType"`
    Data        []byte `json:"data"`
    Inline      bool   `json:"inline"`
}

type EmailPriority string

const (
    PriorityLow    EmailPriority = "low"
    PriorityNormal EmailPriority = "normal"
    PriorityHigh   EmailPriority = "high"
    PriorityUrgent EmailPriority = "urgent"
)

type EmailTemplate struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Subject      string            `json:"subject"`
    TextTemplate string            `json:"textTemplate"`
    HTMLTemplate string            `json:"htmlTemplate"`
    Category     string            `json:"category"`
    Variables    []string          `json:"variables"`
    CreatedAt    time.Time         `json:"createdAt"`
    UpdatedAt    time.Time         `json:"updatedAt"`
}

func NewEmailManager(config *EmailConfig, logger Logger) *EmailManager {
    return &EmailManager{
        config:      config,
        templates:   make(map[string]*EmailTemplate),
        deliveryLog: NewDeliveryLog(),
        rateLimiter: NewEmailRateLimiter(config.RateLimit),
        retryQueue:  NewRetryQueue(),
        logger:      logger,
    }
}

func (em *EmailManager) Initialize() error {
    if !em.config.Enabled {
        em.logger.Info("Email notifications disabled")
        return nil
    }

    // Initialize SMTP client
    var err error
    em.smtpClient, err = NewSMTPClient(em.config)
    if err != nil {
        return fmt.Errorf("failed to initialize SMTP client: %w", err)
    }

    // Load email templates
    if err := em.loadDefaultTemplates(); err != nil {
        return fmt.Errorf("failed to load email templates: %w", err)
    }

    // Start retry queue processor
    go em.processRetryQueue()

    em.logger.Info("Email manager initialized successfully")
    return nil
}

func (em *EmailManager) SendEmail(ctx context.Context, email *Email) error {
    if !em.config.Enabled {
        return fmt.Errorf("email notifications are disabled")
    }

    // Generate unique ID if not provided
    if email.ID == "" {
        email.ID = generateEmailID()
    }

    // Set creation time
    email.CreatedAt = time.Now()

    // Validate email
    if err := em.validateEmail(email); err != nil {
        return fmt.Errorf("email validation failed: %w", err)
    }

    // Check rate limiting
    if !em.rateLimiter.Allow() {
        em.logger.Warn("Email rate limited, queuing for retry",
            Field{Key: "email_id", Value: email.ID})
        em.retryQueue.Add(email, time.Now().Add(em.config.RetryDelay))
        return nil
    }

    // Attempt to send email
    if err := em.sendEmailWithRetry(ctx, email); err != nil {
        em.logger.Error("Failed to send email",
            Field{Key: "email_id", Value: email.ID},
            Field{Key: "error", Value: err.Error()})

        // Add to retry queue if retryable
        if em.isRetryableError(err) && em.getRetryCount(email.ID) < em.config.MaxRetries {
            retryAt := time.Now().Add(em.calculateRetryDelay(em.getRetryCount(email.ID)))
            em.retryQueue.Add(email, retryAt)
            em.logger.Info("Email queued for retry",
                Field{Key: "email_id", Value: email.ID},
                Field{Key: "retry_at", Value: retryAt})
        }

        return err
    }

    em.logger.Info("Email sent successfully",
        Field{Key: "email_id", Value: email.ID},
        Field{Key: "recipients", Value: len(email.To)})

    return nil
}

func (em *EmailManager) sendEmailWithRetry(ctx context.Context, email *Email) error {
    // Create MIME message
    message, err := em.createMIMEMessage(email)
    if err != nil {
        return fmt.Errorf("failed to create MIME message: %w", err)
    }

    // Get all recipients
    recipients := append(email.To, email.CC...)
    recipients = append(recipients, email.BCC...)

    // Send email via SMTP
    if err := em.smtpClient.SendMail(ctx, em.config.FromAddress, recipients, message); err != nil {
        // Log delivery failure
        em.deliveryLog.LogDelivery(&DeliveryRecord{
            EmailID:     email.ID,
            Recipients:  recipients,
            Status:      DeliveryStatusFailed,
            Error:       err.Error(),
            Timestamp:   time.Now(),
        })
        return err
    }

    // Log successful delivery
    em.deliveryLog.LogDelivery(&DeliveryRecord{
        EmailID:    email.ID,
        Recipients: recipients,
        Status:     DeliveryStatusSent,
        Timestamp:  time.Now(),
    })

    return nil
}

func (em *EmailManager) createMIMEMessage(email *Email) ([]byte, error) {
    var buffer bytes.Buffer

    // Write headers
    buffer.WriteString(fmt.Sprintf("From: %s <%s>\r\n", em.config.FromName, em.config.FromAddress))
    buffer.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))

    if len(email.CC) > 0 {
        buffer.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.CC, ", ")))
    }

    buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
    buffer.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
    buffer.WriteString("MIME-Version: 1.0\r\n")

    // Set priority
    if email.Priority != PriorityNormal {
        buffer.WriteString(fmt.Sprintf("X-Priority: %s\r\n", em.getPriorityValue(email.Priority)))
    }

    // Add custom headers
    for key, value := range email.Headers {
        buffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
    }

    // Create multipart message if we have attachments or HTML
    if len(email.Attachments) > 0 || email.HTMLBody != "" {
        writer := multipart.NewWriter(&buffer)
        boundary := writer.Boundary()

        buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

        // Write text/HTML parts
        if err := em.writeBodyParts(writer, email); err != nil {
            return nil, err
        }

        // Write attachments
        for _, attachment := range email.Attachments {
            if err := em.writeAttachment(writer, &attachment); err != nil {
                return nil, err
            }
        }

        writer.Close()
    } else {
        // Simple text message
        buffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
        buffer.WriteString(email.Body)
    }

    return buffer.Bytes(), nil
}

// Email Templates
func (em *EmailManager) CreateEmailFromTemplate(templateID string, data map[string]interface{}) (*Email, error) {
    em.mutex.RLock()
    template, exists := em.templates[templateID]
    em.mutex.RUnlock()

    if !exists {
        return nil, fmt.Errorf("template not found: %s", templateID)
    }

    // Execute templates
    subject, err := em.executeTemplate(template.Subject, data)
    if err != nil {
        return nil, fmt.Errorf("failed to execute subject template: %w", err)
    }

    textBody, err := em.executeTemplate(template.TextTemplate, data)
    if err != nil {
        return nil, fmt.Errorf("failed to execute text template: %w", err)
    }

    var htmlBody string
    if template.HTMLTemplate != "" {
        htmlBody, err = em.executeTemplate(template.HTMLTemplate, data)
        if err != nil {
            return nil, fmt.Errorf("failed to execute HTML template: %w", err)
        }
    }

    email := &Email{
        Subject:  subject,
        Body:     textBody,
        HTMLBody: htmlBody,
        Category: template.Category,
        Priority: PriorityNormal,
    }

    return email, nil
}

func (em *EmailManager) executeTemplate(templateStr string, data map[string]interface{}) (string, error) {
    tmpl, err := template.New("email").Parse(templateStr)
    if err != nil {
        return "", err
    }

    var buffer bytes.Buffer
    if err := tmpl.Execute(&buffer, data); err != nil {
        return "", err
    }

    return buffer.String(), nil
}

func (em *EmailManager) loadDefaultTemplates() error {
    templates := map[string]*EmailTemplate{
        "endpoint_failure": {
            ID:       "endpoint_failure",
            Name:     "Endpoint Failure Alert",
            Subject:  "🚨 NetMonitor Alert: {{.EndpointName}} is Down",
            Category: "alert",
            TextTemplate: `
NetMonitor Alert: Endpoint Failure

Endpoint: {{.EndpointName}}
Region: {{.Region}}
Type: {{.Type}}
Address: {{.Address}}
Status: DOWN
Last Seen: {{.LastSeen}}
Duration: {{.Duration}}

This endpoint has been unavailable for {{.Duration}}. Please investigate the issue.

Dashboard: {{.DashboardURL}}
`,
            HTMLTemplate: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>NetMonitor Alert</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background-color: #dc3545; color: white; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
        .alert-icon { font-size: 24px; margin-right: 10px; }
        .details { background-color: #f8f9fa; padding: 15px; border-radius: 4px; margin: 15px 0; }
        .detail-row { margin: 8px 0; }
        .label { font-weight: bold; color: #495057; }
        .value { color: #212529; }
        .footer { margin-top: 20px; padding-top: 15px; border-top: 1px solid #dee2e6; font-size: 12px; color: #6c757d; }
        .action-button { background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; display: inline-block; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <span class="alert-icon">🚨</span>
            <strong>NetMonitor Alert: Endpoint Failure</strong>
        </div>

        <p>The following endpoint is currently down and requires immediate attention:</p>

        <div class="details">
            <div class="detail-row">
                <span class="label">Endpoint:</span>
                <span class="value">{{.EndpointName}}</span>
            </div>
            <div class="detail-row">
                <span class="label">Region:</span>
                <span class="value">{{.Region}}</span>
            </div>
            <div class="detail-row">
                <span class="label">Type:</span>
                <span class="value">{{.Type}}</span>
            </div>
            <div class="detail-row">
                <span class="label">Address:</span>
                <span class="value">{{.Address}}</span>
            </div>
            <div class="detail-row">
                <span class="label">Status:</span>
                <span class="value" style="color: #dc3545; font-weight: bold;">DOWN</span>
            </div>
            <div class="detail-row">
                <span class="label">Last Seen:</span>
                <span class="value">{{.LastSeen}}</span>
            </div>
            <div class="detail-row">
                <span class="label">Duration:</span>
                <span class="value">{{.Duration}}</span>
            </div>
        </div>

        <p>This endpoint has been unavailable for {{.Duration}}. Please investigate the issue immediately.</p>

        <a href="{{.DashboardURL}}" class="action-button">View Dashboard</a>

        <div class="footer">
            This alert was generated by NetMonitor. If you no longer wish to receive these notifications, please update your preferences in the application settings.
        </div>
    </div>
</body>
</html>
`,
        },
        "threshold_breach": {
            ID:       "threshold_breach",
            Name:     "Threshold Breach Alert",
            Subject:  "⚠️ NetMonitor Warning: {{.Metric}} threshold exceeded for {{.EndpointName}}",
            Category: "alert",
            TextTemplate: `
NetMonitor Warning: Threshold Breach

Endpoint: {{.EndpointName}}
Region: {{.Region}}
Metric: {{.Metric}}
Current Value: {{.CurrentValue}}
Threshold: {{.ThresholdValue}}
Duration: {{.Duration}}

The {{.Metric}} for {{.EndpointName}} has exceeded the configured threshold.

Dashboard: {{.DashboardURL}}
`,
            HTMLTemplate: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>NetMonitor Warning</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background-color: #ffc107; color: #212529; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
        .warning-icon { font-size: 24px; margin-right: 10px; }
        .details { background-color: #fff3cd; padding: 15px; border-radius: 4px; margin: 15px 0; border-left: 4px solid #ffc107; }
        .metric-comparison { background-color: #f8f9fa; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .current-value { font-size: 18px; font-weight: bold; color: #dc3545; }
        .threshold-value { font-size: 16px; color: #6c757d; }
        .footer { margin-top: 20px; padding-top: 15px; border-top: 1px solid #dee2e6; font-size: 12px; color: #6c757d; }
        .action-button { background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; display: inline-block; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <span class="warning-icon">⚠️</span>
            <strong>NetMonitor Warning: Threshold Breach</strong>
        </div>

        <p>A performance threshold has been exceeded for the following endpoint:</p>

        <div class="details">
            <h3>{{.EndpointName}} ({{.Region}})</h3>
            <div class="metric-comparison">
                <div><strong>{{.Metric}}</strong></div>
                <div class="current-value">Current: {{.CurrentValue}}</div>
                <div class="threshold-value">Threshold: {{.ThresholdValue}}</div>
            </div>
            <p><strong>Duration:</strong> {{.Duration}}</p>
        </div>

        <a href="{{.DashboardURL}}" class="action-button">View Dashboard</a>

        <div class="footer">
            This warning was generated by NetMonitor. You can adjust thresholds in the application settings.
        </div>
    </div>
</body>
</html>
`,
        },
        "daily_report": {
            ID:       "daily_report",
            Name:     "Daily Monitoring Report",
            Subject:  "📊 NetMonitor Daily Report - {{.Date}}",
            Category: "report",
            TextTemplate: `
NetMonitor Daily Report - {{.Date}}

Summary:
- Total Endpoints: {{.TotalEndpoints}}
- Healthy Endpoints: {{.HealthyEndpoints}}
- Warning Endpoints: {{.WarningEndpoints}}
- Down Endpoints: {{.DownEndpoints}}
- Average Uptime: {{.AverageUptime}}%

Regional Performance:
{{range .Regions}}
- {{.Name}}: {{.Uptime}}% uptime, {{.AvgLatency}}ms avg latency
{{end}}

{{if .Incidents}}
Incidents Today:
{{range .Incidents}}
- {{.Time}}: {{.EndpointName}} - {{.Description}}
{{end}}
{{end}}

Dashboard: {{.DashboardURL}}
`,
        },
    }

    for id, template := range templates {
        template.CreatedAt = time.Now()
        template.UpdatedAt = time.Now()
        em.templates[id] = template
    }

    return nil
}

// Integration with Alert System
func (em *EmailManager) SendAlertEmail(alert *Alert, recipients []string) error {
    var templateID string
    switch alert.Severity {
    case SeverityCritical, SeverityEmergency:
        templateID = "endpoint_failure"
    default:
        templateID = "threshold_breach"
    }

    data := map[string]interface{}{
        "EndpointName":   alert.Labels["endpoint_name"],
        "Region":         alert.Labels["region"],
        "Type":           alert.Labels["type"],
        "Address":        alert.Labels["address"],
        "LastSeen":       alert.StartsAt.Format("2006-01-02 15:04:05"),
        "Duration":       time.Since(alert.StartsAt).String(),
        "DashboardURL":   em.config.DashboardURL,
        "Metric":         alert.Labels["metric"],
        "CurrentValue":   alert.Labels["current_value"],
        "ThresholdValue": alert.Labels["threshold_value"],
    }

    email, err := em.CreateEmailFromTemplate(templateID, data)
    if err != nil {
        return fmt.Errorf("failed to create email from template: %w", err)
    }

    email.To = recipients
    email.Priority = em.mapAlertSeverityToEmailPriority(alert.Severity)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return em.SendEmail(ctx, email)
}

// Configuration and Testing
func (em *EmailManager) TestEmailConfiguration() error {
    if em.config.TestRecipient == "" {
        return fmt.Errorf("test recipient not configured")
    }

    testEmail := &Email{
        To:       []string{em.config.TestRecipient},
        Subject:  "NetMonitor Email Test",
        Body:     "This is a test email from NetMonitor. If you receive this, your email configuration is working correctly.",
        Priority: PriorityNormal,
        Category: "test",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return em.SendEmail(ctx, testEmail)
}

func (em *EmailManager) UpdateEmailConfig(config *EmailConfig) error {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    em.config = config

    // Reinitialize SMTP client with new config
    if config.Enabled {
        smtpClient, err := NewSMTPClient(config)
        if err != nil {
            return fmt.Errorf("failed to reinitialize SMTP client: %w", err)
        }
        em.smtpClient = smtpClient
    }

    return nil
}
```

## SMTP Client Implementation
```go
type SMTPClient interface {
    SendMail(ctx context.Context, from string, to []string, msg []byte) error
    Close() error
}

type DefaultSMTPClient struct {
    config *EmailConfig
    auth   smtp.Auth
}

func NewSMTPClient(config *EmailConfig) (SMTPClient, error) {
    client := &DefaultSMTPClient{
        config: config,
    }

    // Setup authentication
    if config.Authentication != AuthNone {
        client.auth = smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
    }

    return client, nil
}

func (c *DefaultSMTPClient) SendMail(ctx context.Context, from string, to []string, msg []byte) error {
    addr := fmt.Sprintf("%s:%d", c.config.SMTPHost, c.config.SMTPPort)

    // Connect with timeout
    conn, err := smtp.Dial(addr)
    if err != nil {
        return fmt.Errorf("failed to connect to SMTP server: %w", err)
    }
    defer conn.Close()

    // STARTTLS if required
    if c.config.Security == SecuritySTARTTLS {
        tlsConfig := &tls.Config{ServerName: c.config.SMTPHost}
        if err = conn.StartTLS(tlsConfig); err != nil {
            return fmt.Errorf("failed to start TLS: %w", err)
        }
    }

    // Authenticate
    if c.auth != nil {
        if err = conn.Auth(c.auth); err != nil {
            return fmt.Errorf("authentication failed: %w", err)
        }
    }

    // Send email
    if err = conn.Mail(from); err != nil {
        return fmt.Errorf("failed to set sender: %w", err)
    }

    for _, recipient := range to {
        if err = conn.Rcpt(recipient); err != nil {
            return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
        }
    }

    dataWriter, err := conn.Data()
    if err != nil {
        return fmt.Errorf("failed to get data writer: %w", err)
    }

    _, err = dataWriter.Write(msg)
    if err != nil {
        return fmt.Errorf("failed to write message: %w", err)
    }

    return dataWriter.Close()
}
```

## Application Integration
```go
// App integration with email notifications
func (a *App) initializeEmailNotifications() error {
    a.emailManager = NewEmailManager(a.config.Email, a.logger)
    return a.emailManager.Initialize()
}

// API methods for email management
func (a *App) UpdateEmailConfig(config *EmailConfig) error {
    if err := a.emailManager.UpdateEmailConfig(config); err != nil {
        return err
    }
    a.config.Email = config
    return a.SaveConfiguration()
}

func (a *App) TestEmailConfiguration() error {
    return a.emailManager.TestEmailConfiguration()
}

func (a *App) SendDailyReport(recipients []string) error {
    // Generate daily report data
    reportData := a.generateDailyReportData()

    email, err := a.emailManager.CreateEmailFromTemplate("daily_report", reportData)
    if err != nil {
        return err
    }

    email.To = recipients
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return a.emailManager.SendEmail(ctx, email)
}
```

## Verification Steps
1. Test SMTP connection - should connect to email server successfully
2. Verify email templates - should render with correct data
3. Test email delivery - should send emails to configured recipients
4. Verify HTML/text rendering - should display correctly in email clients
5. Test attachment handling - should attach files correctly
6. Verify rate limiting - should respect configured sending limits
7. Test retry logic - should retry failed deliveries appropriately
8. Verify integration with alerts - should send emails for configured alert conditions

## Dependencies
- T042: Alert Management System
- T041: System Notifications
- T039: Comprehensive Logging System
- T003: Configuration System

## Notes
- Test with multiple email providers (Gmail, Outlook, SMTP servers)
- Consider implementing email tracking and analytics
- Plan for future features like email scheduling and templates
- Ensure compliance with anti-spam regulations
- Consider implementing email signatures and branding
- Plan for integration with email marketing platforms if needed