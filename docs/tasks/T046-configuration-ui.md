# T046: Configuration UI

## Overview
Create a comprehensive configuration interface for all application settings, providing an intuitive and organized way for users to manage NetMonitor preferences, monitoring targets, notification rules, and system settings.

## Context
As part of the NetMonitor application, users need an easy-to-use interface to configure all aspects of the monitoring system. This includes general application settings, monitoring targets, alert thresholds, notification preferences, and system integration options. The configuration UI should be well-organized, responsive, and provide immediate feedback for validation errors.

## Task Description
Implement a comprehensive configuration user interface that allows users to:
- Access all application settings through a centralized interface
- Organize settings into logical categories with tabs or sections
- Validate configuration changes in real-time
- Apply settings with immediate feedback
- Reset settings to defaults
- Import/export configuration (preparation for T049)

## Acceptance Criteria
- [ ] Configuration UI is accessible from the main application menu
- [ ] Settings are organized into logical categories (General, Monitoring, Notifications, Advanced)
- [ ] All configuration options from the specification are accessible
- [ ] Real-time validation provides immediate feedback for invalid inputs
- [ ] Changes can be applied, canceled, or reset to defaults
- [ ] UI is responsive and works across different screen sizes
- [ ] Keyboard navigation is supported for accessibility
- [ ] Configuration changes trigger appropriate system updates
- [ ] Help tooltips provide guidance for complex settings
- [ ] Unsaved changes are properly handled (warning before navigation)

## Implementation Details

### Frontend Configuration Component
Create `frontend/config.html`:
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NetMonitor Configuration</title>
    <link rel="stylesheet" href="style.css">
    <link rel="stylesheet" href="config.css">
