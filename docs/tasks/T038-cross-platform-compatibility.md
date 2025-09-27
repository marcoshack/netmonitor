# T038: Cross-Platform Compatibility

## Overview
Ensure NetMonitor works consistently across Windows, macOS, and Linux with platform-specific optimizations and feature adaptations.

## Context
NetMonitor must provide a consistent user experience across different operating systems while adapting to platform-specific conventions, file systems, and system integration requirements.

## Task Description
Implement comprehensive cross-platform compatibility with platform-specific adaptations, testing frameworks, and optimization for each supported operating system.

## Acceptance Criteria
- [ ] Consistent functionality across Windows, macOS, and Linux
- [ ] Platform-specific UI adaptations and conventions
- [ ] File system compatibility and path handling
- [ ] Network stack optimizations for each platform
- [ ] Build and packaging for all platforms
- [ ] Platform-specific testing and validation
- [ ] Documentation for platform-specific features
- [ ] Graceful degradation for platform-specific limitations

## Platform Abstraction Layer
```go
package platform

import (
    "runtime"
    "path/filepath"
    "os"
)

type Platform interface {
    GetConfigDir() (string, error)
    GetDataDir() (string, error)
    GetLogDir() (string, error)
    GetTempDir() (string, error)
    OpenFileManager(path string) error
    OpenURL(url string) error
    GetSystemInfo() (*SystemInfo, error)
    SupportsSystemTray() bool
    SupportsNotifications() bool
    GetNetworkInterfaces() ([]NetworkInterface, error)
}

type SystemInfo struct {
    OS           string `json:"os"`
    Version      string `json:"version"`
    Architecture string `json:"architecture"`
    Hostname     string `json:"hostname"`
    Username     string `json:"username"`
    HomeDir      string `json:"homeDir"`
    TempDir      string `json:"tempDir"`
}

type NetworkInterface struct {
    Name         string   `json:"name"`
    DisplayName  string   `json:"displayName"`
    Type         string   `json:"type"`
    Status       string   `json:"status"`
    IPAddresses  []string `json:"ipAddresses"`
    MACAddress   string   `json:"macAddress"`
    MTU          int      `json:"mtu"`
}

func NewPlatform() Platform {
    switch runtime.GOOS {
    case "windows":
        return &WindowsPlatform{}
    case "darwin":
        return &DarwinPlatform{}
    case "linux":
        return &LinuxPlatform{}
    default:
        return &GenericPlatform{}
    }
}

// Generic implementation as fallback
type GenericPlatform struct{}

func (p *GenericPlatform) GetConfigDir() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, ".config", "netmonitor"), nil
}

func (p *GenericPlatform) GetDataDir() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, ".local", "share", "netmonitor"), nil
}

func (p *GenericPlatform) GetLogDir() (string, error) {
    dataDir, err := p.GetDataDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dataDir, "logs"), nil
}

func (p *GenericPlatform) GetTempDir() (string, error) {
    return filepath.Join(os.TempDir(), "netmonitor"), nil
}

func (p *GenericPlatform) SupportsSystemTray() bool {
    return false
}

func (p *GenericPlatform) SupportsNotifications() bool {
    return false
}
```

