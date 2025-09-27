# T036: System Tray Integration

## Overview
Implement system tray integration for NetMonitor, allowing the application to run in the background with system tray icon, context menu, and status indicators.

## Context
NetMonitor should operate as a background monitoring application with minimal desktop footprint. The system tray provides quick access to monitoring status, controls, and the main dashboard without cluttering the taskbar.

## Task Description
Create comprehensive system tray integration with status-aware icons, context menus, notifications, and seamless window management for cross-platform operation.

## Acceptance Criteria
- [ ] System tray icon with status indicators
- [ ] Context menu with essential actions
- [ ] Click handling for show/hide main window
- [ ] Status-aware icon changes (healthy, warning, critical)
- [ ] Cross-platform compatibility (Windows, macOS, Linux)
- [ ] Graceful degradation when system tray not available
- [ ] Balloon/toast notifications from tray
- [ ] Tray icon tooltip with current status
- [ ] Application lifecycle management through tray

## System Tray Implementation (Go)
```go
package main

import (
    "context"
    "embed"
    "fmt"
    "log"

    "github.com/getlantern/systray"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

//go:embed icons/tray-healthy.ico
var iconHealthy []byte

//go:embed icons/tray-warning.ico
var iconWarning []byte

//go:embed icons/tray-critical.ico
var iconCritical []byte

type SystemTray struct {
    app           *App
    isVisible     bool
    currentStatus string
    menuItems     map[string]*systray.MenuItem
}

func NewSystemTray(app *App) *SystemTray {
    return &SystemTray{
        app:       app,
        isVisible: true,
        menuItems: make(map[string]*systray.MenuItem),
    }
}

func (st *SystemTray) Initialize() error {
    go func() {
        systray.Run(st.onReady, st.onExit)
    }()
    return nil
}

func (st *SystemTray) onReady() {
    // Set initial icon and tooltip
    st.setTrayIcon("healthy")
    systray.SetTooltip("NetMonitor - Network monitoring is running")

    // Create menu items
    st.createMenuItems()

    // Setup menu item handlers
    st.setupMenuHandlers()

    // Start status monitoring
    go st.statusMonitorLoop()
}

func (st *SystemTray) createMenuItems() {
    // Show/Hide window
    st.menuItems["toggle"] = systray.AddMenuItem("Show NetMonitor", "Show or hide the main window")

    systray.AddSeparator()

    // Status section
    st.menuItems["status"] = systray.AddMenuItem("Status: Healthy", "Current monitoring status")
    st.menuItems["status"].Disable()

    st.menuItems["lastTest"] = systray.AddMenuItem("Last test: 2 minutes ago", "Time since last test")
    st.menuItems["lastTest"].Disable()

    systray.AddSeparator()

    // Quick actions
    st.menuItems["runTest"] = systray.AddMenuItem("Run Manual Test", "Execute tests on all endpoints")
    st.menuItems["pauseMonitoring"] = systray.AddMenuItem("Pause Monitoring", "Temporarily stop monitoring")

    systray.AddSeparator()

    // Regions submenu
    st.menuItems["regions"] = systray.AddMenuItem("Regions", "View regional status")
    regionsMenu := st.menuItems["regions"].AddSubMenu()

    st.menuItems["regionNA"] = regionsMenu.AddSubMenuItem("NA-East: Healthy", "North America East region")
    st.menuItems["regionEU"] = regionsMenu.AddSubMenuItem("EU-West: Warning", "Europe West region")
    st.menuItems["regionAP"] = regionsMenu.AddSubMenuItem("Asia-Pacific: Healthy", "Asia Pacific region")

    systray.AddSeparator()

    // Settings and controls
    st.menuItems["settings"] = systray.AddMenuItem("Settings", "Open settings window")
    st.menuItems["about"] = systray.AddMenuItem("About", "About NetMonitor")

    systray.AddSeparator()

    // Exit
    st.menuItems["quit"] = systray.AddMenuItem("Quit NetMonitor", "Exit the application")
}

func (st *SystemTray) setupMenuHandlers() {
    // Toggle main window
    go func() {
        for {
            select {
            case <-st.menuItems["toggle"].ClickedCh:
                st.toggleMainWindow()
            }
        }
    }()

    // Run manual test
    go func() {
        for {
            select {
            case <-st.menuItems["runTest"].ClickedCh:
                st.runManualTest()
            }
        }
    }()

    // Pause/Resume monitoring
    go func() {
        for {
            select {
            case <-st.menuItems["pauseMonitoring"].ClickedCh:
                st.toggleMonitoring()
            }
        }
    }()

    // Settings
    go func() {
        for {
            select {
            case <-st.menuItems["settings"].ClickedCh:
                st.openSettings()
            }
        }
    }()

    // About
    go func() {
        for {
            select {
            case <-st.menuItems["about"].ClickedCh:
                st.showAbout()
            }
        }
    }()

    // Quit application
    go func() {
        for {
            select {
            case <-st.menuItems["quit"].ClickedCh:
                st.quitApplication()
            }
        }
    }()
}

func (st *SystemTray) statusMonitorLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            st.updateTrayStatus()
        }
    }
}

func (st *SystemTray) updateTrayStatus() {
    status, err := st.app.GetMonitoringStatus()
    if err != nil {
        log.Printf("Failed to get monitoring status: %v", err)
        return
    }

    // Update tray icon based on overall health
    overallStatus := st.calculateOverallStatus(status)
    if overallStatus != st.currentStatus {
        st.setTrayIcon(overallStatus)
        st.currentStatus = overallStatus
    }

    // Update menu items
    st.updateMenuItems(status)

    // Update tooltip
    tooltip := fmt.Sprintf("NetMonitor - %s\nEndpoints: %d healthy, %d warning, %d down",
        st.formatStatus(overallStatus),
        status.HealthyCount,
        status.WarningCount,
        status.DownCount)
    systray.SetTooltip(tooltip)
}

func (st *SystemTray) setTrayIcon(status string) {
    var iconData []byte

    switch status {
    case "healthy":
        iconData = iconHealthy
    case "warning":
        iconData = iconWarning
    case "critical":
        iconData = iconCritical
    default:
        iconData = iconHealthy
    }

    systray.SetIcon(iconData)
}

func (st *SystemTray) updateMenuItems(status *MonitoringStatus) {
    // Update status text
    statusText := fmt.Sprintf("Status: %s", st.formatStatus(st.currentStatus))
    st.menuItems["status"].SetTitle(statusText)

    // Update last test time
    lastTestText := fmt.Sprintf("Last test: %s", st.formatRelativeTime(status.LastTestTime))
    st.menuItems["lastTest"].SetTitle(lastTestText)

    // Update pause/resume button
    if status.Running {
        st.menuItems["pauseMonitoring"].SetTitle("Pause Monitoring")
    } else {
        st.menuItems["pauseMonitoring"].SetTitle("Resume Monitoring")
    }

    // Update regional status
    for regionName, regionStatus := range status.Regions {
        if menuItem, exists := st.menuItems[fmt.Sprintf("region%s", regionName)]; exists {
            title := fmt.Sprintf("%s: %s", regionName, st.formatStatus(regionStatus.Health))
            menuItem.SetTitle(title)
        }
    }
}

func (st *SystemTray) toggleMainWindow() {
    if st.isVisible {
        st.app.HideWindow()
        st.menuItems["toggle"].SetTitle("Show NetMonitor")
        st.isVisible = false
    } else {
        st.app.ShowWindow()
        st.menuItems["toggle"].SetTitle("Hide NetMonitor")
        st.isVisible = true
    }
}

func (st *SystemTray) runManualTest() {
    // Disable menu item during test
    st.menuItems["runTest"].Disable()
    st.menuItems["runTest"].SetTitle("Running tests...")

    go func() {
        defer func() {
            st.menuItems["runTest"].Enable()
            st.menuItems["runTest"].SetTitle("Run Manual Test")
        }()

        results, err := st.app.RunAllTests()
        if err != nil {
            st.showNotification("Test Failed", fmt.Sprintf("Manual test failed: %v", err))
            return
        }

        // Show notification with results
        successCount := 0
        for _, result := range results {
            if result.Status == "success" {
                successCount++
            }
        }

        st.showNotification("Test Completed",
            fmt.Sprintf("Manual test completed: %d/%d endpoints successful",
                successCount, len(results)))
    }()
}

func (st *SystemTray) toggleMonitoring() {
    status, err := st.app.GetMonitoringStatus()
    if err != nil {
        st.showNotification("Error", "Failed to get monitoring status")
        return
    }

    if status.Running {
        err = st.app.StopMonitoring()
        if err != nil {
            st.showNotification("Error", "Failed to stop monitoring")
        } else {
            st.showNotification("Monitoring Paused", "Network monitoring has been paused")
        }
    } else {
        err = st.app.StartMonitoring()
        if err != nil {
            st.showNotification("Error", "Failed to start monitoring")
        } else {
            st.showNotification("Monitoring Resumed", "Network monitoring has been resumed")
        }
    }
}

func (st *SystemTray) openSettings() {
    st.app.ShowWindow()
    st.app.NavigateToSettings()
}

func (st *SystemTray) showAbout() {
    st.showNotification("NetMonitor",
        fmt.Sprintf("NetMonitor v%s\nNetwork monitoring and performance analysis tool", st.app.GetVersion()))
}

func (st *SystemTray) quitApplication() {
    st.app.Quit()
}

func (st *SystemTray) showNotification(title, message string) {
    // Platform-specific notification implementation
    systray.ShowNotification(title, message)
}

func (st *SystemTray) calculateOverallStatus(status *MonitoringStatus) string {
    if status.DownCount > 0 {
        return "critical"
    }
    if status.WarningCount > 0 {
        return "warning"
    }
    return "healthy"
}

func (st *SystemTray) formatStatus(status string) string {
    switch status {
    case "healthy":
        return "Healthy"
    case "warning":
        return "Warning"
    case "critical":
        return "Critical"
    default:
        return "Unknown"
    }
}

func (st *SystemTray) formatRelativeTime(t time.Time) string {
    duration := time.Since(t)
    if duration < time.Minute {
        return "just now"
    } else if duration < time.Hour {
        return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
    } else {
        return fmt.Sprintf("%d hours ago", int(duration.Hours()))
    }
}

func (st *SystemTray) onExit() {
    // Cleanup when systray exits
    log.Println("System tray exiting")
}
```