</head>
<body>
    <div class="config-container">
        <header class="config-header">
            <h1>NetMonitor Configuration</h1>
            <div class="config-actions">
                <button id="resetBtn" class="btn btn-secondary">Reset to Defaults</button>
                <button id="cancelBtn" class="btn btn-secondary">Cancel</button>
                <button id="saveBtn" class="btn btn-primary">Save Changes</button>
            </div>
        </header>

        <nav class="config-nav">
            <button class="nav-tab active" data-tab="general">General</button>
            <button class="nav-tab" data-tab="monitoring">Monitoring</button>
            <button class="nav-tab" data-tab="notifications">Notifications</button>
            <button class="nav-tab" data-tab="advanced">Advanced</button>
        </nav>

        <main class="config-content">
            <!-- General Settings Tab -->
            <div id="general-tab" class="tab-content active">
                <section class="config-section">
                    <h2>Application Settings</h2>
                    <div class="form-group">
                        <label for="appName">Application Name</label>
                        <input type="text" id="appName" name="appName" placeholder="NetMonitor">
                        <span class="help-text">Display name for the application</span>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="autoStart" name="autoStart">
                            Start with system
                        </label>
                        <span class="help-text">Automatically start NetMonitor when the system boots</span>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="minimizeToTray" name="minimizeToTray">
                            Minimize to system tray
                        </label>
                        <span class="help-text">Keep application running in system tray when closed</span>
                    </div>
                    <div class="form-group">
                        <label for="theme">Theme</label>
                        <select id="theme" name="theme">
                            <option value="light">Light</option>
                            <option value="dark">Dark</option>
                            <option value="auto">Auto (System)</option>
                        </select>
                    </div>
                </section>

                <section class="config-section">
                    <h2>Data Management</h2>
                    <div class="form-group">
                        <label for="dataRetention">Data Retention (days)</label>
                        <input type="number" id="dataRetention" name="dataRetention" min="1" max="365" value="30">
                        <span class="help-text">How long to keep monitoring data (1-365 days)</span>
                    </div>
                    <div class="form-group">
                        <label for="logLevel">Log Level</label>
                        <select id="logLevel" name="logLevel">
                            <option value="debug">Debug</option>
                            <option value="info">Info</option>
                            <option value="warn">Warning</option>
                            <option value="error">Error</option>
                        </select>
                    </div>
                </section>
            </div>

            <!-- Monitoring Settings Tab -->
            <div id="monitoring-tab" class="tab-content">
                <section class="config-section">
                    <h2>Default Monitoring Settings</h2>
                    <div class="form-group">
                        <label for="defaultInterval">Default Check Interval (seconds)</label>
                        <input type="number" id="defaultInterval" name="defaultInterval" min="5" max="3600" value="60">
                        <span class="help-text">Default interval for new monitoring targets</span>
                    </div>
                    <div class="form-group">
                        <label for="defaultTimeout">Default Timeout (seconds)</label>
                        <input type="number" id="defaultTimeout" name="defaultTimeout" min="1" max="120" value="10">
                        <span class="help-text">Default timeout for monitoring requests</span>
                    </div>
                    <div class="form-group">
                        <label for="maxConcurrent">Max Concurrent Checks</label>
                        <input type="number" id="maxConcurrent" name="maxConcurrent" min="1" max="100" value="10">
                        <span class="help-text">Maximum number of simultaneous monitoring checks</span>
                    </div>
                </section>

                <section class="config-section">
                    <h2>Alert Thresholds</h2>
                    <div class="form-group">
                        <label for="responseThreshold">Response Time Threshold (ms)</label>
                        <input type="number" id="responseThreshold" name="responseThreshold" min="100" max="10000" value="5000">
                        <span class="help-text">Response time that triggers slow response alerts</span>
                    </div>
                    <div class="form-group">
                        <label for="failureThreshold">Failure Threshold</label>
                        <input type="number" id="failureThreshold" name="failureThreshold" min="1" max="10" value="3">
                        <span class="help-text">Consecutive failures before triggering alerts</span>
                    </div>
                    <div class="form-group">
                        <label for="recoveryThreshold">Recovery Threshold</label>
                        <input type="number" id="recoveryThreshold" name="recoveryThreshold" min="1" max="10" value="2">
                        <span class="help-text">Consecutive successes before marking as recovered</span>
                    </div>
                </section>
            </div>

            <!-- Notifications Settings Tab -->
            <div id="notifications-tab" class="tab-content">
                <section class="config-section">
                    <h2>System Notifications</h2>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="enableSystemNotifications" name="enableSystemNotifications">
                            Enable system notifications
                        </label>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="notifyOnFailure" name="notifyOnFailure">
                            Notify on failures
                        </label>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="notifyOnRecovery" name="notifyOnRecovery">
                            Notify on recovery
                        </label>
                    </div>
                </section>

                <section class="config-section">
                    <h2>Email Notifications</h2>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="enableEmail" name="enableEmail">
                            Enable email notifications
                        </label>
                    </div>
                    <div class="form-group">
                        <label for="smtpServer">SMTP Server</label>
                        <input type="text" id="smtpServer" name="smtpServer" placeholder="smtp.example.com">
                    </div>
                    <div class="form-group">
                        <label for="smtpPort">SMTP Port</label>
                        <input type="number" id="smtpPort" name="smtpPort" min="1" max="65535" value="587">
                    </div>
                    <div class="form-group">
                        <label for="smtpUsername">Username</label>
                        <input type="text" id="smtpUsername" name="smtpUsername">
                    </div>
                    <div class="form-group">
                        <label for="smtpPassword">Password</label>
                        <input type="password" id="smtpPassword" name="smtpPassword">
                    </div>
                    <div class="form-group">
                        <label for="fromEmail">From Email</label>
                        <input type="email" id="fromEmail" name="fromEmail" placeholder="netmonitor@example.com">
                    </div>
                    <div class="form-group">
                        <label for="toEmails">To Emails (comma-separated)</label>
                        <textarea id="toEmails" name="toEmails" placeholder="admin@example.com, ops@example.com"></textarea>
                    </div>
                    <button type="button" id="testEmailBtn" class="btn btn-secondary">Test Email</button>
                </section>
            </div>

            <!-- Advanced Settings Tab -->
            <div id="advanced-tab" class="tab-content">
                <section class="config-section">
                    <h2>Network Settings</h2>
                    <div class="form-group">
                        <label for="userAgent">HTTP User Agent</label>
                        <input type="text" id="userAgent" name="userAgent" placeholder="NetMonitor/1.0">
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="followRedirects" name="followRedirects">
                            Follow HTTP redirects
                        </label>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="verifySsl" name="verifySsl">
                            Verify SSL certificates
                        </label>
                    </div>
                </section>

                <section class="config-section">
                    <h2>Performance</h2>
                    <div class="form-group">
                        <label for="workerThreads">Worker Threads</label>
                        <input type="number" id="workerThreads" name="workerThreads" min="1" max="16" value="4">
                        <span class="help-text">Number of worker threads for monitoring tasks</span>
                    </div>
                    <div class="form-group">
                        <label for="cacheSize">Cache Size (MB)</label>
                        <input type="number" id="cacheSize" name="cacheSize" min="10" max="1000" value="100">
                        <span class="help-text">Maximum memory cache size for monitoring data</span>
                    </div>
                </section>

                <section class="config-section">
                    <h2>Development</h2>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="debugMode" name="debugMode">
                            Enable debug mode
                        </label>
                        <span class="help-text">Enable detailed logging and debugging features</span>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="apiAccess" name="apiAccess">
                            Enable API access
                        </label>
                        <span class="help-text">Allow external API access to monitoring data</span>
                    </div>
                </section>
            </div>
        </main>
    </div>

    <script src="config.js"></script>
