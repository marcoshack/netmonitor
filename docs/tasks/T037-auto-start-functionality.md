# T037: Auto-Start Functionality

## Overview
Implement auto-start functionality that allows NetMonitor to automatically launch when the system starts, with cross-platform support and user configuration options.

## Context
NetMonitor is designed to be a continuous monitoring tool that should run in the background. Users need the option to have the application start automatically with their system to ensure uninterrupted monitoring.

## Task Description
Create cross-platform auto-start functionality with proper system integration, user configuration options, and fallback mechanisms for different operating systems.

## Acceptance Criteria
- [ ] Cross-platform auto-start support (Windows, macOS, Linux)
- [ ] User configuration to enable/disable auto-start
- [ ] Proper system integration (registry, launch agents, desktop files)
- [ ] Auto-start status detection and reporting
- [ ] Silent start option (start minimized to tray)
- [ ] Error handling for permission issues
- [ ] Uninstall cleanup of auto-start entries
- [ ] Command-line options for controlling auto-start behavior

## Auto-Start Manager Implementation
```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
)

type AutoStartManager struct {
    appName      string
    executablePath string
    startMinimized bool
}

type AutoStartConfig struct {
    Enabled        bool `json:"enabled"`
    StartMinimized bool `json:"startMinimized"`
    DelaySeconds   int  `json:"delaySeconds"`
}

func NewAutoStartManager(appName string) *AutoStartManager {
    executable, _ := os.Executable()
    return &AutoStartManager{
        appName:        appName,
        executablePath: executable,
        startMinimized: true,
    }
}

func (asm *AutoStartManager) Enable(config AutoStartConfig) error {
    switch runtime.GOOS {
    case "windows":
        return asm.enableWindows(config)
    case "darwin":
        return asm.enableMacOS(config)
    case "linux":
        return asm.enableLinux(config)
    default:
        return fmt.Errorf("auto-start not supported on %s", runtime.GOOS)
    }
}

func (asm *AutoStartManager) Disable() error {
    switch runtime.GOOS {
    case "windows":
        return asm.disableWindows()
    case "darwin":
        return asm.disableMacOS()
    case "linux":
        return asm.disableLinux()
    default:
        return fmt.Errorf("auto-start not supported on %s", runtime.GOOS)
    }
}

func (asm *AutoStartManager) IsEnabled() (bool, error) {
    switch runtime.GOOS {
    case "windows":
        return asm.isEnabledWindows()
    case "darwin":
        return asm.isEnabledMacOS()
    case "linux":
        return asm.isEnabledLinux()
    default:
        return false, fmt.Errorf("auto-start not supported on %s", runtime.GOOS)
    }
}
```

## Windows Implementation
```go
//go:build windows

package main

import (
    "fmt"
    "strings"
    "syscall"
    "unsafe"

    "golang.org/x/sys/windows/registry"
)

const (
    registryKey = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
)

func (asm *AutoStartManager) enableWindows(config AutoStartConfig) error {
    key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
    if err != nil {
        return fmt.Errorf("failed to open registry key: %v", err)
    }
    defer key.Close()

    // Build command line arguments
    args := []string{asm.executablePath}
    if config.StartMinimized {
        args = append(args, "--start-minimized")
    }
    if config.DelaySeconds > 0 {
        args = append(args, fmt.Sprintf("--start-delay=%d", config.DelaySeconds))
    }

    command := strings.Join(args, " ")

    err = key.SetStringValue(asm.appName, command)
    if err != nil {
        return fmt.Errorf("failed to set registry value: %v", err)
    }

    return nil
}

func (asm *AutoStartManager) disableWindows() error {
    key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
    if err != nil {
        return fmt.Errorf("failed to open registry key: %v", err)
    }
    defer key.Close()

    err = key.DeleteValue(asm.appName)
    if err != nil && err != registry.ErrNotExist {
        return fmt.Errorf("failed to delete registry value: %v", err)
    }

    return nil
}

func (asm *AutoStartManager) isEnabledWindows() (bool, error) {
    key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.QUERY_VALUE)
    if err != nil {
        return false, nil // Key doesn't exist, auto-start is disabled
    }
    defer key.Close()

    _, _, err = key.GetStringValue(asm.appName)
    if err == registry.ErrNotExist {
        return false, nil
    }
    if err != nil {
        return false, fmt.Errorf("failed to query registry value: %v", err)
    }

    return true, nil
}

// Windows-specific helper to check if running as admin
func (asm *AutoStartManager) isRunningAsAdmin() bool {
    var sid *syscall.SID
    err := syscall.AllocateAndInitializeSid(
        &syscall.SECURITY_NT_AUTHORITY,
        2,
        syscall.SECURITY_BUILTIN_DOMAIN_RID,
        syscall.DOMAIN_ALIAS_RID_ADMINS,
        0, 0, 0, 0, 0, 0,
        &sid)
    if err != nil {
        return false
    }
    defer syscall.FreeSid(sid)

    token := syscall.Token(0)
    member, err := token.IsMember(sid)
    return err == nil && member
}
```

