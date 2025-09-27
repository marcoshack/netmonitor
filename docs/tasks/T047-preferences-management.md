# T047: Preferences Management

## Overview
Implement a robust preferences management system that handles user preferences, application state persistence, theme management, and user interface customizations beyond basic configuration settings.

## Context
As part of the NetMonitor application, users need a way to customize their experience beyond the core configuration settings. This includes UI preferences, window states, dashboard layouts, display options, and personal customizations. The preferences management system should handle automatic saving, loading, and synchronization of these settings across application sessions.

## Task Description
Implement a comprehensive preferences management system that:
- Manages user interface preferences and customizations
- Handles window state persistence (position, size, layout)
- Manages theme settings and customizations
- Stores dashboard layout preferences
- Handles view preferences (chart types, time ranges, etc.)
- Automatically saves and restores user preferences
- Provides preference migration for version updates

## Acceptance Criteria
- [ ] Preferences are automatically saved when changed
- [ ] Window position and size are restored on application restart
- [ ] Dashboard layout preferences are persisted
- [ ] Theme settings are applied on application startup
- [ ] View preferences (chart types, filters) are remembered
- [ ] Preferences file structure supports future additions
- [ ] Preference migration handles version updates gracefully
- [ ] Invalid or corrupted preferences are handled safely
- [ ] Preferences can be reset to defaults
- [ ] Export/import preparation is included for T049

## Implementation Details