## Windows Platform Implementation
```go
//go:build windows

package platform

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "unsafe"

    "golang.org/x/sys/windows"
    "golang.org/x/sys/windows/registry"
)

type WindowsPlatform struct{}

func (p *WindowsPlatform) GetConfigDir() (string, error) {
    appData := os.Getenv("APPDATA")
    if appData == "" {
        return "", fmt.Errorf("APPDATA environment variable not set")
    }
    return filepath.Join(appData, "NetMonitor"), nil
}

func (p *WindowsPlatform) GetDataDir() (string, error) {
    localAppData := os.Getenv("LOCALAPPDATA")
    if localAppData == "" {
        return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
    }
    return filepath.Join(localAppData, "NetMonitor"), nil
}

func (p *WindowsPlatform) GetLogDir() (string, error) {
    dataDir, err := p.GetDataDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dataDir, "Logs"), nil
}

func (p *WindowsPlatform) GetTempDir() (string, error) {
    return filepath.Join(os.TempDir(), "NetMonitor"), nil
}

func (p *WindowsPlatform) OpenFileManager(path string) error {
    return exec.Command("explorer", path).Start()
}

func (p *WindowsPlatform) OpenURL(url string) error {
    return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func (p *WindowsPlatform) GetSystemInfo() (*SystemInfo, error) {
    hostname, _ := os.Hostname()
    username := os.Getenv("USERNAME")
    homeDir, _ := os.UserHomeDir()

    version, err := p.getWindowsVersion()
    if err != nil {
        version = "Unknown"
    }

    return &SystemInfo{
        OS:           "Windows",
        Version:      version,
        Architecture: runtime.GOARCH,
        Hostname:     hostname,
        Username:     username,
        HomeDir:      homeDir,
        TempDir:      os.TempDir(),
    }, nil
}

func (p *WindowsPlatform) getWindowsVersion() (string, error) {
    k, err := registry.OpenKey(registry.LOCAL_MACHINE,
        `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
    if err != nil {
        return "", err
    }
    defer k.Close()

    productName, _, err := k.GetStringValue("ProductName")
    if err != nil {
        return "", err
    }

    buildNumber, _, err := k.GetStringValue("CurrentBuild")
    if err != nil {
        return productName, nil
    }

    return fmt.Sprintf("%s (Build %s)", productName, buildNumber), nil
}

func (p *WindowsPlatform) SupportsSystemTray() bool {
    return true
}

func (p *WindowsPlatform) SupportsNotifications() bool {
    return true
}

func (p *WindowsPlatform) GetNetworkInterfaces() ([]NetworkInterface, error) {
    // Windows-specific network interface enumeration
    return p.getWindowsNetworkInterfaces()
}

func (p *WindowsPlatform) getWindowsNetworkInterfaces() ([]NetworkInterface, error) {
    // Use Windows APIs to get detailed network interface information
    // This would involve calling GetAdaptersAddresses and related APIs
    interfaces := []NetworkInterface{}

    // Implementation would use Windows IP Helper API
    // For brevity, showing structure only

    return interfaces, nil
}

// Windows-specific privilege checking
func (p *WindowsPlatform) IsRunningAsAdmin() bool {
    var sid *windows.SID
    err := windows.AllocateAndInitializeSid(
        &windows.SECURITY_NT_AUTHORITY,
        2,
        windows.SECURITY_BUILTIN_DOMAIN_RID,
        windows.DOMAIN_ALIAS_RID_ADMINS,
        0, 0, 0, 0, 0, 0,
        &sid)
    if err != nil {
        return false
    }
    defer windows.FreeSid(sid)

    token := windows.Token(0)
    member, err := token.IsMember(sid)
    return err == nil && member
}
```

## macOS Platform Implementation
```go
//go:build darwin

package platform

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "syscall"
)

type DarwinPlatform struct{}

func (p *DarwinPlatform) GetConfigDir() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, "Library", "Preferences", "NetMonitor"), nil
}

func (p *DarwinPlatform) GetDataDir() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, "Library", "Application Support", "NetMonitor"), nil
}

func (p *DarwinPlatform) GetLogDir() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, "Library", "Logs", "NetMonitor"), nil
}

func (p *DarwinPlatform) OpenFileManager(path string) error {
    return exec.Command("open", path).Start()
}

func (p *DarwinPlatform) OpenURL(url string) error {
    return exec.Command("open", url).Start()
}

