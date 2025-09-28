package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Manager handles configuration loading and management
type Manager struct {
	config     *Config
	configPath string
	ctx        context.Context
	watcher    *fsnotify.Watcher
	mutex      sync.RWMutex
	stopChan   chan struct{}
	callbacks  []ConfigCallback
}

// ConfigCallback is called when configuration changes
type ConfigCallback func(*Config)

// Config represents the application configuration
type Config struct {
	Regions  map[string]*Region `json:"regions"`
	Settings *Settings          `json:"settings"`
}

// Region represents a geographical monitoring region
type Region struct {
	Endpoints  []*Endpoint `json:"endpoints"`
	Thresholds *Thresholds `json:"thresholds"`
}

// Endpoint represents a monitoring endpoint
type Endpoint struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
	Timeout int    `json:"timeout"`
}

// Thresholds represents alert thresholds for a region
type Thresholds struct {
	LatencyMs           int     `json:"latency_ms"`
	AvailabilityPercent float64 `json:"availability_percent"`
}

// Settings represents global application settings
type Settings struct {
	TestIntervalMinutes  int  `json:"test_interval_minutes"`
	DataRetentionDays    int  `json:"data_retention_days"`
	NotificationsEnabled bool `json:"notifications_enabled"`
}

// NewManager creates a new configuration manager
func NewManager(ctx context.Context, configDir string) (*Manager, error) {
	configPath := filepath.Join(configDir, "config.json")

	manager := &Manager{
		configPath: configPath,
		ctx:        ctx,
		stopChan:   make(chan struct{}),
		callbacks:  make([]ConfigCallback, 0),
	}

	// Initialize file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	manager.watcher = watcher

	// Load existing config or create default
	if err := manager.load(); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to load configuration, using defaults")
		// If loading fails, create default config
		manager.config = manager.getDefaultConfig()
		if err := manager.save(); err != nil {
			return nil, fmt.Errorf("failed to save default configuration: %w", err)
		}
	}

	// Start file watching
	go manager.watchConfig()

	return manager, nil
}

// AddCallback adds a callback to be called when configuration changes
func (m *Manager) AddCallback(callback ConfigCallback) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.callbacks = append(m.callbacks, callback)
}

// GetConfig returns a copy of the current configuration
func (m *Manager) GetConfig() *Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Return a deep copy
	data, _ := json.Marshal(m.config)
	var copy Config
	json.Unmarshal(data, &copy)
	return &copy
}

// UpdateConfig updates the configuration
func (m *Manager) UpdateConfig(config *Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate configuration
	if err := m.validateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	oldConfig := m.config
	m.config = config
	
	if err := m.save(); err != nil {
		// Rollback on save failure
		m.config = oldConfig
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	log.Ctx(m.ctx).Info().Msg("Configuration updated successfully")
	
	// Notify callbacks
	go m.notifyCallbacks(config)

	return nil
}

// Close stops the configuration manager and releases resources
func (m *Manager) Close() error {
	close(m.stopChan)
	
	if m.watcher != nil {
		return m.watcher.Close()
	}
	
	return nil
}

// watchConfig watches for configuration file changes
func (m *Manager) watchConfig() {
	// Add the config file to watcher
	if err := m.watcher.Add(m.configPath); err != nil {
		log.Ctx(m.ctx).Error().Err(err).Msg("Failed to watch config file")
		return
	}

	log.Ctx(m.ctx).Info().Str("path", m.configPath).Msg("Configuration file watcher started")

	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			
			if event.Has(fsnotify.Write) {
				log.Ctx(m.ctx).Info().Str("path", event.Name).Msg("Configuration file changed, reloading")
				if err := m.reload(); err != nil {
					log.Ctx(m.ctx).Error().Err(err).Msg("Failed to reload configuration")
				}
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Ctx(m.ctx).Error().Err(err).Msg("Configuration file watcher error")

		case <-m.stopChan:
			log.Ctx(m.ctx).Info().Msg("Configuration file watcher stopped")
			return
		}
	}
}

// reload reloads configuration from file
func (m *Manager) reload() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := m.load(); err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Notify callbacks
	go m.notifyCallbacks(m.config)
	
	return nil
}

// notifyCallbacks notifies all registered callbacks
func (m *Manager) notifyCallbacks(config *Config) {
	for _, callback := range m.callbacks {
		callback(config)
	}
}

// load loads configuration from file
func (m *Manager) load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Validate loaded configuration
	if err := m.validateConfig(&config); err != nil {
		return fmt.Errorf("loaded configuration is invalid: %w", err)
	}

	m.config = &config
	return nil
}