### Preferences Structure
Create `internal/preferences/preferences.go`:
```go
package preferences

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// Preferences holds all user preferences
type Preferences struct {
    Version    string          `json:"version"`
    UpdatedAt  time.Time       `json:"updatedAt"`
    Window     *WindowPrefs    `json:"window"`
    Theme      *ThemePrefs     `json:"theme"`
    Dashboard  *DashboardPrefs `json:"dashboard"`
    Charts     *ChartPrefs     `json:"charts"`
    Filters    *FilterPrefs    `json:"filters"`
    UI         *UIPrefs        `json:"ui"`
}

// WindowPrefs holds window-related preferences
type WindowPrefs struct {
    X          int  `json:"x"`
    Y          int  `json:"y"`
    Width      int  `json:"width"`
    Height     int  `json:"height"`
    Maximized  bool `json:"maximized"`
    Minimized  bool `json:"minimized"`
    AlwaysOnTop bool `json:"alwaysOnTop"`
}

// ThemePrefs holds theme-related preferences
type ThemePrefs struct {
    Mode           string            `json:"mode"`           // light, dark, auto
    AccentColor    string            `json:"accentColor"`
    CustomColors   map[string]string `json:"customColors"`
    FontSize       string            `json:"fontSize"`       // small, medium, large
    HighContrast   bool              `json:"highContrast"`
    ReducedMotion  bool              `json:"reducedMotion"`
}

// DashboardPrefs holds dashboard layout preferences
type DashboardPrefs struct {
    Layout         string                 `json:"layout"`         // grid, list, compact
    WidgetOrder    []string               `json:"widgetOrder"`
    HiddenWidgets  []string               `json:"hiddenWidgets"`
    WidgetSizes    map[string]string      `json:"widgetSizes"`    // small, medium, large
    CustomLayouts  map[string]interface{} `json:"customLayouts"`
    AutoRefresh    int                    `json:"autoRefresh"`    // seconds
    ShowLegends    bool                   `json:"showLegends"`
}

// ChartPrefs holds chart display preferences
type ChartPrefs struct {
    DefaultType       string            `json:"defaultType"`       // line, bar, area
    TimeRange         string            `json:"timeRange"`         // 1h, 6h, 24h, 7d, 30d
    ShowDataPoints    bool              `json:"showDataPoints"`
    SmoothLines       bool              `json:"smoothLines"`
    ShowGrid          bool              `json:"showGrid"`
    AnimationSpeed    string            `json:"animationSpeed"`    // none, slow, normal, fast
    ColorScheme       string            `json:"colorScheme"`
    CustomColors      map[string]string `json:"customColors"`
    ShowTooltips      bool              `json:"showTooltips"`
    ShowZoom          bool              `json:"showZoom"`
}

// FilterPrefs holds filter preferences
type FilterPrefs struct {
    DefaultFilters    map[string]interface{} `json:"defaultFilters"`
    SavedFilters      map[string]interface{} `json:"savedFilters"`
    QuickFilters      []string               `json:"quickFilters"`
    FilterHistory     []string               `json:"filterHistory"`
    AutoApplyFilters  bool                   `json:"autoApplyFilters"`
    RememberFilters   bool                   `json:"rememberFilters"`
}

// UIPrefs holds general UI preferences
type UIPrefs struct {
    SidebarWidth      int                    `json:"sidebarWidth"`
    SidebarCollapsed  bool                   `json:"sidebarCollapsed"`
    ShowStatusBar     bool                   `json:"showStatusBar"`
    ShowToolbar       bool                   `json:"showToolbar"`
    TablePageSize     int                    `json:"tablePageSize"`
    TableDensity      string                 `json:"tableDensity"`      // compact, normal, comfortable
    DateFormat        string                 `json:"dateFormat"`
    TimeFormat        string                 `json:"timeFormat"`
    NumberFormat      string                 `json:"numberFormat"`
    Language          string                 `json:"language"`
    Shortcuts         map[string]string      `json:"shortcuts"`
    RecentTargets     []string               `json:"recentTargets"`
    BookmarkedViews   []string               `json:"bookmarkedViews"`
}

// PreferencesManager handles loading, saving, and managing preferences
type PreferencesManager struct {
    preferences  *Preferences
    filePath     string
    mutex        sync.RWMutex
    autoSave     bool
    saveInterval time.Duration
    stopChan     chan struct{}
}

// NewPreferencesManager creates a new preferences manager
func NewPreferencesManager(dataDir string) (*PreferencesManager, error) {
    filePath := filepath.Join(dataDir, "preferences.json")

    pm := &PreferencesManager{
        filePath:     filePath,
        autoSave:     true,
        saveInterval: 5 * time.Second,
        stopChan:     make(chan struct{}),
    }

    if err := pm.load(); err != nil {
        // If loading fails, use defaults
        pm.preferences = pm.getDefaultPreferences()
        if err := pm.save(); err != nil {
            return nil, fmt.Errorf("failed to save default preferences: %w", err)
        }
    }

    // Start auto-save goroutine
    go pm.autoSaveRoutine()

    return pm, nil
}

// GetPreferences returns a copy of current preferences
func (pm *PreferencesManager) GetPreferences() *Preferences {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    // Return a deep copy
    data, _ := json.Marshal(pm.preferences)
    var copy Preferences
    json.Unmarshal(data, &copy)
    return &copy
}

// UpdatePreferences updates preferences with the provided data
func (pm *PreferencesManager) UpdatePreferences(updates map[string]interface{}) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    // Convert current preferences to map for easier updates
    data, err := json.Marshal(pm.preferences)
    if err != nil {
        return fmt.Errorf("failed to marshal preferences: %w", err)
    }

    var prefMap map[string]interface{}
    if err := json.Unmarshal(data, &prefMap); err != nil {
        return fmt.Errorf("failed to unmarshal preferences: %w", err)
    }

    // Apply updates
    pm.applyUpdates(prefMap, updates)

    // Convert back to struct
    updatedData, err := json.Marshal(prefMap)
    if err != nil {
        return fmt.Errorf("failed to marshal updated preferences: %w", err)
    }

    var newPrefs Preferences
    if err := json.Unmarshal(updatedData, &newPrefs); err != nil {
        return fmt.Errorf("failed to unmarshal updated preferences: %w", err)
    }

    newPrefs.UpdatedAt = time.Now()
    pm.preferences = &newPrefs

    return nil
}

// applyUpdates recursively applies updates to the preference map
func (pm *PreferencesManager) applyUpdates(target map[string]interface{}, updates map[string]interface{}) {
    for key, value := range updates {
        if valueMap, ok := value.(map[string]interface{}); ok {
            if targetMap, exists := target[key].(map[string]interface{}); exists {
                pm.applyUpdates(targetMap, valueMap)
            } else {
                target[key] = valueMap
            }
        } else {
            target[key] = value
        }
    }
}

// SetWindowPreferences updates window-specific preferences
func (pm *PreferencesManager) SetWindowPreferences(prefs *WindowPrefs) error {
    return pm.UpdatePreferences(map[string]interface{}{
        "window": prefs,
    })
}

// SetThemePreferences updates theme-specific preferences
func (pm *PreferencesManager) SetThemePreferences(prefs *ThemePrefs) error {
    return pm.UpdatePreferences(map[string]interface{}{
        "theme": prefs,
    })
}

// SetDashboardPreferences updates dashboard-specific preferences
func (pm *PreferencesManager) SetDashboardPreferences(prefs *DashboardPrefs) error {
    return pm.UpdatePreferences(map[string]interface{}{
        "dashboard": prefs,
    })
}

// Save forces an immediate save of preferences
func (pm *PreferencesManager) Save() error {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    return pm.save()
}

// load loads preferences from file
func (pm *PreferencesManager) load() error {
    data, err := os.ReadFile(pm.filePath)
    if err != nil {
        if os.IsNotExist(err) {
            pm.preferences = pm.getDefaultPreferences()
            return nil
        }
        return fmt.Errorf("failed to read preferences file: %w", err)
    }

    var prefs Preferences
    if err := json.Unmarshal(data, &prefs); err != nil {
        return fmt.Errorf("failed to unmarshal preferences: %w", err)
    }

    // Migrate preferences if needed
    if err := pm.migratePreferences(&prefs); err != nil {
        return fmt.Errorf("failed to migrate preferences: %w", err)
    }

    pm.preferences = &prefs
    return nil
}

// save saves preferences to file
func (pm *PreferencesManager) save() error {
    pm.preferences.UpdatedAt = time.Now()

    data, err := json.MarshalIndent(pm.preferences, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal preferences: %w", err)
    }

    // Write to temporary file first, then rename (atomic operation)
    tempPath := pm.filePath + ".tmp"
    if err := os.WriteFile(tempPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write preferences file: %w", err)
    }

    if err := os.Rename(tempPath, pm.filePath); err != nil {
        os.Remove(tempPath) // Clean up temp file
        return fmt.Errorf("failed to rename preferences file: %w", err)
    }

    return nil
}

// autoSaveRoutine handles automatic saving of preferences
func (pm *PreferencesManager) autoSaveRoutine() {
    ticker := time.NewTicker(pm.saveInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if pm.autoSave {
                pm.Save()
            }
        case <-pm.stopChan:
            return
        }
    }
}

// getDefaultPreferences returns default preferences
func (pm *PreferencesManager) getDefaultPreferences() *Preferences {
    return &Preferences{
        Version:   "1.0.0",
        UpdatedAt: time.Now(),
        Window: &WindowPrefs{
            X:           100,
            Y:           100,
            Width:       1200,
            Height:      800,
            Maximized:   false,
            Minimized:   false,
            AlwaysOnTop: false,
        },
        Theme: &ThemePrefs{
            Mode:          "auto",
            AccentColor:   "#007acc",
            CustomColors:  make(map[string]string),
            FontSize:      "medium",
            HighContrast:  false,
            ReducedMotion: false,
        },
        Dashboard: &DashboardPrefs{
            Layout:        "grid",
            WidgetOrder:   []string{"overview", "targets", "alerts", "performance"},
            HiddenWidgets: []string{},
            WidgetSizes:   make(map[string]string),
            CustomLayouts: make(map[string]interface{}),
            AutoRefresh:   30,
            ShowLegends:   true,
        },
        Charts: &ChartPrefs{
            DefaultType:    "line",
            TimeRange:      "24h",
            ShowDataPoints: true,
            SmoothLines:    true,
            ShowGrid:       true,
            AnimationSpeed: "normal",
            ColorScheme:    "default",
            CustomColors:   make(map[string]string),
            ShowTooltips:   true,
            ShowZoom:       true,
        },
        Filters: &FilterPrefs{
            DefaultFilters:   make(map[string]interface{}),
            SavedFilters:     make(map[string]interface{}),
            QuickFilters:     []string{},
            FilterHistory:    []string{},
            AutoApplyFilters: false,
            RememberFilters:  true,
        },
        UI: &UIPrefs{
            SidebarWidth:     250,
            SidebarCollapsed: false,
            ShowStatusBar:    true,
            ShowToolbar:      true,
            TablePageSize:    25,
            TableDensity:     "normal",
            DateFormat:       "YYYY-MM-DD",
            TimeFormat:       "HH:mm:ss",
            NumberFormat:     "en-US",
            Language:         "en",
            Shortcuts:        getDefaultShortcuts(),
            RecentTargets:    []string{},
            BookmarkedViews:  []string{},
        },
    }
}

// getDefaultShortcuts returns default keyboard shortcuts
func getDefaultShortcuts() map[string]string {
    return map[string]string{
        "new_target":     "Ctrl+N",
        "save":           "Ctrl+S",
        "refresh":        "F5",
        "fullscreen":     "F11",
        "search":         "Ctrl+F",
        "settings":       "Ctrl+,",
        "help":           "F1",
        "toggle_sidebar": "Ctrl+B",
        "zoom_in":        "Ctrl+Plus",
        "zoom_out":       "Ctrl+Minus",
    }
}

// migratePreferences handles preference migration for version updates
func (pm *PreferencesManager) migratePreferences(prefs *Preferences) error {
    currentVersion := "1.0.0"

    if prefs.Version == "" {
        // Migrate from pre-versioned preferences
        prefs.Version = "0.9.0"
    }

    if prefs.Version == "0.9.0" {
        // Migrate to 1.0.0
        if prefs.UI == nil {
            prefs.UI = &UIPrefs{
                SidebarWidth:     250,
                SidebarCollapsed: false,
                ShowStatusBar:    true,
                ShowToolbar:      true,
                TablePageSize:    25,
                TableDensity:     "normal",
                DateFormat:       "YYYY-MM-DD",
                TimeFormat:       "HH:mm:ss",
                NumberFormat:     "en-US",
                Language:         "en",
                Shortcuts:        getDefaultShortcuts(),
                RecentTargets:    []string{},
                BookmarkedViews:  []string{},
            }
        }
        prefs.Version = "1.0.0"
    }

    prefs.Version = currentVersion
    return nil
}

// Reset resets preferences to defaults
func (pm *PreferencesManager) Reset() error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    pm.preferences = pm.getDefaultPreferences()
    return pm.save()
}

// Close stops the preferences manager
func (pm *PreferencesManager) Close() error {
    close(pm.stopChan)
    return pm.Save()
}
```