func (p *DarwinPlatform) GetSystemInfo() (*SystemInfo, error) {
    hostname, _ := os.Hostname()
    username := os.Getenv("USER")
    homeDir, _ := os.UserHomeDir()

    version, err := p.getMacOSVersion()
    if err != nil {
        version = "Unknown"
    }

    return &SystemInfo{
        OS:           "macOS",
        Version:      version,
        Architecture: runtime.GOARCH,
        Hostname:     hostname,
        Username:     username,
        HomeDir:      homeDir,
        TempDir:      os.TempDir(),
    }, nil
}

func (p *DarwinPlatform) getMacOSVersion() (string, error) {
    out, err := exec.Command("sw_vers", "-productVersion").Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}

func (p *DarwinPlatform) SupportsSystemTray() bool {
    return true
}

func (p *DarwinPlatform) SupportsNotifications() bool {
    return true
}

func (p *DarwinPlatform) GetNetworkInterfaces() ([]NetworkInterface, error) {
    // macOS-specific network interface enumeration
    return p.getMacOSNetworkInterfaces()
}

func (p *DarwinPlatform) getMacOSNetworkInterfaces() ([]NetworkInterface, error) {
    // Use system commands or syscalls to get network interfaces
    interfaces := []NetworkInterface{}

    // Example using networksetup command
    out, err := exec.Command("networksetup", "-listallhardwareports").Output()
    if err != nil {
        return interfaces, err
    }

    // Parse output and populate interfaces
    // Implementation details would go here

    return interfaces, nil
}

// macOS-specific permission checking
func (p *DarwinPlatform) CheckNetworkPermissions() error {
    // Check if app has necessary network permissions
    // This might involve checking for specific entitlements
    return nil
}
```

## Linux Platform Implementation
```go
//go:build linux

package platform

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

type LinuxPlatform struct{}

func (p *LinuxPlatform) GetConfigDir() (string, error) {
    if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
        return filepath.Join(xdgConfig, "netmonitor"), nil
    }

    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, ".config", "netmonitor"), nil
}

func (p *LinuxPlatform) GetDataDir() (string, error) {
    if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
        return filepath.Join(xdgData, "netmonitor"), nil
    }

    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, ".local", "share", "netmonitor"), nil
}

func (p *LinuxPlatform) GetLogDir() (string, error) {
    dataDir, err := p.GetDataDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dataDir, "logs"), nil
}

func (p *LinuxPlatform) OpenFileManager(path string) error {
    // Try different file managers
    fileManagers := []string{"xdg-open", "nautilus", "dolphin", "thunar", "pcmanfm"}

    for _, fm := range fileManagers {
        if _, err := exec.LookPath(fm); err == nil {
            return exec.Command(fm, path).Start()
        }
    }

    return fmt.Errorf("no suitable file manager found")
}

func (p *LinuxPlatform) OpenURL(url string) error {
    return exec.Command("xdg-open", url).Start()
}

func (p *LinuxPlatform) GetSystemInfo() (*SystemInfo, error) {
    hostname, _ := os.Hostname()
    username := os.Getenv("USER")
    homeDir, _ := os.UserHomeDir()

    version, err := p.getLinuxDistribution()
    if err != nil {
        version = "Unknown Linux"
    }

    return &SystemInfo{
        OS:           "Linux",
        Version:      version,
        Architecture: runtime.GOARCH,
        Hostname:     hostname,
        Username:     username,
        HomeDir:      homeDir,
        TempDir:      os.TempDir(),
    }, nil
}

func (p *LinuxPlatform) getLinuxDistribution() (string, error) {
    // Try /etc/os-release first
    if file, err := os.Open("/etc/os-release"); err == nil {
        defer file.Close()
        scanner := bufio.NewScanner(file)
        var name, version string

        for scanner.Scan() {
            line := scanner.Text()
            if strings.HasPrefix(line, "NAME=") {
                name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
            } else if strings.HasPrefix(line, "VERSION=") {
                version = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
            }
        }

        if name != "" {
            if version != "" {
                return fmt.Sprintf("%s %s", name, version), nil
            }
            return name, nil
        }
    }

    // Fallback to lsb_release
    if out, err := exec.Command("lsb_release", "-d", "-s").Output(); err == nil {
        return strings.TrimSpace(string(out)), nil
    }

    return "Linux", nil
}

