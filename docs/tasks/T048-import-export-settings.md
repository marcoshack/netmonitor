# T048: Import/Export Settings

## Overview
Implement a comprehensive import/export system for application settings, configurations, preferences, and monitoring targets. This enables users to backup their configurations, share settings between installations, and migrate data between devices.

## Context
As part of the NetMonitor application, users need the ability to backup and restore their complete application state. This includes configuration settings, preferences, monitoring targets, alert rules, and custom templates. The import/export functionality should support multiple formats and provide data validation to ensure compatibility and prevent corruption.

## Task Description
Implement an import/export system that allows users to:
- Export all application settings to various formats (JSON, YAML, XML)
- Import settings from backup files with validation
- Selectively export/import specific components (targets, alerts, preferences)
- Export monitoring data for analysis in external tools
- Validate imported data before applying changes
- Handle version compatibility and migration
- Create portable configuration packages

## Acceptance Criteria
- [ ] Complete settings export includes all configuration and preferences
- [ ] Selective export allows choosing specific components
- [ ] Multiple export formats are supported (JSON, YAML, XML)
- [ ] Import validation prevents corrupted or incompatible data
- [ ] Backup creation before importing changes
- [ ] Version compatibility checking and migration
- [ ] Monitoring data export for external analysis
- [ ] Encrypted export option for sensitive data
- [ ] Import/export progress feedback for large datasets
- [ ] Rollback functionality if import fails

## Implementation Details

