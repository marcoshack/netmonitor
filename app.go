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

	schedulerStatus := a.monitor.GetSchedulerStatus()
	cfg := a.config.GetConfig()

	// Get region status
	regionStatus := make(map[string]*RegionStatus)
	for regionName, region := range cfg.Regions {
		regionStatus[regionName] = &RegionStatus{
			Name:          regionName,
			EndpointCount: len(region.Endpoints),
			HealthyCount:  0, // Will be updated with actual data
			WarningCount:  0,
			DownCount:     0,
		}
	}

	status := &MonitoringStatus{
		Running:         a.monitor.IsRunning(),
		StartTime:       schedulerStatus.LastRun,
		TotalEndpoints:  a.getTotalEndpointCount(),
		ActiveEndpoints: a.getTotalEndpointCount(), // All are active by default
		LastTestTime:    schedulerStatus.LastRun,
		NextTestTime:    schedulerStatus.NextRun,
		RegionStatus:    regionStatus,
	}

	return status, nil
}

// MonitoringStatus represents the current monitoring state
type MonitoringStatus struct {
	Running         bool                     `json:"running"`
	StartTime       time.Time                `json:"startTime"`
	TotalEndpoints  int                      `json:"totalEndpoints"`
	ActiveEndpoints int                      `json:"activeEndpoints"`
	LastTestTime    time.Time                `json:"lastTestTime"`
	NextTestTime    time.Time                `json:"nextTestTime"`
	RegionStatus    map[string]*RegionStatus `json:"regionStatus"`
}

// EndpointStatus represents the status of a single endpoint
type EndpointStatus struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Status           string        `json:"status"` // "up", "down", "warning"
	LastLatency      time.Duration `json:"lastLatency"`
	LastTest         time.Time     `json:"lastTest"`
	Uptime           float64       `json:"uptime"` // Percentage
	ConsecutiveFails int           `json:"consecutiveFails"`
}

// RegionStatus represents the status of a region
type RegionStatus struct {
	Name           string  `json:"name"`
	EndpointCount  int     `json:"endpointCount"`
	HealthyCount   int     `json:"healthyCount"`
	WarningCount   int     `json:"warningCount"`
	DownCount      int     `json:"downCount"`
	AverageLatency float64 `json:"averageLatency"`
	OverallHealth  string  `json:"overallHealth"`
}