## macOS Implementation
```go
//go:build darwin

package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

func (asm *AutoStartManager) enableMacOS(config AutoStartConfig) error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("failed to get home directory: %v", err)
    }

    launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
    if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
        return fmt.Errorf("failed to create LaunchAgents directory: %v", err)
    }

    plistPath := filepath.Join(launchAgentsDir, fmt.Sprintf("com.netmonitor.%s.plist", asm.appName))

    // Build program arguments
    args := []string{asm.executablePath}
    if config.StartMinimized {
        args = append(args, "--start-minimized")
    }

    plistContent := asm.generateLaunchAgentPlist(args, config.DelaySeconds)

    err = ioutil.WriteFile(plistPath, []byte(plistContent), 0644)
    if err != nil {
        return fmt.Errorf("failed to write plist file: %v", err)
    }

    return nil
}

func (asm *AutoStartManager) disableMacOS() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("failed to get home directory: %v", err)
    }

    plistPath := filepath.Join(homeDir, "Library", "LaunchAgents",
                              fmt.Sprintf("com.netmonitor.%s.plist", asm.appName))

    err = os.Remove(plistPath)
    if err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to remove plist file: %v", err)
    }

    return nil
}

func (asm *AutoStartManager) isEnabledMacOS() (bool, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return false, fmt.Errorf("failed to get home directory: %v", err)
    }

    plistPath := filepath.Join(homeDir, "Library", "LaunchAgents",
                              fmt.Sprintf("com.netmonitor.%s.plist", asm.appName))

    _, err = os.Stat(plistPath)
    if os.IsNotExist(err) {
        return false, nil
    }
    if err != nil {
        return false, fmt.Errorf("failed to check plist file: %v", err)
    }

    return true, nil
}

func (asm *AutoStartManager) generateLaunchAgentPlist(args []string, delaySeconds int) string {
    argsXML := ""
    for _, arg := range args {
        argsXML += fmt.Sprintf("        <string>%s</string>\n", arg)
    }

    template := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.netmonitor.%s</string>
    <key>ProgramArguments</key>
    <array>
%s    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StartInterval</key>
    <integer>%d</integer>
    <key>StandardErrorPath</key>
    <string>/tmp/netmonitor.err</string>
    <key>StandardOutPath</key>
    <string>/tmp/netmonitor.out</string>
</dict>
</plist>`

    startInterval := 86400 // Run once per day by default
    if delaySeconds > 0 {
        startInterval = delaySeconds
    }

    return fmt.Sprintf(template, asm.appName, argsXML, startInterval)
}
```

## Linux Implementation
```go
//go:build linux

package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

func (asm *AutoStartManager) enableLinux(config AutoStartConfig) error {
    autostartDir, err := asm.getAutostartDir()
    if err != nil {
        return err
    }

    if err := os.MkdirAll(autostartDir, 0755); err != nil {
        return fmt.Errorf("failed to create autostart directory: %v", err)
    }

    desktopFile := filepath.Join(autostartDir, fmt.Sprintf("%s.desktop", asm.appName))

    // Build exec command
    exec := asm.executablePath
    if config.StartMinimized {
        exec += " --start-minimized"
    }
    if config.DelaySeconds > 0 {
        exec = fmt.Sprintf("sh -c 'sleep %d && %s'", config.DelaySeconds, exec)
    }

    desktopContent := asm.generateDesktopFile(exec)

    err = ioutil.WriteFile(desktopFile, []byte(desktopContent), 0644)
    if err != nil {
        return fmt.Errorf("failed to write desktop file: %v", err)
    }

    return nil
}

func (asm *AutoStartManager) disableLinux() error {
    autostartDir, err := asm.getAutostartDir()
    if err != nil {
        return err
    }

    desktopFile := filepath.Join(autostartDir, fmt.Sprintf("%s.desktop", asm.appName))

    err = os.Remove(desktopFile)
    if err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to remove desktop file: %v", err)
    }

    return nil
}

