package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"NetMonitor/internal/config"
	"NetMonitor/internal/monitor"
	"NetMonitor/internal/storage"
)

// App represents the main application context
type App struct {
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *slog.Logger
	config          *config.Manager
	monitor         *monitor.Manager
	storage         *storage.Manager
	mutex           sync.RWMutex
	running         bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		running: false,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	
	// Initialize logging first
	a.initializeLogging()

	a.logger.Info("NetMonitor application starting up")

	// Initialize configuration manager
	configManager, err := config.NewManager(".")
	if err != nil {
		a.logger.Error("Failed to initialize configuration manager", "error", err)
		return
	}
	a.config = configManager

	// Add configuration change callback
	a.config.AddCallback(func(cfg *config.Config) {
		a.logger.Info("Configuration changed", "regions", len(cfg.Regions))
		// TODO: Restart monitoring with new configuration
	})

	// Initialize storage manager
	storageManager, err := storage.NewManager("./data", a.logger)
	if err != nil {
		a.logger.Error("Failed to initialize storage manager", "error", err)
		return
	}
	a.storage = storageManager

	// Initialize monitoring manager
	monitorManager, err := monitor.NewManager(a.config, a.storage, a.logger)
	if err != nil {
		a.logger.Error("Failed to initialize monitor manager", "error", err)
		return
	}
	a.monitor = monitorManager

	a.mutex.Lock()
	a.running = true
	a.mutex.Unlock()

	a.logger.Info("NetMonitor application startup completed successfully")
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	a.logger.Info("NetMonitor application shutting down")

	a.mutex.Lock()
	if !a.running {
		a.mutex.Unlock()
		return nil
	}
	a.running = false
	a.mutex.Unlock()

	// Stop monitoring
	if a.monitor != nil {
		if err := a.monitor.Stop(); err != nil {
			a.logger.Error("Failed to stop monitoring", "error", err)
		}
	}

	// Close storage
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			a.logger.Error("Failed to close storage", "error", err)
		}
	}

	// Close configuration manager
	if a.config != nil {
		if err := a.config.Close(); err != nil {
			a.logger.Error("Failed to close configuration manager", "error", err)
		}
	}

	// Cancel context
	if a.cancel != nil {
		a.cancel()
	}

	a.logger.Info("NetMonitor application shutdown completed")
	return nil
}

// IsRunning returns whether the application is currently running
func (a *App) IsRunning() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.running
}

// GetSystemInfo returns basic system information
func (a *App) GetSystemInfo() (*SystemInfo, error) {
	return &SystemInfo{
		ApplicationName: "NetMonitor",
		Version:         "1.0.0",
		BuildTime:       "2025-09-27",
		Running:         a.IsRunning(),
	}, nil
}

// SystemInfo contains basic system information
type SystemInfo struct {
	ApplicationName string `json:"applicationName"`
	Version         string `json:"version"`
	BuildTime       string `json:"buildTime"`
	Running         bool   `json:"running"`
}

// initializeLogging sets up structured logging
func (a *App) initializeLogging() {
	// Create structured logger with text output for better development experience
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	a.logger = logger
}

// Greet returns a greeting for the given name (temporary method for testing)
func (a *App) Greet(name string) string {
	if a.logger != nil {
		a.logger.Info("Greet method called", "name", name)
	}
	return fmt.Sprintf("Hello %s, Welcome to NetMonitor!", name)
}

// GetConfiguration retrieves current configuration
func (a *App) GetConfiguration() (*config.Config, error) {
	if a.config == nil {
		return nil, fmt.Errorf("configuration manager not initialized")
	}
	
	cfg := a.config.GetConfig()
	a.logger.Info("Configuration retrieved", "regions", len(cfg.Regions))
	return cfg, nil
}

// SetTheme sets the application theme
func (a *App) SetTheme(theme string) error {
	a.logger.Info("Theme change requested", "theme", theme)
	
	// Validate theme
	validThemes := map[string]bool{
		"light":         true,
		"dark":          true,
		"auto":          true,
		"high-contrast": true,
	}
	
	if !validThemes[theme] {
		return fmt.Errorf("invalid theme: %s", theme)
	}
	
	// For now, just log the theme change
	// TODO: Persist theme preference in configuration
	a.logger.Info("Theme set successfully", "theme", theme)
	return nil
}

// GetMonitoringStatus returns the current monitoring status
func (a *App) GetMonitoringStatus() (*MonitoringStatus, error) {
	if a.monitor == nil {
		return nil, fmt.Errorf("monitor manager not initialized")
	}

	status := &MonitoringStatus{
		Running:       a.monitor.IsRunning(),
		LastTestTime:  "Never", // TODO: Implement actual last test time
		NextTestTime:  "5 minutes", // TODO: Calculate based on interval
		TotalEndpoints: a.getTotalEndpointCount(),
	}

	return status, nil
}

// MonitoringStatus represents the current monitoring state
type MonitoringStatus struct {
	Running        bool   `json:"running"`
	LastTestTime   string `json:"lastTestTime"`
	NextTestTime   string `json:"nextTestTime"`
	TotalEndpoints int    `json:"totalEndpoints"`
}

// getTotalEndpointCount counts total configured endpoints
func (a *App) getTotalEndpointCount() int {
	if a.config == nil {
		return 0
	}
	
	cfg := a.config.GetConfig()
	total := 0
	for _, region := range cfg.Regions {
		total += len(region.Endpoints)
	}
	
	return total
}

// StartMonitoring starts the monitoring process
func (a *App) StartMonitoring() error {
	if a.monitor == nil {
		return fmt.Errorf("monitor manager not initialized")
	}
	
	a.logger.Info("Starting monitoring via API")
	return a.monitor.Start()
}

// StopMonitoring stops the monitoring process
func (a *App) StopMonitoring() error {
	if a.monitor == nil {
		return fmt.Errorf("monitor manager not initialized")
	}
	
	a.logger.Info("Stopping monitoring via API")
	return a.monitor.Stop()
}

// RunManualTest executes a manual test for the specified endpoint
func (a *App) RunManualTest(endpointID string) (*storage.TestResult, error) {
	if a.monitor == nil {
		return nil, fmt.Errorf("monitor manager not initialized")
	}
	
	a.logger.Info("Manual test requested via API", "endpoint_id", endpointID)
	return a.monitor.RunManualTest(endpointID)
}