// save saves configuration to file
func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// getDefaultConfig returns the default configuration
func (m *Manager) getDefaultConfig() *Config {
	return &Config{
		Regions: map[string]*Region{
			"NA-East": {
				Endpoints: []*Endpoint{
					{
						Name:    "Google DNS",
						Type:    "ICMP",
						Address: "8.8.8.8",
						Timeout: 5000,
					},
					{
						Name:    "Cloudflare HTTP",
						Type:    "HTTP", 
						Address: "https://1.1.1.1",
						Timeout: 10000,
					},
				},
				Thresholds: &Thresholds{
					LatencyMs:           100,
					AvailabilityPercent: 99.0,
				},
			},
			"EU-West": {
				Endpoints: []*Endpoint{
					{
						Name:    "Cloudflare DNS",
						Type:    "ICMP",
						Address: "1.1.1.1",
						Timeout: 5000,
					},
				},
				Thresholds: &Thresholds{
					LatencyMs:           150,
					AvailabilityPercent: 98.5,
				},
			},
		},
		Settings: &Settings{
			TestIntervalMinutes:  5,
			DataRetentionDays:    90,
			NotificationsEnabled: true,
		},
	}
}

// validateConfig validates the configuration
func (m *Manager) validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if config.Settings == nil {
		return fmt.Errorf("settings section is required")
	}

	// Validate settings
	if err := m.validateSettings(config.Settings); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	// Validate regions
	if len(config.Regions) == 0 {
		return fmt.Errorf("at least one region must be configured")
	}

	for regionName, region := range config.Regions {
		if err := m.validateRegion(regionName, region); err != nil {
			return fmt.Errorf("invalid region '%s': %w", regionName, err)
		}
	}

	return nil
}

// validateSettings validates the settings section
func (m *Manager) validateSettings(settings *Settings) error {
	if settings.TestIntervalMinutes < 1 || settings.TestIntervalMinutes > 1440 {
		return fmt.Errorf("test interval must be between 1 and 1440 minutes")
	}

	if settings.DataRetentionDays < 1 || settings.DataRetentionDays > 365 {
		return fmt.Errorf("data retention must be between 1 and 365 days")
	}

	return nil
}

// validateRegion validates a region configuration
func (m *Manager) validateRegion(regionName string, region *Region) error {
	if region == nil {
		return fmt.Errorf("region cannot be nil")
	}

	if len(region.Endpoints) == 0 {
		return fmt.Errorf("region must have at least one endpoint")
	}

	if region.Thresholds == nil {
		return fmt.Errorf("region thresholds are required")
	}

	// Validate thresholds
	if region.Thresholds.LatencyMs < 1 || region.Thresholds.LatencyMs > 10000 {
		return fmt.Errorf("latency threshold must be between 1 and 10000 milliseconds")
	}

	if region.Thresholds.AvailabilityPercent < 50.0 || region.Thresholds.AvailabilityPercent > 100.0 {
		return fmt.Errorf("availability threshold must be between 50.0 and 100.0 percent")
	}

	// Validate endpoints
	endpointNames := make(map[string]bool)
	for _, endpoint := range region.Endpoints {
		if err := m.validateEndpoint(endpoint); err != nil {
			return fmt.Errorf("invalid endpoint '%s': %w", endpoint.Name, err)
		}

		// Check for duplicate names
		if endpointNames[endpoint.Name] {
			return fmt.Errorf("duplicate endpoint name '%s'", endpoint.Name)
		}
		endpointNames[endpoint.Name] = true
	}

	return nil
}

// validateEndpoint validates an endpoint configuration
func (m *Manager) validateEndpoint(endpoint *Endpoint) error {
	if endpoint.Name == "" {
		return fmt.Errorf("endpoint name cannot be empty")
	}

	if len(endpoint.Name) > 100 {
		return fmt.Errorf("endpoint name cannot exceed 100 characters")
	}

	if endpoint.Address == "" {
		return fmt.Errorf("endpoint address cannot be empty")
	}

	// Validate endpoint type
	validTypes := map[string]bool{
		"HTTP": true,
		"TCP":  true,
		"UDP":  true,
		"ICMP": true,
	}

	if !validTypes[endpoint.Type] {
		return fmt.Errorf("invalid endpoint type '%s', must be one of: HTTP, TCP, UDP, ICMP", endpoint.Type)
	}

	// Validate timeout
	if endpoint.Timeout < 1000 || endpoint.Timeout > 60000 {
		return fmt.Errorf("endpoint timeout must be between 1000 and 60000 milliseconds")
	}

	// TODO: Add more specific validation based on endpoint type
	// For example, validate URL format for HTTP endpoints, host:port for TCP/UDP

	return nil
}

// GenerateDefaultConfig creates and saves a default configuration file
func (m *Manager) GenerateDefaultConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.config = m.getDefaultConfig()
	
	if err := m.save(); err != nil {
		return fmt.Errorf("failed to save default configuration: %w", err)
	}

	log.Ctx(m.ctx).Info().Str("path", m.configPath).Msg("Default configuration generated")
	return nil
}

// GetConfigPath returns the configuration file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}