func (p *LinuxPlatform) SupportsSystemTray() bool {
    // Check if running in a desktop environment that supports system tray
    desktop := os.Getenv("XDG_CURRENT_DESKTOP")
    return desktop != ""
}

func (p *LinuxPlatform) SupportsNotifications() bool {
    // Check if D-Bus notification service is available
    _, err := exec.LookPath("notify-send")
    return err == nil
}

func (p *LinuxPlatform) GetNetworkInterfaces() ([]NetworkInterface, error) {
    return p.getLinuxNetworkInterfaces()
}

func (p *LinuxPlatform) getLinuxNetworkInterfaces() ([]NetworkInterface, error) {
    interfaces := []NetworkInterface{}

    // Read from /proc/net/dev
    file, err := os.Open("/proc/net/dev")
    if err != nil {
        return interfaces, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    // Skip header lines
    scanner.Scan()
    scanner.Scan()

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        fields := strings.Fields(line)
        if len(fields) > 0 {
            name := strings.TrimSuffix(fields[0], ":")
            if name != "lo" { // Skip loopback
                interfaces = append(interfaces, NetworkInterface{
                    Name:        name,
                    DisplayName: name,
                    Type:        p.getInterfaceType(name),
                    Status:      "unknown",
                })
            }
        }
    }

    return interfaces, nil
}

func (p *LinuxPlatform) getInterfaceType(name string) string {
    if strings.HasPrefix(name, "eth") {
        return "ethernet"
    } else if strings.HasPrefix(name, "wlan") || strings.HasPrefix(name, "wifi") {
        return "wireless"
    } else if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") {
        return "bridge"
    }
    return "unknown"
}
```

## Build System for Cross-Platform
```makefile
# Makefile for cross-platform builds
.PHONY: all clean build-windows build-macos build-linux test

APP_NAME := netmonitor
VERSION := $(shell git describe --tags --always)
BUILD_DIR := build

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

all: build-windows build-macos build-linux

clean:
	rm -rf $(BUILD_DIR)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 wails build $(LDFLAGS) -o $(BUILD_DIR)/windows/$(APP_NAME).exe
	@echo "Windows build complete"

build-macos:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)/macos
	GOOS=darwin GOARCH=amd64 wails build $(LDFLAGS) -o $(BUILD_DIR)/macos/$(APP_NAME)
	GOOS=darwin GOARCH=arm64 wails build $(LDFLAGS) -o $(BUILD_DIR)/macos/$(APP_NAME)-arm64
	@echo "macOS builds complete"

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 wails build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME)
	GOOS=linux GOARCH=arm64 wails build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME)-arm64
	@echo "Linux builds complete"

# Create packages for each platform
package-windows: build-windows
	@echo "Creating Windows installer..."
	# Use NSIS or similar to create .msi installer

package-macos: build-macos
	@echo "Creating macOS app bundle..."
	# Create .app bundle and .dmg

package-linux: build-linux
	@echo "Creating Linux packages..."
	# Create .deb, .rpm, and .tar.gz packages

test:
	go test ./...

test-integration:
	@echo "Running integration tests..."
	# Platform-specific integration tests
```

## Platform-Specific Testing
```go
// test/platform_test.go
//go:build integration

package test