</body>
</html>
```

### Configuration Styles
Create `frontend/config.css`:
```css
.config-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.config-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--border-color);
}

.config-header h1 {
    margin: 0;
    color: var(--text-primary);
}

.config-actions {
    display: flex;
    gap: 10px;
}

.config-nav {
    display: flex;
    gap: 5px;
    margin-bottom: 30px;
    border-bottom: 1px solid var(--border-color);
}

.nav-tab {
    padding: 12px 24px;
    border: none;
    background: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all 0.2s ease;
}

.nav-tab:hover {
    color: var(--text-primary);
    background: var(--bg-secondary);
}

.nav-tab.active {
    color: var(--primary-color);
    border-bottom-color: var(--primary-color);
}

.tab-content {
    display: none;
}

.tab-content.active {
    display: block;
}

.config-section {
    margin-bottom: 40px;
    padding: 20px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px solid var(--border-color);
}

.config-section h2 {
    margin: 0 0 20px 0;
    color: var(--text-primary);
    font-size: 1.2em;
    font-weight: 600;
}

.form-group {
    margin-bottom: 20px;
}

.form-group label {
    display: block;
    margin-bottom: 8px;
    color: var(--text-primary);
    font-weight: 500;
}

.form-group input[type="text"],
.form-group input[type="number"],
.form-group input[type="email"],
.form-group input[type="password"],
.form-group select,
.form-group textarea {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 14px;
    transition: border-color 0.2s ease;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
    border-color: var(--primary-color);
    box-shadow: 0 0 0 2px var(--primary-color)20;
}

.form-group input[type="checkbox"] {
    margin-right: 8px;
}

.form-group textarea {
    resize: vertical;
    min-height: 80px;
}

.help-text {
    display: block;
    margin-top: 4px;
    font-size: 12px;
    color: var(--text-secondary);
    font-style: italic;
}

.form-group.error input,
.form-group.error select,
.form-group.error textarea {
    border-color: var(--error-color);
}

.error-message {
    display: block;
    margin-top: 4px;
    font-size: 12px;
    color: var(--error-color);
}

.btn {
    padding: 10px 20px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    transition: all 0.2s ease;
}

.btn-primary {
    background: var(--primary-color);
    color: white;
}

.btn-primary:hover {
    background: var(--primary-hover);
}

.btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
}

.btn-secondary:hover {
    background: var(--bg-secondary);
}

.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

@media (max-width: 768px) {
    .config-header {
        flex-direction: column;
        gap: 15px;
        align-items: stretch;
    }

    .config-actions {
        justify-content: flex-end;
    }

    .nav-tab {
        flex: 1;
        text-align: center;
    }
}
```

### Configuration JavaScript
Create `frontend/config.js`:
```javascript
class ConfigurationManager {
    constructor() {
        this.originalConfig = {};
        this.currentConfig = {};
        this.hasUnsavedChanges = false;
        this.init();
    }