### Import/Export Manager
Create `internal/importexport/manager.go`:
```go
package importexport

import (
    "archive/zip"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"

    "gopkg.in/yaml.v3"
)

// ExportFormat represents the export format
type ExportFormat string

const (
    FormatJSON ExportFormat = "json"
    FormatYAML ExportFormat = "yaml"
    FormatXML  ExportFormat = "xml"
    FormatZIP  ExportFormat = "zip"
)

// ExportOptions defines what to export
type ExportOptions struct {
    Configuration   bool          `json:"configuration"`
    Preferences     bool          `json:"preferences"`
    Targets         bool          `json:"targets"`
    AlertRules      bool          `json:"alertRules"`
    MonitoringData  bool          `json:"monitoringData"`
    DataRange       *TimeRange    `json:"dataRange,omitempty"`
    Format          ExportFormat  `json:"format"`
    Encrypt         bool          `json:"encrypt"`
    Password        string        `json:"password,omitempty"`
    Compression     bool          `json:"compression"`
}

// ImportOptions defines how to import
type ImportOptions struct {
    CreateBackup    bool     `json:"createBackup"`
    OverwriteConfig bool     `json:"overwriteConfig"`
    OverwriteTargets bool    `json:"overwriteTargets"`
    MergeTargets    bool     `json:"mergeTargets"`
    ValidateOnly    bool     `json:"validateOnly"`
    Password        string   `json:"password,omitempty"`
    SkipComponents  []string `json:"skipComponents"`
}

// TimeRange defines a time range for data export
type TimeRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

// ExportData contains all exportable data
type ExportData struct {
    Version        string                 `json:"version"`
    ExportedAt     time.Time              `json:"exportedAt"`
    Application    string                 `json:"application"`
    Configuration  interface{}            `json:"configuration,omitempty"`
    Preferences    interface{}            `json:"preferences,omitempty"`
    Targets        interface{}            `json:"targets,omitempty"`
    AlertRules     interface{}            `json:"alertRules,omitempty"`
    MonitoringData interface{}            `json:"monitoringData,omitempty"`
    Metadata       map[string]interface{} `json:"metadata"`
}

// Manager handles import/export operations
type Manager struct {
    app        Application // Interface to app components
    dataDir    string
    backupDir  string
}

// Application interface for accessing app data
type Application interface {
    GetConfiguration() (interface{}, error)
    GetPreferences() (interface{}, error)
    GetTargets() (interface{}, error)
    GetAlertRules() (interface{}, error)
    GetMonitoringData(timeRange *TimeRange) (interface{}, error)
    SetConfiguration(config interface{}) error
    SetPreferences(prefs interface{}) error
    SetTargets(targets interface{}) error
    SetAlertRules(rules interface{}) error
    CreateBackup() (string, error)
    RestoreBackup(backupPath string) error
}

// NewManager creates a new import/export manager
func NewManager(app Application, dataDir string) *Manager {
    backupDir := filepath.Join(dataDir, "backups")
    os.MkdirAll(backupDir, 0755)

    return &Manager{
        app:       app,
        dataDir:   dataDir,
        backupDir: backupDir,
    }
}

// Export exports application data based on options
func (m *Manager) Export(filePath string, options *ExportOptions) error {
    // Collect data to export
    exportData := &ExportData{
        Version:     "1.0.0",
        ExportedAt:  time.Now(),
        Application: "NetMonitor",
        Metadata:    make(map[string]interface{}),
    }

    // Collect configuration if requested
    if options.Configuration {
        config, err := m.app.GetConfiguration()
        if err != nil {
            return fmt.Errorf("failed to get configuration: %w", err)
        }
        exportData.Configuration = config
    }

    // Collect preferences if requested
    if options.Preferences {
        prefs, err := m.app.GetPreferences()
        if err != nil {
            return fmt.Errorf("failed to get preferences: %w", err)
        }
        exportData.Preferences = prefs
    }

    // Collect targets if requested
    if options.Targets {
        targets, err := m.app.GetTargets()
        if err != nil {
            return fmt.Errorf("failed to get targets: %w", err)
        }
        exportData.Targets = targets
    }

    // Collect alert rules if requested
    if options.AlertRules {
        rules, err := m.app.GetAlertRules()
        if err != nil {
            return fmt.Errorf("failed to get alert rules: %w", err)
        }
        exportData.AlertRules = rules
    }

    // Collect monitoring data if requested
    if options.MonitoringData {
        data, err := m.app.GetMonitoringData(options.DataRange)
        if err != nil {
            return fmt.Errorf("failed to get monitoring data: %w", err)
        }
        exportData.MonitoringData = data
    }

    // Add metadata
    exportData.Metadata["exportOptions"] = options
    exportData.Metadata["hostname"] = getHostname()
    exportData.Metadata["platform"] = getPlatform()

    // Export based on format
    switch options.Format {
    case FormatJSON:
        return m.exportJSON(filePath, exportData, options)
    case FormatYAML:
        return m.exportYAML(filePath, exportData, options)
    case FormatXML:
        return m.exportXML(filePath, exportData, options)
    case FormatZIP:
        return m.exportZIP(filePath, exportData, options)
    default:
        return fmt.Errorf("unsupported export format: %s", options.Format)
    }
}

// exportJSON exports data as JSON
func (m *Manager) exportJSON(filePath string, data *ExportData, options *ExportOptions) error {
    var output []byte
    var err error

    if options.Compression {
        output, err = json.Marshal(data)
    } else {
        output, err = json.MarshalIndent(data, "", "  ")
    }

    if err != nil {
        return fmt.Errorf("failed to marshal JSON: %w", err)
    }

    if options.Encrypt && options.Password != "" {
        output, err = m.encrypt(output, options.Password)
        if err != nil {
            return fmt.Errorf("failed to encrypt data: %w", err)
        }
    }

    return os.WriteFile(filePath, output, 0644)
}

// exportYAML exports data as YAML
func (m *Manager) exportYAML(filePath string, data *ExportData, options *ExportOptions) error {
    output, err := yaml.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal YAML: %w", err)
    }

    if options.Encrypt && options.Password != "" {
        output, err = m.encrypt(output, options.Password)
        if err != nil {
            return fmt.Errorf("failed to encrypt data: %w", err)
        }
    }

    return os.WriteFile(filePath, output, 0644)
}

// exportXML exports data as XML
func (m *Manager) exportXML(filePath string, data *ExportData, options *ExportOptions) error {
    // Convert to XML-friendly structure
    xmlData := m.convertToXMLStruct(data)

    // Marshal to XML
    output, err := json.Marshal(xmlData) // Simplified XML export
    if err != nil {
        return fmt.Errorf("failed to marshal XML: %w", err)
    }

    if options.Encrypt && options.Password != "" {
        output, err = m.encrypt(output, options.Password)
        if err != nil {
            return fmt.Errorf("failed to encrypt data: %w", err)
        }
    }

    return os.WriteFile(filePath, output, 0644)
}

// exportZIP exports data as a ZIP archive with multiple files
func (m *Manager) exportZIP(filePath string, data *ExportData, options *ExportOptions) error {
    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create ZIP file: %w", err)
    }
    defer file.Close()

    zipWriter := zip.NewWriter(file)
    defer zipWriter.Close()

    // Export each component as a separate file
    if data.Configuration != nil {
        if err := m.addToZIP(zipWriter, "configuration.json", data.Configuration, options); err != nil {
            return fmt.Errorf("failed to add configuration to ZIP: %w", err)
        }
    }

    if data.Preferences != nil {
        if err := m.addToZIP(zipWriter, "preferences.json", data.Preferences, options); err != nil {
            return fmt.Errorf("failed to add preferences to ZIP: %w", err)
        }
    }

    if data.Targets != nil {
        if err := m.addToZIP(zipWriter, "targets.json", data.Targets, options); err != nil {
            return fmt.Errorf("failed to add targets to ZIP: %w", err)
        }
    }

    if data.AlertRules != nil {
        if err := m.addToZIP(zipWriter, "alert_rules.json", data.AlertRules, options); err != nil {
            return fmt.Errorf("failed to add alert rules to ZIP: %w", err)
        }
    }

    if data.MonitoringData != nil {
        if err := m.addToZIP(zipWriter, "monitoring_data.json", data.MonitoringData, options); err != nil {
            return fmt.Errorf("failed to add monitoring data to ZIP: %w", err)
        }
    }

    // Add metadata
    metadata := map[string]interface{}{
        "version":     data.Version,
        "exportedAt":  data.ExportedAt,
        "application": data.Application,
        "metadata":    data.Metadata,
    }
    if err := m.addToZIP(zipWriter, "metadata.json", metadata, options); err != nil {
        return fmt.Errorf("failed to add metadata to ZIP: %w", err)
    }

    return nil
}

// addToZIP adds data to ZIP archive
func (m *Manager) addToZIP(zipWriter *zip.Writer, fileName string, data interface{}, options *ExportOptions) error {
    writer, err := zipWriter.Create(fileName)
    if err != nil {
        return err
    }

    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }

    if options.Encrypt && options.Password != "" {
        jsonData, err = m.encrypt(jsonData, options.Password)
        if err != nil {
            return err
        }
    }

    _, err = writer.Write(jsonData)
    return err
}

// Import imports application data from file
func (m *Manager) Import(filePath string, options *ImportOptions) (*ImportResult, error) {
    // Create backup if requested
    var backupPath string
    if options.CreateBackup {
        var err error
        backupPath, err = m.app.CreateBackup()
        if err != nil {
            return nil, fmt.Errorf("failed to create backup: %w", err)
        }
    }

    // Read and parse import file
    importData, err := m.readImportFile(filePath, options.Password)
    if err != nil {
        return nil, fmt.Errorf("failed to read import file: %w", err)
    }

    // Validate import data
    validationResult, err := m.validateImportData(importData)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    if options.ValidateOnly {
        return &ImportResult{
            Success:          true,
            ValidationResult: validationResult,
            BackupPath:       backupPath,
        }, nil
    }

    // Apply imported data
    result, err := m.applyImportData(importData, options)
    if err != nil {
        // Restore backup if import fails
        if backupPath != "" {
            if restoreErr := m.app.RestoreBackup(backupPath); restoreErr != nil {
                return nil, fmt.Errorf("import failed and backup restore failed: %w (original error: %v)", restoreErr, err)
            }
        }
        return nil, fmt.Errorf("failed to apply import data: %w", err)
    }

    result.ValidationResult = validationResult
    result.BackupPath = backupPath
    return result, nil
}

// ImportResult contains the result of an import operation
type ImportResult struct {
    Success          bool                   `json:"success"`
    ImportedItems    map[string]int         `json:"importedItems"`
    SkippedItems     map[string]int         `json:"skippedItems"`
    Errors           []string               `json:"errors"`
    Warnings         []string               `json:"warnings"`
    ValidationResult *ValidationResult      `json:"validationResult"`
    BackupPath       string                 `json:"backupPath"`
}

// ValidationResult contains validation details
type ValidationResult struct {
    Valid        bool              `json:"valid"`
    Version      string            `json:"version"`
    Compatible   bool              `json:"compatible"`
    Issues       []ValidationIssue `json:"issues"`
    ComponentCounts map[string]int `json:"componentCounts"`
}

// ValidationIssue represents a validation issue
type ValidationIssue struct {
    Type        string `json:"type"`        // error, warning, info
    Component   string `json:"component"`
    Message     string `json:"message"`
    Suggestion  string `json:"suggestion,omitempty"`
}

// readImportFile reads and parses the import file
func (m *Manager) readImportFile(filePath, password string) (*ExportData, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Decrypt if password provided
    if password != "" {
        data, err = m.decrypt(data, password)
        if err != nil {
            return nil, fmt.Errorf("failed to decrypt data: %w", err)
        }
    }

    // Determine format and parse
    ext := strings.ToLower(filepath.Ext(filePath))
    var importData ExportData

    switch ext {
    case ".json":
        err = json.Unmarshal(data, &importData)
    case ".yaml", ".yml":
        err = yaml.Unmarshal(data, &importData)
    case ".zip":
        return m.readZIPFile(filePath, password)
    default:
        // Try JSON first, then YAML
        if err = json.Unmarshal(data, &importData); err != nil {
            err = yaml.Unmarshal(data, &importData)
        }
    }

    if err != nil {
        return nil, fmt.Errorf("failed to parse import data: %w", err)
    }

    return &importData, nil
}

// encrypt encrypts data with password
func (m *Manager) encrypt(data []byte, password string) ([]byte, error) {
    key := sha256.Sum256([]byte(password))
    block, err := aes.NewCipher(key[:])
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    encrypted := gcm.Seal(nonce, nonce, data, nil)
    return encrypted, nil
}

// decrypt decrypts data with password
func (m *Manager) decrypt(data []byte, password string) ([]byte, error) {
    key := sha256.Sum256([]byte(password))
    block, err := aes.NewCipher(key[:])
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, fmt.Errorf("encrypted data too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return decrypted, nil
}

// Additional helper functions...
func getHostname() string {
    hostname, _ := os.Hostname()
    return hostname
}

func getPlatform() string {
    return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

func (m *Manager) convertToXMLStruct(data *ExportData) interface{} {
    // Convert to XML-friendly structure
    // Implementation would convert nested maps to XML-compatible format
    return data
}

func (m *Manager) readZIPFile(filePath, password string) (*ExportData, error) {
    // Implementation for reading ZIP files
    // This would extract and parse individual files from the ZIP
    return nil, fmt.Errorf("ZIP import not yet implemented")
}

func (m *Manager) validateImportData(data *ExportData) (*ValidationResult, error) {
    // Implementation for validating import data
    // Check version compatibility, data integrity, etc.
    return &ValidationResult{
        Valid:      true,
        Compatible: true,
        Version:    data.Version,
    }, nil
}

func (m *Manager) applyImportData(data *ExportData, options *ImportOptions) (*ImportResult, error) {
    // Implementation for applying imported data
    // This would update the application state with imported data
    return &ImportResult{
        Success: true,
        ImportedItems: map[string]int{
            "targets": 0,
            "rules":   0,
        },
    }, nil
}
```