## Wails Integration
```go
// App struct modification to support system tray
type App struct {
    ctx      context.Context
    tray     *SystemTray
    window   *wails.Window
    config   *Config
    monitor  *Monitor
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx

    // Initialize system tray
    a.tray = NewSystemTray(a)
    if err := a.tray.Initialize(); err != nil {
        log.Printf("Failed to initialize system tray: %v", err)
    }
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
    // Check if we should minimize to tray instead of closing
    if a.config.Settings.MinimizeToTray {
        a.HideWindow()
        return true // Prevent actual close
    }
    return false // Allow close
}

func (a *App) HideWindow() {
    runtime.Hide(a.ctx)
    a.tray.isVisible = false
    a.tray.menuItems["toggle"].SetTitle("Show NetMonitor")
}

func (a *App) ShowWindow() {
    runtime.Show(a.ctx)
    runtime.WindowSetAlwaysOnTop(a.ctx, false)
    a.tray.isVisible = true
    a.tray.menuItems["toggle"].SetTitle("Hide NetMonitor")
}

func (a *App) NavigateToSettings() {
    // Emit event to frontend to navigate to settings
    runtime.EventsEmit(a.ctx, "navigate", "settings")
}

func (a *App) StartMonitoring() error {
    return a.monitor.Start()
}

func (a *App) StopMonitoring() error {
    return a.monitor.Stop()
}

func (a *App) GetVersion() string {
    return "1.0.0" // Or get from build info
}

func (a *App) Quit() {
    runtime.Quit(a.ctx)
}
```

