package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcoshack/netmonitor/internal/models"
)

// DefaultConfig returns a default configuration structure
func DefaultConfig() *models.Configuration {
	return &models.Configuration{
		Regions: map[string]models.Region{
			"Default": {
				Endpoints: []models.Endpoint{
					{Name: "Google DNS", Type: models.TypeICMP, Address: "8.8.8.8", Timeout: 1000},
				},
				Thresholds: models.Thresholds{
					LatencyMs:           100,
					AvailabilityPercent: 99.0,
				},
			},
		},
		Settings: models.AppSettings{
			TestIntervalSeconds:  300,
			DataRetentionDays:    90,
			NotificationsEnabled: true,
			WindowWidth:          1024,
			WindowHeight:         880,
			WindowX:              -1,
			WindowY:              -1,
		},
	}
}

// LoadConfig reads the configuration from the specified file path
func LoadConfig(path string) (*models.Configuration, []models.Notification, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		cfg := DefaultConfig()
		// Attempt to save the default config so the user has a starting point
		_ = SaveConfig(path, cfg)
		return cfg, []models.Notification{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var cfg models.Configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, nil, err
	}

	if cfg.Settings.TestIntervalSeconds < 1 {
		cfg.Settings.TestIntervalSeconds = 300
	}

	// Validation for duplicates
	seen := make(map[string]bool)
	var notifications []models.Notification

	for regionName, region := range cfg.Regions {
		var validEndpoints []models.Endpoint
		for _, ep := range region.Endpoints {
			id := fmt.Sprintf("%s:%s", ep.Type, ep.Address)
			if seen[id] {
				notifications = append(notifications, models.Notification{
					Level:   "error",
					Type:    "config",
					Message: fmt.Sprintf("Duplicate endpoint ignored: %s (%s) in region %s", ep.Name, id, regionName),
				})
			} else {
				seen[id] = true
				validEndpoints = append(validEndpoints, ep)
			}
		}
		// Update endpoints with only valid ones
		region.Endpoints = validEndpoints
		cfg.Regions[regionName] = region
	}

	return &cfg, notifications, nil
}

// SaveConfig writes the configuration to the specified file path
func SaveConfig(path string, cfg *models.Configuration) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