### Frontend Preferences Handler
Create `frontend/preferences.js`:
```javascript
class PreferencesHandler {
    constructor() {
        this.preferences = {};
        this.observers = new Map();
        this.debounceTimers = new Map();
        this.init();
    }

    async init() {
        try {
            this.preferences = await window.go.main.App.GetPreferences();
            this.applyPreferences();
            this.setupEventListeners();
        } catch (error) {
            console.error('Failed to load preferences:', error);
        }
    }

    // Apply preferences to the UI
    applyPreferences() {
        this.applyThemePreferences();
        this.applyWindowPreferences();
        this.applyUIPreferences();
    }

    applyThemePreferences() {
        const theme = this.preferences.theme;
        if (!theme) return;

        // Set theme mode
        document.documentElement.setAttribute('data-theme', theme.mode);

        // Set accent color
        if (theme.accentColor) {
            document.documentElement.style.setProperty('--accent-color', theme.accentColor);
        }

        // Apply custom colors
        if (theme.customColors) {
            Object.entries(theme.customColors).forEach(([key, value]) => {
                document.documentElement.style.setProperty(`--${key}`, value);
            });
        }

        // Set font size
        if (theme.fontSize) {
            document.documentElement.setAttribute('data-font-size', theme.fontSize);
        }

        // High contrast mode
        if (theme.highContrast) {
            document.documentElement.classList.add('high-contrast');
        }

        // Reduced motion
        if (theme.reducedMotion) {
            document.documentElement.classList.add('reduced-motion');
        }
    }

    applyWindowPreferences() {
        const window = this.preferences.window;
        if (!window) return;

        // Window preferences are typically handled by the backend
        // But we can notify the backend of any changes needed
    }

    applyUIPreferences() {
        const ui = this.preferences.ui;
        if (!ui) return;

        // Sidebar width
        if (ui.sidebarWidth) {
            document.documentElement.style.setProperty('--sidebar-width', `${ui.sidebarWidth}px`);
        }

        // Sidebar collapsed state
        if (ui.sidebarCollapsed) {
            document.body.classList.add('sidebar-collapsed');
        }

        // Status bar
        const statusBar = document.querySelector('.status-bar');
        if (statusBar) {
            statusBar.style.display = ui.showStatusBar ? 'block' : 'none';
        }

        // Toolbar
        const toolbar = document.querySelector('.toolbar');
        if (toolbar) {
            toolbar.style.display = ui.showToolbar ? 'block' : 'none';
        }

        // Table density
        if (ui.tableDensity) {
            document.documentElement.setAttribute('data-table-density', ui.tableDensity);
        }
    }

    // Update preferences
    async updatePreferences(updates) {
        try {
            await window.go.main.App.UpdatePreferences(updates);

            // Update local copy
            this.mergeUpdates(this.preferences, updates);

            // Apply changes
            this.applyPreferences();

            // Notify observers
            this.notifyObservers(updates);
        } catch (error) {
            console.error('Failed to update preferences:', error);
            throw error;
        }
    }

    // Merge updates into preferences object
    mergeUpdates(target, updates) {
        Object.keys(updates).forEach(key => {
            if (typeof updates[key] === 'object' && updates[key] !== null && !Array.isArray(updates[key])) {
                if (!target[key]) target[key] = {};
                this.mergeUpdates(target[key], updates[key]);
            } else {
                target[key] = updates[key];
            }
        });
    }

    // Debounced update for frequent changes
    updatePreferencesDebounced(updates, delay = 500) {
        const key = JSON.stringify(Object.keys(updates).sort());

        if (this.debounceTimers.has(key)) {
            clearTimeout(this.debounceTimers.get(key));
        }

        const timer = setTimeout(() => {
            this.updatePreferences(updates);
            this.debounceTimers.delete(key);
        }, delay);

        this.debounceTimers.set(key, timer);
    }

    // Set window preferences
    async setWindowPreferences(prefs) {
        return this.updatePreferences({ window: prefs });
    }

    // Set theme preferences
    async setThemePreferences(prefs) {
        return this.updatePreferences({ theme: prefs });
    }

    // Set dashboard preferences
    async setDashboardPreferences(prefs) {
        return this.updatePreferences({ dashboard: prefs });
    }

    // Set chart preferences
    async setChartPreferences(prefs) {
        return this.updatePreferences({ charts: prefs });
    }

    // Set UI preferences
    async setUIPreferences(prefs) {
        return this.updatePreferences({ ui: prefs });
    }

    // Observe preference changes
    observe(path, callback) {
        if (!this.observers.has(path)) {
            this.observers.set(path, new Set());
        }
        this.observers.get(path).add(callback);

        // Return unsubscribe function
        return () => {
            const callbacks = this.observers.get(path);
            if (callbacks) {
                callbacks.delete(callback);
                if (callbacks.size === 0) {
                    this.observers.delete(path);
                }
            }
        };
    }

    // Notify observers of changes
    notifyObservers(updates) {
        this.observers.forEach((callbacks, path) => {
            if (this.hasPathUpdate(updates, path)) {
                callbacks.forEach(callback => {
                    try {
                        callback(this.getValueByPath(this.preferences, path));
                    } catch (error) {
                        console.error('Error in preference observer:', error);
                    }
                });
            }
        });
    }

    // Check if updates contain changes for a specific path
    hasPathUpdate(updates, path) {
        const parts = path.split('.');
        let current = updates;

        for (const part of parts) {
            if (current && typeof current === 'object' && part in current) {
                current = current[part];
            } else {
                return false;
            }
        }

        return true;
    }

    // Get value by path
    getValueByPath(obj, path) {
        const parts = path.split('.');
        let current = obj;

        for (const part of parts) {
            if (current && typeof current === 'object' && part in current) {
                current = current[part];
            } else {
                return undefined;
            }
        }

        return current;
    }

    // Setup event listeners for automatic preference updates
    setupEventListeners() {
        // Window resize
        let resizeTimeout;
        window.addEventListener('resize', () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                this.updatePreferencesDebounced({
                    window: {
                        width: window.innerWidth,
                        height: window.innerHeight
                    }
                });
            }, 250);
        });

        // Theme changes
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        mediaQuery.addEventListener('change', (e) => {
            if (this.preferences.theme?.mode === 'auto') {
                this.applyThemePreferences();
            }
        });

        // Sidebar resize observer
        this.setupSidebarObserver();
    }

    setupSidebarObserver() {
        const sidebar = document.querySelector('.sidebar');
        if (!sidebar) return;

        const resizeObserver = new ResizeObserver(entries => {
            for (const entry of entries) {
                const width = entry.contentRect.width;
                this.updatePreferencesDebounced({
                    ui: { sidebarWidth: width }
                }, 1000);
            }
        });

        resizeObserver.observe(sidebar);
    }

    // Get current preferences
    getPreferences() {
        return { ...this.preferences };
    }

    // Reset preferences to defaults
    async resetPreferences() {
        try {
            await window.go.main.App.ResetPreferences();
            this.preferences = await window.go.main.App.GetPreferences();
            this.applyPreferences();
        } catch (error) {
            console.error('Failed to reset preferences:', error);
            throw error;
        }
    }
}

// Global preferences instance
window.preferences = new PreferencesHandler();

// Export for modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PreferencesHandler;
}
```

