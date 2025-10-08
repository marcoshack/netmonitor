package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marcoshack/netmonitor/internal/aggregation"
	"github.com/marcoshack/netmonitor/internal/config"
	"github.com/marcoshack/netmonitor/internal/monitor"
	"github.com/marcoshack/netmonitor/internal/storage"
)

// App represents the main application context
type App struct {
	ctx        context.Context
	cancel     context.CancelFunc
	config     *config.Manager
	monitor    *monitor.Manager
	storage    *storage.Manager
	aggregator *aggregation.Aggregator
	mutex      sync.RWMutex
	running    bool
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

	log.Ctx(ctx).Info().Msg("NetMonitor application starting up")

	// Initialize configuration manager
	configManager, err := config.NewManager(ctx, ".")
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to initialize configuration manager")
		return
	}
	a.config = configManager

	// Add configuration change callback
	a.config.AddCallback(func(cfg *config.Config) {
		log.Ctx(ctx).Info().Int("regions", len(cfg.Regions)).Msg("Configuration changed")
		// TODO: Restart monitoring with new configuration
	})

	// Initialize storage manager
	storageManager, err := storage.NewManager(ctx, "./data")
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to initialize storage manager")
		return
	}
	a.storage = storageManager

	// Initialize aggregation manager
	aggregator := aggregation.NewAggregator(ctx, storageManager)
	a.aggregator = aggregator

	// Initialize monitoring manager
	monitorManager, err := monitor.NewManager(ctx, a.config, a.storage)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to initialize monitor manager")
		return
	}
	a.monitor = monitorManager

	a.mutex.Lock()
	a.running = true
	a.mutex.Unlock()

	log.Ctx(ctx).Info().Msg("NetMonitor application startup completed successfully")
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	log.Ctx(a.ctx).Info().Msg("NetMonitor application shutting down")

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
			log.Ctx(a.ctx).Error().Err(err).Msg("Failed to stop monitoring")
		}
	}

	// Close storage
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			log.Ctx(a.ctx).Error().Err(err).Msg("Failed to close storage")
		}
	}

	// Close configuration manager
	if a.config != nil {
		if err := a.config.Close(); err != nil {
			log.Ctx(a.ctx).Error().Err(err).Msg("Failed to close configuration manager")
		}
	}

	// Cancel context
	if a.cancel != nil {
		a.cancel()
	}

	log.Ctx(a.ctx).Info().Msg("NetMonitor application shutdown completed")
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

// Greet returns a greeting for the given name (temporary method for testing)
func (a *App) Greet(name string) string {
	log.Ctx(a.ctx).Info().Str("name", name).Msg("Greet method called")
	return fmt.Sprintf("Hello %s, Welcome to NetMonitor!", name)
}

// GetConfiguration retrieves current configuration
func (a *App) GetConfiguration() (*config.Config, error) {
	if a.config == nil {
		return nil, fmt.Errorf("configuration manager not initialized")
	}

	cfg := a.config.GetConfig()
	log.Ctx(a.ctx).Info().Int("regions", len(cfg.Regions)).Msg("Configuration retrieved")
	return cfg, nil
}

// SetTheme sets the application theme
func (a *App) SetTheme(theme string) error {
	log.Ctx(a.ctx).Info().Str("theme", theme).Msg("Theme change requested")

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
	log.Ctx(a.ctx).Info().Str("theme", theme).Msg("Theme set successfully")
	return nil
}

// GetMonitoringStatus returns the current monitoring status
func (a *App) GetMonitoringStatus() (*MonitoringStatus, error) {
	if a.monitor == nil {
		return nil, fmt.Errorf("monitor manager not initialized")
	}

	status := &MonitoringStatus{
		Running:        a.monitor.IsRunning(),
		LastTestTime:   "Never",     // TODO: Implement actual last test time
		NextTestTime:   "5 minutes", // TODO: Calculate based on interval
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

	log.Ctx(a.ctx).Info().Msg("Starting monitoring via API")
	return a.monitor.Start()
}

// StopMonitoring stops the monitoring process
func (a *App) StopMonitoring() error {
	if a.monitor == nil {
		return fmt.Errorf("monitor manager not initialized")
	}

	log.Ctx(a.ctx).Info().Msg("Stopping monitoring via API")
	return a.monitor.Stop()
}

// RunManualTest executes a manual test for the specified endpoint
func (a *App) RunManualTest(endpointID string) (*storage.TestResult, error) {
	if a.monitor == nil {
		return nil, fmt.Errorf("monitor manager not initialized")
	}

	log.Ctx(a.ctx).Info().Str("endpoint_id", endpointID).Msg("Manual test requested via API")
	result, err := a.monitor.RunManualTest(a.ctx, endpointID)

	return result, err
}

// RunManualTestDetailed executes a manual test with detailed timing information
func (a *App) RunManualTestDetailed(endpointID string) (*storage.DetailedTestResult, error) {
	if a.monitor == nil {
		return nil, fmt.Errorf("monitor manager not initialized")
	}

	log.Ctx(a.ctx).Info().Str("endpoint_id", endpointID).Msg("Detailed manual test requested via API")
	result, err := a.monitor.RunManualTestDetailed(a.ctx, endpointID)

	return result, err
}

// RunRegionTests executes manual tests for all endpoints in a region
func (a *App) RunRegionTests(regionName string) ([]*storage.DetailedTestResult, error) {
	if a.monitor == nil {
		return nil, fmt.Errorf("monitor manager not initialized")
	}

	log.Ctx(a.ctx).Info().Str("region", regionName).Msg("Region tests requested via API")
	results, err := a.monitor.RunRegionTests(a.ctx, regionName)

	return results, err
}

// RunAllTests executes manual tests for all configured endpoints
func (a *App) RunAllTests() ([]*storage.DetailedTestResult, error) {
	if a.monitor == nil {
		return nil, fmt.Errorf("monitor manager not initialized")
	}

	log.Ctx(a.ctx).Info().Msg("All tests requested via API")
	results, err := a.monitor.RunAllTests(a.ctx)

	return results, err
}

// GetAggregatedData retrieves aggregated test results for an endpoint
func (a *App) GetAggregatedData(endpointID, regionName, period string, hours int) ([]*aggregation.AggregatedResult, error) {
	if a.aggregator == nil {
		return nil, fmt.Errorf("aggregator not initialized")
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	log.Ctx(a.ctx).Info().
		Str("endpoint_id", endpointID).
		Str("period", period).
		Int("hours", hours).
		Msg("Aggregated data requested via API")

	var results []*aggregation.AggregatedResult
	var err error

	switch period {
	case "hourly":
		results, err = a.aggregator.GetHourlyAggregations(endpointID, regionName, startTime, endTime)
	case "daily":
		results, err = a.aggregator.GetDailyAggregations(endpointID, regionName, startTime, endTime)
	default:
		return nil, fmt.Errorf("invalid period: %s (must be 'hourly' or 'daily')", period)
	}

	return results, err
}
