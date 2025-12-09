package config

import (
	"encoding/json"
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
			TestIntervalMinutes:  5,
			DataRetentionDays:    90,
			NotificationsEnabled: true,
		},
	}
}

// LoadConfig reads the configuration from the specified file path
func LoadConfig(path string) (*models.Configuration, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		cfg := DefaultConfig()
		// Attempt to save the default config so the user has a starting point
		_ = SaveConfig(path, cfg)
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg models.Configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
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