// SystemHealth represents system health information
type SystemHealth struct {
	Healthy        bool              `json:"healthy"`
	Issues         []string          `json:"issues"`
	Warnings       []string          `json:"warnings"`
	StorageStatus  string            `json:"storageStatus"`
	SchedulerState string            `json:"schedulerState"`
	ConfigValid    bool              `json:"configValid"`
	Metrics        map[string]string `json:"metrics"`
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	MemoryUsageMB    float64 `json:"memoryUsageMB"`
	CPUUsagePercent  float64 `json:"cpuUsagePercent"`
	GoroutineCount   int     `json:"goroutineCount"`
	ActiveTests      int     `json:"activeTests"`
	CompletedTests   int64   `json:"completedTests"`
	StorageSizeBytes int64   `json:"storageSizeBytes"`
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

// AddEndpoint adds a new endpoint to a region
func (a *App) AddEndpoint(regionName string, endpoint *config.Endpoint) error {
	if a.config == nil {
		return fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("region", regionName).
		Str("endpoint", endpoint.Name).
		Msg("Add endpoint requested via API")

	return a.config.AddEndpoint(regionName, endpoint)
}

// UpdateEndpoint updates an existing endpoint
func (a *App) UpdateEndpoint(regionName, endpointName string, endpoint *config.Endpoint) error {
	if a.config == nil {
		return fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("region", regionName).
		Str("endpoint", endpointName).
		Msg("Update endpoint requested via API")

	return a.config.UpdateEndpoint(regionName, endpointName, endpoint)
}

// RemoveEndpoint removes an endpoint from a region
func (a *App) RemoveEndpoint(regionName, endpointName string) error {
	if a.config == nil {
		return fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("region", regionName).
		Str("endpoint", endpointName).
		Msg("Remove endpoint requested via API")

	return a.config.RemoveEndpoint(regionName, endpointName)
}

// MoveEndpoint moves an endpoint between regions
func (a *App) MoveEndpoint(sourceRegion, targetRegion, endpointName string) error {
	if a.config == nil {
		return fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("from_region", sourceRegion).
		Str("to_region", targetRegion).
		Str("endpoint", endpointName).
		Msg("Move endpoint requested via API")

	return a.config.MoveEndpoint(sourceRegion, targetRegion, endpointName)
}

// CreateRegion creates a new region
func (a *App) CreateRegion(regionName string, thresholds *config.Thresholds) error {
	if a.config == nil {
		return fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("region", regionName).
		Msg("Create region requested via API")

	return a.config.CreateRegion(regionName, thresholds)
}

// UpdateRegion updates a region's thresholds
func (a *App) UpdateRegion(regionName string, thresholds *config.Thresholds) error {
	if a.config == nil {
		return fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("region", regionName).
		Msg("Update region requested via API")

	return a.config.UpdateRegion(regionName, thresholds)
}

// RemoveRegion removes a region
func (a *App) RemoveRegion(regionName string) error {
	if a.config == nil {
		return fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("region", regionName).
		Msg("Remove region requested via API")

	return a.config.RemoveRegion(regionName)
}

// ValidateEndpoint validates an endpoint configuration
func (a *App) ValidateEndpoint(endpoint *config.Endpoint) (*config.ValidationResult, error) {
	if a.config == nil {
		return nil, fmt.Errorf("configuration manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("endpoint", endpoint.Name).
		Msg("Validate endpoint requested via API")

	return a.config.ValidateEndpointConfig(endpoint), nil
}

// GetRecentResults retrieves recent test results for an endpoint
func (a *App) GetRecentResults(endpointID string, hours int) ([]*storage.TestResult, error) {
	if a.storage == nil {
		return nil, fmt.Errorf("storage manager not initialized")
	}

	log.Ctx(a.ctx).Info().
		Str("endpoint_id", endpointID).
		Int("hours", hours).
		Msg("Recent results requested via API")

	// Get results for the last N hours
	var allResults []*storage.TestResult
	now := time.Now()

	for i := 0; i < hours/24+1; i++ {
		date := now.AddDate(0, 0, -i)
		results, err := a.storage.GetResults(date)
		if err != nil {
			log.Ctx(a.ctx).Warn().
				Err(err).
				Time("date", date).
				Msg("Failed to get results for date")
			continue
		}

		// Filter for this endpoint and time range
		cutoff := now.Add(-time.Duration(hours) * time.Hour)
		for _, result := range results {
			if result.EndpointID == endpointID && result.Timestamp.After(cutoff) {
				allResults = append(allResults, result)
			}
		}
	}

	return allResults, nil
}

// GetSystemHealth returns system health information
func (a *App) GetSystemHealth() (*SystemHealth, error) {
	health := &SystemHealth{
		Healthy:       true,
		Issues:        []string{},
		Warnings:      []string{},
		ConfigValid:   true,
		Metrics:       make(map[string]string),
	}

	// Check monitoring status
	if a.monitor != nil {
		if a.monitor.IsRunning() {
			health.SchedulerState = "running"
		} else {
			health.SchedulerState = "stopped"
			health.Warnings = append(health.Warnings, "Monitoring is not running")
		}
	} else {
		health.Healthy = false
		health.Issues = append(health.Issues, "Monitor manager not initialized")
		health.SchedulerState = "error"
	}

	// Check storage
	if a.storage != nil {
		stats, err := a.storage.GetStorageStats()
		if err != nil {
			health.Warnings = append(health.Warnings, fmt.Sprintf("Failed to get storage stats: %v", err))
			health.StorageStatus = "warning"
		} else {
			health.StorageStatus = "healthy"
			health.Metrics["storage_files"] = fmt.Sprintf("%d", stats.TotalFiles)
			health.Metrics["storage_size_mb"] = fmt.Sprintf("%.2f", float64(stats.TotalSizeBytes)/1024/1024)
		}
	} else {
		health.Healthy = false
		health.Issues = append(health.Issues, "Storage manager not initialized")
		health.StorageStatus = "error"
	}

	// Check configuration
	if a.config != nil {
		cfg := a.config.GetConfig()
		health.Metrics["total_regions"] = fmt.Sprintf("%d", len(cfg.Regions))
		health.Metrics["total_endpoints"] = fmt.Sprintf("%d", a.getTotalEndpointCount())
	} else {
		health.Healthy = false
		health.Issues = append(health.Issues, "Configuration manager not initialized")
		health.ConfigValid = false
	}

	log.Ctx(a.ctx).Info().
		Bool("healthy", health.Healthy).
		Int("issues", len(health.Issues)).
		Int("warnings", len(health.Warnings)).
		Msg("System health requested via API")

	return health, nil
}

// GetPerformanceMetrics returns performance metrics
func (a *App) GetPerformanceMetrics() (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{
		GoroutineCount: 0, // Will be updated with actual runtime data
	}

	// Get scheduler metrics
	if a.monitor != nil {
		schedulerStatus := a.monitor.GetSchedulerStatus()
		metrics.ActiveTests = schedulerStatus.ActiveTests
		metrics.CompletedTests = schedulerStatus.CompletedTests
	}

	// Get storage metrics
	if a.storage != nil {
		stats, err := a.storage.GetStorageStats()
		if err == nil {
			metrics.StorageSizeBytes = stats.TotalSizeBytes
		}
	}

	log.Ctx(a.ctx).Info().Msg("Performance metrics requested via API")

	return metrics, nil
}
