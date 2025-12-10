package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadSaveConfig(t *testing.T) {
	tmpFile := "test_config.json"
	defer os.Remove(tmpFile)

	// Test Default Load
	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.Settings.TestIntervalSeconds != 300 {
		t.Errorf("Expected default interval 300, got %d", cfg.Settings.TestIntervalSeconds)
	}

	// Test Save
	cfg.Settings.TestIntervalSeconds = 10
	err = SaveConfig(tmpFile, cfg)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Reload
	cfg2, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	if cfg2.Settings.TestIntervalSeconds != 10 {
		t.Errorf("Expected interval 10, got %d", cfg2.Settings.TestIntervalSeconds)
	}

	if !reflect.DeepEqual(cfg, cfg2) {
		t.Errorf("Configs do not match")
	}
}