## Cross-Platform Considerations
```go
// Platform-specific tray icon handling
//go:build windows
package main

import "github.com/getlantern/systray"

func (st *SystemTray) showNotification(title, message string) {
    // Windows toast notification
    systray.ShowNotification(title, message)
}

//go:build darwin
package main

import "github.com/getlantern/systray"

func (st *SystemTray) showNotification(title, message string) {
    // macOS notification center
    systray.ShowNotification(title, message)
}

//go:build linux
package main

import "github.com/getlantern/systray"

func (st *SystemTray) showNotification(title, message string) {
    // Linux libnotify
    systray.ShowNotification(title, message)
}
```

## Frontend Integration
```javascript
// Handle navigation events from system tray
window.addEventListener('wails:runtime', () => {
    window.runtime.EventsOn('navigate', (section) => {
        // Navigate to specific section
        switch (section) {
            case 'settings':
                showSettingsPanel();
                break;
            case 'overview':
                showOverviewPanel();
                break;
        }
    });
});

// Handle window visibility events
window.runtime.EventsOn('window-hidden', () => {
    // Pause expensive operations when hidden
    pauseRealTimeUpdates();
});

window.runtime.EventsOn('window-shown', () => {
    // Resume operations when shown
    resumeRealTimeUpdates();
    refreshAllData();
});
```

## Configuration Options
```go
type SystemTrayConfig struct {
    Enabled                bool   `json:"enabled"`
    ShowNotifications      bool   `json:"showNotifications"`
    MinimizeToTray        bool   `json:"minimizeToTray"`
    StartMinimized        bool   `json:"startMinimized"`
    NotificationThreshold string `json:"notificationThreshold"` // "warning", "critical"
    UpdateInterval        int    `json:"updateInterval"`        // seconds
}
```

## Verification Steps
1. Test system tray icon appearance - should show in system tray
2. Verify context menu functionality - should show all menu items and handle clicks
3. Test icon status changes - should change icon based on monitoring status
4. Verify window show/hide - should minimize to tray and restore from tray
5. Test notifications - should show balloon/toast notifications
6. Verify cross-platform behavior - should work on Windows, macOS, and Linux
7. Test application lifecycle - should handle quit properly from tray
8. Verify tooltip updates - should show current status information

## Dependencies
- T002: Basic Application Structure
- T015: Monitoring Status API
- T012: Manual Test Execution
- systray library for Go

## Notes
- Design icons at multiple resolutions for different screen densities
- Consider implementing tray icon animations for active monitoring
- Handle graceful degradation when system tray is not available
- Implement proper error handling for tray operations
- Consider adding quick shortcuts for common actions
- Plan for future customizable tray menu options