### Integration with App
Update `app.go` to include preferences methods:
```go
// Add preferences manager to App struct
type App struct {
    // ... existing fields
    preferencesManager *preferences.PreferencesManager
}

// GetPreferences returns current preferences
func (a *App) GetPreferences(ctx context.Context) (*preferences.Preferences, error) {
    return a.preferencesManager.GetPreferences(), nil
}

// UpdatePreferences updates preferences
func (a *App) UpdatePreferences(ctx context.Context, updates map[string]interface{}) error {
    return a.preferencesManager.UpdatePreferences(updates)
}

// ResetPreferences resets preferences to defaults
func (a *App) ResetPreferences(ctx context.Context) error {
    return a.preferencesManager.Reset()
}

// SetWindowPreferences updates window preferences
func (a *App) SetWindowPreferences(ctx context.Context, prefs *preferences.WindowPrefs) error {
    return a.preferencesManager.SetWindowPreferences(prefs)
}

// Initialize preferences manager in startup
func (a *App) startup(ctx context.Context) {
    // ... existing startup code

    var err error
    a.preferencesManager, err = preferences.NewPreferencesManager(a.dataDir)
    if err != nil {
        a.logger.Error("Failed to initialize preferences manager", "error", err)
        return
    }
}
```

## Verification Steps
1. Start the application and verify preferences are loaded correctly
2. Change window size/position and verify they are restored on restart
3. Modify theme settings and verify they persist across sessions
4. Test dashboard layout changes are remembered
5. Verify chart preferences are applied correctly
6. Test preference migration with simulated version updates
7. Verify corrupted preferences file is handled gracefully
8. Test preference reset functionality
9. Verify auto-save works correctly
10. Test preference observation system works

## Dependencies
- T003: Application Structure (for basic app framework)
- T004: Configuration Management (for configuration infrastructure)
- T030: Responsive Dashboard (for dashboard layout preferences)

## Notes
- Preferences are automatically saved periodically and on application shutdown
- The system supports migration for version updates
- Window state preferences may require platform-specific handling
- The preference observation system allows components to react to changes
- Debouncing prevents excessive saves during frequent updates
- The system is designed to be extensible for future preference additions