    init() {
        this.setupTabNavigation();
        this.setupFormHandlers();
        this.setupValidation();
        this.loadConfiguration();
        this.setupBeforeUnload();
    }

    setupTabNavigation() {
        const tabs = document.querySelectorAll('.nav-tab');
        const contents = document.querySelectorAll('.tab-content');

        tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const targetTab = tab.dataset.tab;

                // Update active tab
                tabs.forEach(t => t.classList.remove('active'));
                tab.classList.add('active');

                // Update active content
                contents.forEach(content => {
                    content.classList.remove('active');
                    if (content.id === `${targetTab}-tab`) {
                        content.classList.add('active');
                    }
                });
            });
        });
    }

    setupFormHandlers() {
        // Save button
        document.getElementById('saveBtn').addEventListener('click', () => {
            this.saveConfiguration();
        });

        // Cancel button
        document.getElementById('cancelBtn').addEventListener('click', () => {
            this.cancelChanges();
        });

        // Reset button
        document.getElementById('resetBtn').addEventListener('click', () => {
            this.resetToDefaults();
        });

        // Test email button
        document.getElementById('testEmailBtn').addEventListener('click', () => {
            this.testEmailConfiguration();
        });

        // Track changes
        const inputs = document.querySelectorAll('input, select, textarea');
        inputs.forEach(input => {
            input.addEventListener('change', () => {
                this.markAsChanged();
                this.validateField(input);
            });
        });
    }

    setupValidation() {
        const validators = {
            'dataRetention': (value) => {
                const num = parseInt(value);
                return num >= 1 && num <= 365 ? null : 'Must be between 1 and 365 days';
            },
            'defaultInterval': (value) => {
                const num = parseInt(value);
                return num >= 5 && num <= 3600 ? null : 'Must be between 5 and 3600 seconds';
            },
            'defaultTimeout': (value) => {
                const num = parseInt(value);
                return num >= 1 && num <= 120 ? null : 'Must be between 1 and 120 seconds';
            },
            'smtpPort': (value) => {
                const num = parseInt(value);
                return num >= 1 && num <= 65535 ? null : 'Must be a valid port number';
            },
            'fromEmail': (value) => {
                const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                return emailRegex.test(value) || value === '' ? null : 'Invalid email format';
            },
            'toEmails': (value) => {
                if (!value.trim()) return null;
                const emails = value.split(',').map(e => e.trim());
                const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                const invalid = emails.find(email => !emailRegex.test(email));
                return invalid ? `Invalid email: ${invalid}` : null;
            }
        };

        this.validators = validators;
    }

    validateField(input) {
        const validator = this.validators[input.name];
        const formGroup = input.closest('.form-group');
        const existingError = formGroup.querySelector('.error-message');

        if (existingError) {
            existingError.remove();
        }

        formGroup.classList.remove('error');

        if (validator) {
            const error = validator(input.value);
            if (error) {
                formGroup.classList.add('error');
                const errorElement = document.createElement('span');
                errorElement.className = 'error-message';
                errorElement.textContent = error;
                input.parentNode.appendChild(errorElement);
                return false;
            }
        }
        return true;
    }

    validateAll() {
        const inputs = document.querySelectorAll('input, select, textarea');
        let isValid = true;

        inputs.forEach(input => {
            if (!this.validateField(input)) {
                isValid = false;
            }
        });

        return isValid;
    }

    async loadConfiguration() {
        try {
            const config = await window.go.main.App.GetConfiguration();
            this.originalConfig = { ...config };
            this.currentConfig = { ...config };
            this.populateForm(config);
        } catch (error) {
            console.error('Failed to load configuration:', error);
            this.showNotification('Failed to load configuration', 'error');
        }
    }

    populateForm(config) {
        Object.keys(config).forEach(key => {
            const element = document.querySelector(`[name="${key}"]`);
            if (element) {
                if (element.type === 'checkbox') {
                    element.checked = config[key];
                } else {
                    element.value = config[key];
                }
            }
        });
    }

    collectFormData() {
        const formData = {};
        const inputs = document.querySelectorAll('input, select, textarea');

        inputs.forEach(input => {
            if (input.name) {
                if (input.type === 'checkbox') {
                    formData[input.name] = input.checked;
                } else if (input.type === 'number') {
                    formData[input.name] = parseInt(input.value) || 0;
                } else {
                    formData[input.name] = input.value;
                }
            }
        });

        return formData;
    }

    async saveConfiguration() {
        if (!this.validateAll()) {
            this.showNotification('Please fix validation errors before saving', 'error');
            return;
        }

        const config = this.collectFormData();

        try {
            await window.go.main.App.SaveConfiguration(config);
            this.originalConfig = { ...config };
            this.currentConfig = { ...config };
            this.hasUnsavedChanges = false;
            this.updateSaveButton();
            this.showNotification('Configuration saved successfully', 'success');
        } catch (error) {
            console.error('Failed to save configuration:', error);
            this.showNotification('Failed to save configuration', 'error');
        }
    }

    cancelChanges() {
        if (this.hasUnsavedChanges) {
            if (confirm('You have unsaved changes. Are you sure you want to cancel?')) {
                this.populateForm(this.originalConfig);
                this.currentConfig = { ...this.originalConfig };
                this.hasUnsavedChanges = false;
                this.updateSaveButton();
            }
        }
    }

    async resetToDefaults() {
        if (confirm('This will reset all settings to their default values. Continue?')) {
            try {
                const defaults = await window.go.main.App.GetDefaultConfiguration();
                this.populateForm(defaults);
                this.markAsChanged();
            } catch (error) {
                console.error('Failed to get default configuration:', error);
                this.showNotification('Failed to load default configuration', 'error');
            }
        }
    }

    async testEmailConfiguration() {
        const emailConfig = {
            smtpServer: document.getElementById('smtpServer').value,
            smtpPort: parseInt(document.getElementById('smtpPort').value),
            smtpUsername: document.getElementById('smtpUsername').value,
            smtpPassword: document.getElementById('smtpPassword').value,
            fromEmail: document.getElementById('fromEmail').value,
            toEmails: document.getElementById('toEmails').value
        };

        try {
            const testBtn = document.getElementById('testEmailBtn');
            testBtn.disabled = true;
            testBtn.textContent = 'Testing...';

            await window.go.main.App.TestEmailConfiguration(emailConfig);
            this.showNotification('Test email sent successfully', 'success');
        } catch (error) {
            console.error('Email test failed:', error);
            this.showNotification(`Email test failed: ${error.message}`, 'error');
        } finally {
            const testBtn = document.getElementById('testEmailBtn');
            testBtn.disabled = false;
            testBtn.textContent = 'Test Email';
        }
    }

    markAsChanged() {
        this.hasUnsavedChanges = true;
        this.updateSaveButton();
    }

    updateSaveButton() {
        const saveBtn = document.getElementById('saveBtn');
        saveBtn.disabled = !this.hasUnsavedChanges;
        saveBtn.textContent = this.hasUnsavedChanges ? 'Save Changes *' : 'Save Changes';
    }

    setupBeforeUnload() {
        window.addEventListener('beforeunload', (e) => {
            if (this.hasUnsavedChanges) {
                e.preventDefault();
                e.returnValue = 'You have unsaved changes. Are you sure you want to leave?';
            }
        });
    }

    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;

        // Add to page
        document.body.appendChild(notification);

        // Auto-remove after 3 seconds
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new ConfigurationManager();
});
```

### Backend Configuration Handler
Update `app.go` to include configuration methods:
```go
// Configuration structure
type Configuration struct {
    // General settings
    AppName          string `json:"appName"`
    AutoStart        bool   `json:"autoStart"`
    MinimizeToTray   bool   `json:"minimizeToTray"`
    Theme            string `json:"theme"`
    DataRetention    int    `json:"dataRetention"`
    LogLevel         string `json:"logLevel"`

    // Monitoring settings
    DefaultInterval    int  `json:"defaultInterval"`
    DefaultTimeout     int  `json:"defaultTimeout"`
    MaxConcurrent      int  `json:"maxConcurrent"`
    ResponseThreshold  int  `json:"responseThreshold"`
    FailureThreshold   int  `json:"failureThreshold"`
    RecoveryThreshold  int  `json:"recoveryThreshold"`

    // Notification settings
    EnableSystemNotifications bool   `json:"enableSystemNotifications"`
    NotifyOnFailure          bool   `json:"notifyOnFailure"`
    NotifyOnRecovery         bool   `json:"notifyOnRecovery"`
    EnableEmail              bool   `json:"enableEmail"`
    SMTPServer              string `json:"smtpServer"`
    SMTPPort                int    `json:"smtpPort"`
    SMTPUsername            string `json:"smtpUsername"`
    SMTPPassword            string `json:"smtpPassword"`
    FromEmail               string `json:"fromEmail"`
    ToEmails                string `json:"toEmails"`

    // Advanced settings
    UserAgent        string `json:"userAgent"`
    FollowRedirects  bool   `json:"followRedirects"`
    VerifySSL        bool   `json:"verifySsl"`
    WorkerThreads    int    `json:"workerThreads"`
    CacheSize        int    `json:"cacheSize"`
    DebugMode        bool   `json:"debugMode"`
    APIAccess        bool   `json:"apiAccess"`
}