func (asm *AutoStartManager) isEnabledLinux() (bool, error) {
    autostartDir, err := asm.getAutostartDir()
    if err != nil {
        return false, err
    }

    desktopFile := filepath.Join(autostartDir, fmt.Sprintf("%s.desktop", asm.appName))

    _, err = os.Stat(desktopFile)
    if os.IsNotExist(err) {
        return false, nil
    }
    if err != nil {
        return false, fmt.Errorf("failed to check desktop file: %v", err)
    }

    return true, nil
}

func (asm *AutoStartManager) getAutostartDir() (string, error) {
    // Check XDG_CONFIG_HOME first
    if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
        return filepath.Join(xdgConfig, "autostart"), nil
    }

    // Fall back to ~/.config/autostart
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("failed to get home directory: %v", err)
    }

    return filepath.Join(homeDir, ".config", "autostart"), nil
}

func (asm *AutoStartManager) generateDesktopFile(exec string) string {
    template := `[Desktop Entry]
Name=%s
Comment=Network monitoring and performance analysis
Exec=%s
Icon=%s
Terminal=false
Type=Application
Categories=Network;System;
StartupNotify=false
X-GNOME-Autostart-enabled=true
Hidden=false`

    iconPath := filepath.Join(filepath.Dir(asm.executablePath), "netmonitor.png")
    return fmt.Sprintf(template, asm.appName, exec, iconPath)
}
```

## Application Integration
```go
// App methods for auto-start management
func (a *App) EnableAutoStart(config AutoStartConfig) error {
    autoStart := NewAutoStartManager("NetMonitor")
    err := autoStart.Enable(config)
    if err != nil {
        return fmt.Errorf("failed to enable auto-start: %v", err)
    }

    // Update configuration
    a.config.Settings.AutoStart = config
    return a.SaveConfiguration()
}

func (a *App) DisableAutoStart() error {
    autoStart := NewAutoStartManager("NetMonitor")
    err := autoStart.Disable()
    if err != nil {
        return fmt.Errorf("failed to disable auto-start: %v", err)
    }

    // Update configuration
    a.config.Settings.AutoStart.Enabled = false
    return a.SaveConfiguration()
}

func (a *App) GetAutoStartStatus() (*AutoStartStatus, error) {
    autoStart := NewAutoStartManager("NetMonitor")
    enabled, err := autoStart.IsEnabled()
    if err != nil {
        return nil, fmt.Errorf("failed to check auto-start status: %v", err)
    }

    return &AutoStartStatus{
        Enabled:        enabled,
        Supported:      true,
        CurrentConfig:  a.config.Settings.AutoStart,
        RequiresElevation: asm.requiresElevation(),
    }, nil
}

type AutoStartStatus struct {
    Enabled           bool            `json:"enabled"`
    Supported         bool            `json:"supported"`
    CurrentConfig     AutoStartConfig `json:"currentConfig"`
    RequiresElevation bool            `json:"requiresElevation"`
}

func (asm *AutoStartManager) requiresElevation() bool {
    switch runtime.GOOS {
    case "windows":
        // Check if we need admin rights for system-wide installation
        return false // HKEY_CURRENT_USER doesn't require elevation
    case "darwin":
        return false // User LaunchAgents don't require elevation
    case "linux":
        return false // User autostart doesn't require elevation
    default:
        return false
    }
}
```

## Command Line Arguments
```go
// Add to main.go for handling auto-start arguments
func init() {
    flag.BoolVar(&startMinimized, "start-minimized", false, "Start application minimized to system tray")
    flag.IntVar(&startDelay, "start-delay", 0, "Delay in seconds before starting")
    flag.BoolVar(&enableAutoStart, "enable-autostart", false, "Enable auto-start (requires configuration)")
    flag.BoolVar(&disableAutoStart, "disable-autostart", false, "Disable auto-start")
}

var (
    startMinimized   bool
    startDelay       int
    enableAutoStart  bool
    disableAutoStart bool
)