import (
    "testing"
    "runtime"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPlatformSpecificFeatures(t *testing.T) {
    platform := platform.NewPlatform()

    t.Run("ConfigDirectory", func(t *testing.T) {
        configDir, err := platform.GetConfigDir()
        require.NoError(t, err)
        assert.NotEmpty(t, configDir)

        // Platform-specific assertions
        switch runtime.GOOS {
        case "windows":
            assert.Contains(t, configDir, "AppData")
        case "darwin":
            assert.Contains(t, configDir, "Library/Preferences")
        case "linux":
            assert.True(t,
                strings.Contains(configDir, ".config") ||
                strings.Contains(configDir, "XDG_CONFIG_HOME"))
        }
    })

    t.Run("SystemTraySupport", func(t *testing.T) {
        supported := platform.SupportsSystemTray()

        switch runtime.GOOS {
        case "windows", "darwin":
            assert.True(t, supported)
        case "linux":
            // Depends on desktop environment
            // Don't assert specific value
        }
    })

    t.Run("NetworkInterfaces", func(t *testing.T) {
        interfaces, err := platform.GetNetworkInterfaces()
        require.NoError(t, err)

        // Should have at least one interface
        assert.NotEmpty(t, interfaces)

        // Validate interface structure
        for _, iface := range interfaces {
            assert.NotEmpty(t, iface.Name)
            assert.NotEmpty(t, iface.Type)
        }
    })
}

func TestFileSystemPermissions(t *testing.T) {
    platform := platform.NewPlatform()

    dirs := []func() (string, error){
        platform.GetConfigDir,
        platform.GetDataDir,
        platform.GetLogDir,
        platform.GetTempDir,
    }

    for _, getDir := range dirs {
        dir, err := getDir()
        require.NoError(t, err)

        // Test creating directory
        err = os.MkdirAll(dir, 0755)
        assert.NoError(t, err)

        // Test writing file
        testFile := filepath.Join(dir, "test.txt")
        err = ioutil.WriteFile(testFile, []byte("test"), 0644)
        assert.NoError(t, err)

        // Cleanup
        os.RemoveAll(dir)
    }
}
```

## Application Integration
```go
// main.go - Platform integration
func main() {
    // Initialize platform abstraction
    platform := platform.NewPlatform()

    // Get platform-specific directories
    configDir, err := platform.GetConfigDir()
    if err != nil {
        log.Fatalf("Failed to get config directory: %v", err)
    }

    dataDir, err := platform.GetDataDir()
    if err != nil {
        log.Fatalf("Failed to get data directory: %v", err)
    }

    // Create directories if they don't exist
    for _, dir := range []string{configDir, dataDir} {
        if err := os.MkdirAll(dir, 0755); err != nil {
            log.Fatalf("Failed to create directory %s: %v", dir, err)
        }
    }

    // Initialize app with platform-specific paths
    app := &App{
        platform:  platform,
        configDir: configDir,
        dataDir:   dataDir,
    }

    // Platform-specific application options
    options := &wails.Options{
        Title:         "NetMonitor",
        Width:         1024,
        Height:        768,
        DisableResize: false,
        Fullscreen:    false,
        // Platform-specific options can be set here
    }

    // Adjust for platform conventions
    switch runtime.GOOS {
    case "darwin":
        options.TitleBarAppearsTransparent = true
        options.WebviewIsTransparent = true
    case "linux":
        options.Icon = getLinuxIcon()
    }

    err = wails.Run(options)
    if err != nil {
        log.Fatalf("Failed to start application: %v", err)
    }
}
```

## Verification Steps
1. Test builds on all platforms - should compile without errors
2. Verify file system operations - should use correct platform directories
3. Test system integration - should work with native platform features
4. Verify UI conventions - should follow platform-specific guidelines
5. Test network operations - should work correctly on all platforms
6. Verify packaging - should create appropriate installers/packages
7. Test auto-start functionality - should work on all supported platforms
8. Verify permission handling - should handle platform-specific permissions

## Dependencies
- T036: System Tray Integration
- T037: Auto-Start Functionality
- T003: Configuration System
- Wails v2 framework

## Notes
- Test on real hardware for each platform, not just virtual machines
- Consider platform-specific performance optimizations
- Implement proper error handling for platform differences
- Document platform-specific requirements and limitations
- Plan for future platform support (BSD, etc.)
- Consider using build tags for platform-specific code organization