// GetConfiguration returns the current configuration
func (a *App) GetConfiguration(ctx context.Context) (*Configuration, error) {
    a.configMutex.RLock()
    defer a.configMutex.RUnlock()

    // Return a copy of the current configuration
    config := *a.config
    return &config, nil
}

// SaveConfiguration saves the provided configuration
func (a *App) SaveConfiguration(ctx context.Context, config *Configuration) error {
    a.configMutex.Lock()
    defer a.configMutex.Unlock()

    // Validate configuration
    if err := a.validateConfiguration(config); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }

    // Save to file
    configPath := filepath.Join(a.dataDir, "config.json")
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal configuration: %w", err)
    }

    if err := os.WriteFile(configPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write configuration file: %w", err)
    }

    // Update runtime configuration
    oldConfig := a.config
    a.config = config

    // Apply configuration changes
    if err := a.applyConfigurationChanges(oldConfig, config); err != nil {
        a.config = oldConfig // Rollback on error
        return fmt.Errorf("failed to apply configuration changes: %w", err)
    }

    a.logger.Info("Configuration saved and applied successfully")
    return nil
}

// GetDefaultConfiguration returns the default configuration
func (a *App) GetDefaultConfiguration(ctx context.Context) (*Configuration, error) {
    return &Configuration{
        AppName:                  "NetMonitor",
        AutoStart:               false,
        MinimizeToTray:          true,
        Theme:                   "auto",
        DataRetention:           30,
        LogLevel:                "info",
        DefaultInterval:         60,
        DefaultTimeout:          10,
        MaxConcurrent:           10,
        ResponseThreshold:       5000,
        FailureThreshold:        3,
        RecoveryThreshold:       2,
        EnableSystemNotifications: true,
        NotifyOnFailure:         true,
        NotifyOnRecovery:        true,
        EnableEmail:             false,
        SMTPServer:              "",
        SMTPPort:                587,
        SMTPUsername:            "",
        SMTPPassword:            "",
        FromEmail:               "",
        ToEmails:                "",
        UserAgent:               "NetMonitor/1.0",
        FollowRedirects:         true,
        VerifySSL:               true,
        WorkerThreads:           4,
        CacheSize:               100,
        DebugMode:               false,
        APIAccess:               false,
    }, nil
}