func main() {
    flag.Parse()

    // Handle auto-start management from command line
    if enableAutoStart {
        autoStart := NewAutoStartManager("NetMonitor")
        config := AutoStartConfig{
            Enabled:        true,
            StartMinimized: true,
            DelaySeconds:   0,
        }
        if err := autoStart.Enable(config); err != nil {
            log.Fatalf("Failed to enable auto-start: %v", err)
        }
        fmt.Println("Auto-start enabled successfully")
        return
    }

    if disableAutoStart {
        autoStart := NewAutoStartManager("NetMonitor")
        if err := autoStart.Disable(); err != nil {
            log.Fatalf("Failed to disable auto-start: %v", err)
        }
        fmt.Println("Auto-start disabled successfully")
        return
    }

    // Handle start delay
    if startDelay > 0 {
        log.Printf("Waiting %d seconds before starting...", startDelay)
        time.Sleep(time.Duration(startDelay) * time.Second)
    }

    // Normal application startup
    app := NewApp()

    // Configure startup behavior
    options := &wails.Options{
        Title:         "NetMonitor",
        Width:         1024,
        Height:        768,
        WindowStartState: options.Normal,
    }

    if startMinimized {
        options.WindowStartState = options.Minimised
    }

    // ... rest of Wails app initialization
}
```

## Frontend Settings Integration
```javascript
// Settings component for auto-start configuration
class AutoStartSettings {
    constructor(container) {
        this.container = container;
        this.init();
    }

    async init() {
        await this.loadCurrentStatus();
        this.renderSettings();
        this.bindEvents();
    }

    async loadCurrentStatus() {
        try {
            this.status = await window.go.App.GetAutoStartStatus();
        } catch (error) {
            console.error('Failed to load auto-start status:', error);
            this.status = { enabled: false, supported: false };
        }
    }

    renderSettings() {
        this.container.innerHTML = `
            <div class="auto-start-settings">
                <h3>Startup Options</h3>

                <div class="setting-item">
                    <label class="setting-label">
                        <input type="checkbox" id="auto-start-enabled"
                               ${this.status.enabled ? 'checked' : ''}
                               ${!this.status.supported ? 'disabled' : ''}>
                        Start NetMonitor when system starts
                    </label>
                    ${!this.status.supported ? '<span class="setting-note">Not supported on this platform</span>' : ''}
                </div>

                <div class="setting-item ${!this.status.enabled ? 'disabled' : ''}">
                    <label class="setting-label">
                        <input type="checkbox" id="start-minimized"
                               ${this.status.currentConfig?.startMinimized ? 'checked' : ''}>
                        Start minimized to system tray
                    </label>
                </div>

                <div class="setting-item ${!this.status.enabled ? 'disabled' : ''}">
                    <label class="setting-label">
                        Startup delay:
                        <input type="number" id="start-delay" min="0" max="300"
                               value="${this.status.currentConfig?.delaySeconds || 0}">
                        seconds
                    </label>
                </div>

                <button id="apply-auto-start" class="btn btn-primary">Apply Settings</button>
            </div>
        `;
    }

    bindEvents() {
        const enabledCheckbox = this.container.querySelector('#auto-start-enabled');
        const applyButton = this.container.querySelector('#apply-auto-start');

        enabledCheckbox.addEventListener('change', (e) => {
            const dependentItems = this.container.querySelectorAll('.setting-item:not(:first-child)');
            dependentItems.forEach(item => {
                item.classList.toggle('disabled', !e.target.checked);
            });
        });

        applyButton.addEventListener('click', () => this.applySettings());
    }

    async applySettings() {
        const enabled = this.container.querySelector('#auto-start-enabled').checked;
        const startMinimized = this.container.querySelector('#start-minimized').checked;
        const delaySeconds = parseInt(this.container.querySelector('#start-delay').value) || 0;

        try {
            if (enabled) {
                await window.go.App.EnableAutoStart({
                    enabled: true,
                    startMinimized,
                    delaySeconds
                });
            } else {
                await window.go.App.DisableAutoStart();
            }

            this.showSuccess('Auto-start settings updated successfully');
        } catch (error) {
            this.showError(`Failed to update auto-start settings: ${error.message}`);
        }
    }
}
```

## Verification Steps
1. Test auto-start enablement - should create appropriate system entries
2. Verify cross-platform functionality - should work on Windows, macOS, and Linux
3. Test auto-start disabling - should remove system entries cleanly
4. Verify command-line options - should handle start minimized and delay
5. Test permission handling - should work without requiring elevation
6. Verify uninstall cleanup - should remove auto-start entries when uninstalling
7. Test configuration persistence - should remember auto-start settings
8. Verify system integration - should start properly with system boot

## Dependencies
- T036: System Tray Integration
- T003: Configuration System
- Operating system APIs for auto-start management

## Notes
- Implement proper error handling for insufficient permissions
- Consider implementing elevated privilege requests when necessary
- Test thoroughly on different operating system versions
- Handle edge cases like missing directories or corrupted entries
- Plan for future installer integration
- Consider implementing verification that auto-start is working correctly