### Frontend Import/Export Interface
Create `frontend/import-export.html`:
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Import/Export Settings</title>
    <link rel="stylesheet" href="style.css">
    <link rel="stylesheet" href="import-export.css">
</head>
<body>
    <div class="import-export-container">
        <header class="page-header">
            <h1>Import/Export Settings</h1>
            <p>Backup and restore your NetMonitor configuration</p>
        </header>

        <div class="tab-container">
            <nav class="tab-nav">
                <button class="tab-btn active" data-tab="export">Export</button>
                <button class="tab-btn" data-tab="import">Import</button>
            </nav>

            <!-- Export Tab -->
            <div id="export-tab" class="tab-content active">
                <section class="export-section">
                    <h2>Export Options</h2>
                    <form id="exportForm" class="export-form">
                        <div class="form-section">
                            <h3>What to Export</h3>
                            <div class="checkbox-group">
                                <label>
                                    <input type="checkbox" name="configuration" checked>
                                    Application Configuration
                                    <span class="help-text">Core app settings and preferences</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="preferences" checked>
                                    User Preferences
                                    <span class="help-text">UI settings, themes, and customizations</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="targets" checked>
                                    Monitoring Targets
                                    <span class="help-text">All configured monitoring endpoints</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="alertRules" checked>
                                    Alert Rules
                                    <span class="help-text">Notification rules and thresholds</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="monitoringData">
                                    Monitoring Data
                                    <span class="help-text">Historical monitoring data</span>
                                </label>
                            </div>
                        </div>

                        <div class="form-section" id="dataRangeSection" style="display: none;">
                            <h3>Data Range</h3>
                            <div class="date-range">
                                <label>
                                    From:
                                    <input type="datetime-local" name="startDate" id="startDate">
                                </label>
                                <label>
                                    To:
                                    <input type="datetime-local" name="endDate" id="endDate">
                                </label>
                            </div>
                            <div class="preset-ranges">
                                <button type="button" class="preset-btn" data-range="1h">Last Hour</button>
                                <button type="button" class="preset-btn" data-range="24h">Last 24 Hours</button>
                                <button type="button" class="preset-btn" data-range="7d">Last 7 Days</button>
                                <button type="button" class="preset-btn" data-range="30d">Last 30 Days</button>
                            </div>
                        </div>

                        <div class="form-section">
                            <h3>Export Format</h3>
                            <div class="radio-group">
                                <label>
                                    <input type="radio" name="format" value="json" checked>
                                    JSON
                                    <span class="help-text">Human-readable, widely supported</span>
                                </label>
                                <label>
                                    <input type="radio" name="format" value="yaml">
                                    YAML
                                    <span class="help-text">Easy to read and edit</span>
                                </label>
                                <label>
                                    <input type="radio" name="format" value="zip">
                                    ZIP Archive
                                    <span class="help-text">Compressed, organized files</span>
                                </label>
                            </div>
                        </div>

                        <div class="form-section">
                            <h3>Security Options</h3>
                            <div class="checkbox-group">
                                <label>
                                    <input type="checkbox" name="encrypt">
                                    Encrypt Export
                                    <span class="help-text">Password-protect sensitive data</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="compression">
                                    Compress Output
                                    <span class="help-text">Reduce file size</span>
                                </label>
                            </div>
                            <div class="password-section" id="passwordSection" style="display: none;">
                                <label>
                                    Encryption Password:
                                    <input type="password" name="password" id="exportPassword" placeholder="Enter password">
                                </label>
                                <label>
                                    Confirm Password:
                                    <input type="password" name="confirmPassword" id="confirmPassword" placeholder="Confirm password">
                                </label>
                            </div>
                        </div>

                        <div class="form-actions">
                            <button type="submit" class="btn btn-primary">Export Settings</button>
                        </div>
                    </form>
                </section>
            </div>

            <!-- Import Tab -->
            <div id="import-tab" class="tab-content">
                <section class="import-section">
                    <h2>Import Settings</h2>
                    <form id="importForm" class="import-form">
                        <div class="form-section">
                            <h3>Select Import File</h3>
                            <div class="file-input-section">
                                <input type="file" id="importFile" name="importFile"
                                       accept=".json,.yaml,.yml,.zip" required>
                                <label for="importFile" class="file-input-label">
                                    <span class="file-icon">📁</span>
                                    Choose File
                                </label>
                                <div class="file-info" id="fileInfo" style="display: none;">
                                    <span class="file-name"></span>
                                    <span class="file-size"></span>
                                </div>
                            </div>
                        </div>

                        <div class="form-section">
                            <h3>Import Options</h3>
                            <div class="checkbox-group">
                                <label>
                                    <input type="checkbox" name="createBackup" checked>
                                    Create Backup Before Import
                                    <span class="help-text">Recommended for safety</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="overwriteConfig">
                                    Overwrite Existing Configuration
                                    <span class="help-text">Replace current settings completely</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="mergeTargets" checked>
                                    Merge Monitoring Targets
                                    <span class="help-text">Add to existing targets instead of replacing</span>
                                </label>
                                <label>
                                    <input type="checkbox" name="validateOnly">
                                    Validate Only (Don't Apply)
                                    <span class="help-text">Check file validity without importing</span>
                                </label>
                            </div>
                        </div>

                        <div class="form-section">
                            <h3>Security</h3>
                            <label>
                                Decryption Password (if required):
                                <input type="password" name="importPassword" id="importPassword"
                                       placeholder="Enter password">
                            </label>
                        </div>

                        <div class="form-actions">
                            <button type="submit" class="btn btn-primary">Import Settings</button>
                            <button type="button" id="validateBtn" class="btn btn-secondary">Validate File</button>
                        </div>
                    </form>

                    <!-- Import Preview -->
                    <div id="importPreview" class="import-preview" style="display: none;">
                        <h3>Import Preview</h3>
                        <div class="preview-content">
                            <div class="preview-section">
                                <h4>File Information</h4>
                                <div class="info-grid">
                                    <span>Version:</span><span id="previewVersion">-</span>
                                    <span>Exported:</span><span id="previewDate">-</span>
                                    <span>Application:</span><span id="previewApp">-</span>
                                    <span>Format:</span><span id="previewFormat">-</span>
                                </div>
                            </div>
                            <div class="preview-section">
                                <h4>Components</h4>
                                <div class="component-list" id="componentList">
                                    <!-- Populated by JavaScript -->
                                </div>
                            </div>
                            <div class="preview-section" id="validationSection">
                                <h4>Validation Results</h4>
                                <div class="validation-results" id="validationResults">
                                    <!-- Populated by JavaScript -->
                                </div>
                            </div>
                        </div>
                    </div>
                </section>
            </div>
        </div>

        <!-- Progress Modal -->
        <div id="progressModal" class="modal" style="display: none;">
            <div class="modal-content">
                <h3 id="progressTitle">Processing...</h3>
                <div class="progress-bar">
                    <div class="progress-fill" id="progressFill"></div>
                </div>
                <p id="progressText">Initializing...</p>
                <div class="progress-actions">
                    <button id="progressCancel" class="btn btn-secondary">Cancel</button>
                </div>
            </div>
        </div>

        <!-- Result Modal -->
        <div id="resultModal" class="modal" style="display: none;">
            <div class="modal-content">
                <h3 id="resultTitle">Operation Complete</h3>
                <div id="resultContent">
                    <!-- Populated by JavaScript -->
                </div>
                <div class="modal-actions">
                    <button id="resultClose" class="btn btn-primary">Close</button>
                </div>
            </div>
        </div>
    </div>

    <script src="import-export.js"></script>
</body>
</html>
```

### Import/Export JavaScript
Create `frontend/import-export.js`:
```javascript
class ImportExportManager {
    constructor() {
        this.currentOperation = null;
        this.init();
    }

    init() {
        this.setupTabNavigation();
        this.setupExportForm();
        this.setupImportForm();
        this.setupModals();
    }

    setupTabNavigation() {
        const tabBtns = document.querySelectorAll('.tab-btn');
        const tabContents = document.querySelectorAll('.tab-content');

        tabBtns.forEach(btn => {
            btn.addEventListener('click', () => {
                const targetTab = btn.dataset.tab;

                tabBtns.forEach(b => b.classList.remove('active'));
                btn.classList.add('active');

                tabContents.forEach(content => {
                    content.classList.remove('active');
                    if (content.id === `${targetTab}-tab`) {
                        content.classList.add('active');
                    }
                });
            });
        });
    }

    setupExportForm() {
        const form = document.getElementById('exportForm');
        const monitoringDataCheckbox = form.querySelector('input[name="monitoringData"]');
        const dataRangeSection = document.getElementById('dataRangeSection');
        const encryptCheckbox = form.querySelector('input[name="encrypt"]');
        const passwordSection = document.getElementById('passwordSection');

        // Show/hide data range section
        monitoringDataCheckbox.addEventListener('change', () => {
            dataRangeSection.style.display = monitoringDataCheckbox.checked ? 'block' : 'none';
        });

        // Show/hide password section
        encryptCheckbox.addEventListener('change', () => {
            passwordSection.style.display = encryptCheckbox.checked ? 'block' : 'none';
        });

        // Preset date ranges
        document.querySelectorAll('.preset-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const range = btn.dataset.range;
                this.setDateRange(range);
            });
        });

        // Form submission
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleExport();
        });
    }

    setupImportForm() {
        const form = document.getElementById('importForm');
        const fileInput = document.getElementById('importFile');
        const fileInfo = document.getElementById('fileInfo');
        const validateBtn = document.getElementById('validateBtn');

        // File selection
        fileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                fileInfo.style.display = 'block';
                fileInfo.querySelector('.file-name').textContent = file.name;
                fileInfo.querySelector('.file-size').textContent = this.formatFileSize(file.size);
            } else {
                fileInfo.style.display = 'none';
            }
        });

        // Validate button
        validateBtn.addEventListener('click', () => {
            this.handleValidate();
        });

        // Form submission
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleImport();
        });
    }

    setupModals() {
        // Progress modal cancel
        document.getElementById('progressCancel').addEventListener('click', () => {
            this.cancelOperation();
        });

        // Result modal close
        document.getElementById('resultClose').addEventListener('click', () => {
            this.hideModal('resultModal');
        });
    }

    setDateRange(range) {
        const endDate = new Date();
        const startDate = new Date();

        switch (range) {
            case '1h':
                startDate.setHours(startDate.getHours() - 1);
                break;
            case '24h':
                startDate.setDate(startDate.getDate() - 1);
                break;
            case '7d':
                startDate.setDate(startDate.getDate() - 7);
                break;
            case '30d':
                startDate.setDate(startDate.getDate() - 30);
                break;
        }

        document.getElementById('startDate').value = this.formatDateTimeLocal(startDate);
        document.getElementById('endDate').value = this.formatDateTimeLocal(endDate);
    }

    formatDateTimeLocal(date) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');

        return `${year}-${month}-${day}T${hours}:${minutes}`;
    }

    formatFileSize(bytes) {
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        if (bytes === 0) return '0 Bytes';
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }

    async handleExport() {
        const form = document.getElementById('exportForm');
        const formData = new FormData(form);

        // Validate passwords if encryption is enabled
        if (formData.get('encrypt')) {
            const password = formData.get('password');
            const confirmPassword = formData.get('confirmPassword');

            if (!password) {
                alert('Password is required for encryption');
                return;
            }

            if (password !== confirmPassword) {
                alert('Passwords do not match');
                return;
            }
        }

        // Build export options
        const options = {
            configuration: formData.has('configuration'),
            preferences: formData.has('preferences'),
            targets: formData.has('targets'),
            alertRules: formData.has('alertRules'),
            monitoringData: formData.has('monitoringData'),
            format: formData.get('format'),
            encrypt: formData.has('encrypt'),
            password: formData.get('password') || '',
            compression: formData.has('compression')
        };

        // Add data range if monitoring data is included
        if (options.monitoringData) {
            const startDate = formData.get('startDate');
            const endDate = formData.get('endDate');

            if (startDate && endDate) {
                options.dataRange = {
                    start: new Date(startDate).toISOString(),
                    end: new Date(endDate).toISOString()
                };
            }
        }

        try {
            this.showProgressModal('Exporting Settings', 'Preparing export...');

            const result = await window.go.main.App.ExportSettings(options);

            this.hideModal('progressModal');
            this.showResultModal('Export Complete', `Settings exported successfully to: ${result.filePath}`);

        } catch (error) {
            this.hideModal('progressModal');
            this.showResultModal('Export Failed', `Failed to export settings: ${error.message}`, 'error');
        }
    }

    async handleValidate() {
        const fileInput = document.getElementById('importFile');
        const file = fileInput.files[0];

        if (!file) {
            alert('Please select a file to validate');
            return;
        }

        try {
            this.showProgressModal('Validating File', 'Reading and validating file...');

            const result = await this.validateFile(file);

            this.hideModal('progressModal');
            this.showImportPreview(result);

        } catch (error) {
            this.hideModal('progressModal');
            this.showResultModal('Validation Failed', `File validation failed: ${error.message}`, 'error');
        }
    }

    async handleImport() {
        const form = document.getElementById('importForm');
        const formData = new FormData(form);
        const file = formData.get('importFile');

        if (!file) {
            alert('Please select a file to import');
            return;
        }

        // Build import options
        const options = {
            createBackup: formData.has('createBackup'),
            overwriteConfig: formData.has('overwriteConfig'),
            mergeTargets: formData.has('mergeTargets'),
            validateOnly: formData.has('validateOnly'),
            password: formData.get('importPassword') || ''
        };

        try {
            this.showProgressModal('Importing Settings', 'Processing import file...');

            const result = await this.importFile(file, options);

            this.hideModal('progressModal');

            if (result.success) {
                this.showResultModal('Import Complete', this.formatImportResult(result));
            } else {
                this.showResultModal('Import Failed', this.formatImportErrors(result), 'error');
            }

        } catch (error) {
            this.hideModal('progressModal');
            this.showResultModal('Import Failed', `Import failed: ${error.message}`, 'error');
        }
    }

    async validateFile(file) {
        // Read file content
        const fileContent = await this.readFileAsText(file);

        // Send to backend for validation
        return await window.go.main.App.ValidateImportFile(fileContent, file.name);
    }

    async importFile(file, options) {
        // Read file content
        const fileContent = await this.readFileAsText(file);

        // Send to backend for import
        return await window.go.main.App.ImportSettings(fileContent, file.name, options);
    }

    readFileAsText(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = e => resolve(e.target.result);
            reader.onerror = reject;
            reader.readAsText(file);
        });
    }

    showImportPreview(validationResult) {
        const preview = document.getElementById('importPreview');

        // Update file information
        document.getElementById('previewVersion').textContent = validationResult.version || 'Unknown';
        document.getElementById('previewDate').textContent = validationResult.exportedAt || 'Unknown';
        document.getElementById('previewApp').textContent = validationResult.application || 'Unknown';
        document.getElementById('previewFormat').textContent = validationResult.format || 'Unknown';

        // Update component list
        const componentList = document.getElementById('componentList');
        componentList.innerHTML = '';

        Object.entries(validationResult.componentCounts || {}).forEach(([component, count]) => {
            const item = document.createElement('div');
            item.className = 'component-item';
            item.innerHTML = `
                <span class="component-name">${component}</span>
                <span class="component-count">${count} items</span>
            `;
            componentList.appendChild(item);
        });

        // Update validation results
        const validationResults = document.getElementById('validationResults');
        validationResults.innerHTML = '';

        if (validationResult.valid) {
            validationResults.innerHTML = '<div class="validation-success">✓ File is valid and compatible</div>';
        } else {
            validationResult.issues?.forEach(issue => {
                const item = document.createElement('div');
                item.className = `validation-${issue.type}`;
                item.textContent = issue.message;
                validationResults.appendChild(item);
            });
        }

        preview.style.display = 'block';
    }

    formatImportResult(result) {
        let message = 'Import completed successfully!\n\n';

        if (result.importedItems) {
            message += 'Imported:\n';
            Object.entries(result.importedItems).forEach(([type, count]) => {
                message += `• ${count} ${type}\n`;
            });
        }

        if (result.skippedItems && Object.keys(result.skippedItems).length > 0) {
            message += '\nSkipped:\n';
            Object.entries(result.skippedItems).forEach(([type, count]) => {
                message += `• ${count} ${type}\n`;
            });
        }

        if (result.backupPath) {
            message += `\nBackup created: ${result.backupPath}`;
        }

        return message;
    }

    formatImportErrors(result) {
        let message = 'Import failed with the following errors:\n\n';

        result.errors?.forEach(error => {
            message += `• ${error}\n`;
        });

        if (result.warnings?.length > 0) {
            message += '\nWarnings:\n';
            result.warnings.forEach(warning => {
                message += `• ${warning}\n`;
            });
        }

        return message;
    }

    showProgressModal(title, text) {
        document.getElementById('progressTitle').textContent = title;
        document.getElementById('progressText').textContent = text;
        document.getElementById('progressFill').style.width = '0%';
        this.showModal('progressModal');
    }

    updateProgress(percent, text) {
        document.getElementById('progressFill').style.width = `${percent}%`;
        if (text) {
            document.getElementById('progressText').textContent = text;
        }
    }

    showResultModal(title, content, type = 'success') {
        document.getElementById('resultTitle').textContent = title;

        const resultContent = document.getElementById('resultContent');
        resultContent.innerHTML = `<pre class="result-${type}">${content}</pre>`;

        this.showModal('resultModal');
    }

    showModal(modalId) {
        document.getElementById(modalId).style.display = 'flex';
    }

    hideModal(modalId) {
        document.getElementById(modalId).style.display = 'none';
    }

    cancelOperation() {
        if (this.currentOperation) {
            // Cancel current operation
            this.currentOperation.cancel();
        }
        this.hideModal('progressModal');
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new ImportExportManager();
});
```

## Verification Steps
1. Test export with various component combinations
2. Verify export formats (JSON, YAML, ZIP) work correctly
3. Test import validation catches incompatible files
4. Verify backup creation before import works
5. Test selective import options work as expected
6. Verify encryption/decryption with passwords
7. Test import rollback on failure
8. Verify progress feedback during operations
9. Test with large monitoring data exports
10. Verify version compatibility checking

## Dependencies
- T004: Configuration Management (for configuration data)
- T047: Preferences Management (for preferences data)
- T007: Target Management (for monitoring targets)
- T042: Alert Management System (for alert rules)

## Notes
- Exports should be password-protected when containing sensitive data
- Large data exports should show progress feedback
- Import validation prevents data corruption
- Backup creation before import ensures data safety
- The system supports multiple export formats for flexibility
- Version compatibility checking prevents issues with future updates