// TestEmailConfiguration tests the email configuration
func (a *App) TestEmailConfiguration(ctx context.Context, emailConfig map[string]interface{}) error {
    // Extract email configuration
    smtpServer, _ := emailConfig["smtpServer"].(string)
    smtpPort, _ := emailConfig["smtpPort"].(float64)
    username, _ := emailConfig["smtpUsername"].(string)
    password, _ := emailConfig["smtpPassword"].(string)
    fromEmail, _ := emailConfig["fromEmail"].(string)
    toEmails, _ := emailConfig["toEmails"].(string)

    if smtpServer == "" || fromEmail == "" || toEmails == "" {
        return fmt.Errorf("SMTP server, from email, and to emails are required")
    }

    // Create test email
    toList := strings.Split(toEmails, ",")
    for i, email := range toList {
        toList[i] = strings.TrimSpace(email)
    }

    message := fmt.Sprintf(`Subject: NetMonitor Test Email
From: %s
To: %s
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8

This is a test email from NetMonitor configuration.

If you received this email, your email configuration is working correctly.

Sent at: %s
`, fromEmail, toEmails, time.Now().Format(time.RFC3339))

    // Send test email
    auth := smtp.PlainAuth("", username, password, smtpServer)
    addr := fmt.Sprintf("%s:%d", smtpServer, int(smtpPort))

    err := smtp.SendMail(addr, auth, fromEmail, toList, []byte(message))
    if err != nil {
        return fmt.Errorf("failed to send test email: %w", err)
    }

    a.logger.Info("Test email sent successfully", "to", toEmails)
    return nil
}

// validateConfiguration validates the configuration values
func (a *App) validateConfiguration(config *Configuration) error {
    if config.DataRetention < 1 || config.DataRetention > 365 {
        return fmt.Errorf("data retention must be between 1 and 365 days")
    }

    if config.DefaultInterval < 5 || config.DefaultInterval > 3600 {
        return fmt.Errorf("default interval must be between 5 and 3600 seconds")
    }

    if config.DefaultTimeout < 1 || config.DefaultTimeout > 120 {
        return fmt.Errorf("default timeout must be between 1 and 120 seconds")
    }

    if config.EnableEmail {
        if config.SMTPServer == "" {
            return fmt.Errorf("SMTP server is required when email notifications are enabled")
        }
        if config.FromEmail == "" {
            return fmt.Errorf("from email is required when email notifications are enabled")
        }
        // Basic email validation
        emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
        if !emailRegex.MatchString(config.FromEmail) {
            return fmt.Errorf("invalid from email format")
        }
    }

    return nil
}

// applyConfigurationChanges applies configuration changes that require runtime updates
func (a *App) applyConfigurationChanges(oldConfig, newConfig *Configuration) error {
    // Update logger level if changed
    if oldConfig.LogLevel != newConfig.LogLevel {
        if err := a.updateLogLevel(newConfig.LogLevel); err != nil {
            return fmt.Errorf("failed to update log level: %w", err)
        }
    }

    // Update auto-start setting if changed
    if oldConfig.AutoStart != newConfig.AutoStart {
        if err := a.updateAutoStart(newConfig.AutoStart); err != nil {
            return fmt.Errorf("failed to update auto-start setting: %w", err)
        }
    }

    // Restart monitoring if intervals changed
    if oldConfig.DefaultInterval != newConfig.DefaultInterval ||
       oldConfig.DefaultTimeout != newConfig.DefaultTimeout ||
       oldConfig.MaxConcurrent != newConfig.MaxConcurrent {
        if err := a.restartMonitoring(); err != nil {
            return fmt.Errorf("failed to restart monitoring with new settings: %w", err)
        }
    }

    return nil
}
```

## Verification Steps
1. Open the configuration interface from the main menu
2. Verify all tabs are accessible and contain the expected settings
3. Test form validation by entering invalid values
4. Modify settings and verify the save button becomes enabled
5. Save changes and verify they persist after restarting the application
6. Test the email configuration with the test email feature
7. Reset settings to defaults and verify they are restored
8. Test keyboard navigation through the form
9. Verify responsive behavior on different screen sizes
10. Test unsaved changes warning when navigating away

## Dependencies
- T003: Application Structure (for basic app framework)
- T004: Configuration Management (for configuration infrastructure)
- T041: System Notifications (for notification settings)
- T043: Email Notifications (for email configuration)

## Notes
- The configuration UI should be accessible from the main application menu
- All configuration changes should be validated before saving
- The interface should provide clear feedback for validation errors
- Settings should be organized logically with helpful descriptions
- The test email feature helps users verify their SMTP configuration
- Keyboard navigation support improves accessibility
- Responsive design ensures usability across different screen sizes