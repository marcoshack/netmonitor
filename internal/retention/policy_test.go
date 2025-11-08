package retention

import (
	"testing"
	"time"
)

func TestDefaultRetentionPolicy(t *testing.T) {
	policy := DefaultRetentionPolicy()

	if policy.RawDataDays != 90 {
		t.Errorf("Expected default RawDataDays to be 90, got %d", policy.RawDataDays)
	}

	if policy.AggregatedDataDays != 365 {
		t.Errorf("Expected default AggregatedDataDays to be 365, got %d", policy.AggregatedDataDays)
	}

	if policy.ConfigBackupDays != 30 {
		t.Errorf("Expected default ConfigBackupDays to be 30, got %d", policy.ConfigBackupDays)
	}

	if !policy.AutoCleanupEnabled {
		t.Error("Expected default AutoCleanupEnabled to be true")
	}

	if policy.CleanupTime != "02:00" {
		t.Errorf("Expected default CleanupTime to be '02:00', got '%s'", policy.CleanupTime)
	}
}

func TestRetentionPolicyValidation(t *testing.T) {
	tests := []struct {
		name      string
		policy    *RetentionPolicy
		expectErr bool
	}{
		{
			name:      "Valid policy",
			policy:    DefaultRetentionPolicy(),
			expectErr: false,
		},
		{
			name: "RawDataDays too low",
			policy: &RetentionPolicy{
				RawDataDays:        5,
				AggregatedDataDays: 365,
				ConfigBackupDays:   30,
				AutoCleanupEnabled: true,
				CleanupTime:        "02:00",
			},
			expectErr: true,
		},
		{
			name: "RawDataDays too high",
			policy: &RetentionPolicy{
				RawDataDays:        400,
				AggregatedDataDays: 730,
				ConfigBackupDays:   30,
				AutoCleanupEnabled: true,
				CleanupTime:        "02:00",
			},
			expectErr: true,
		},
		{
			name: "AggregatedDataDays too low",
			policy: &RetentionPolicy{
				RawDataDays:        90,
				AggregatedDataDays: 5,
				ConfigBackupDays:   30,
				AutoCleanupEnabled: true,
				CleanupTime:        "02:00",
			},
			expectErr: true,
		},
		{
			name: "AggregatedDataDays too high",
			policy: &RetentionPolicy{
				RawDataDays:        90,
				AggregatedDataDays: 800,
				ConfigBackupDays:   30,
				AutoCleanupEnabled: true,
				CleanupTime:        "02:00",
			},
			expectErr: true,
		},
		{
			name: "ConfigBackupDays too low",
			policy: &RetentionPolicy{
				RawDataDays:        90,
				AggregatedDataDays: 365,
				ConfigBackupDays:   0,
				AutoCleanupEnabled: true,
				CleanupTime:        "02:00",
			},
			expectErr: true,
		},
		{
			name: "ConfigBackupDays too high",
			policy: &RetentionPolicy{
				RawDataDays:        90,
				AggregatedDataDays: 365,
				ConfigBackupDays:   400,
				AutoCleanupEnabled: true,
				CleanupTime:        "02:00",
			},
			expectErr: true,
		},
		{
			name: "Invalid cleanup time format",
			policy: &RetentionPolicy{
				RawDataDays:        90,
				AggregatedDataDays: 365,
				ConfigBackupDays:   30,
				AutoCleanupEnabled: true,
				CleanupTime:        "25:00", // Invalid hour
			},
			expectErr: true,
		},
		{
			name: "Aggregated data retention less than raw data",
			policy: &RetentionPolicy{
				RawDataDays:        90,
				AggregatedDataDays: 30,
				ConfigBackupDays:   30,
				AutoCleanupEnabled: true,
				CleanupTime:        "02:00",
			},
			expectErr: true,
		},
		{
			name: "Valid custom policy",
			policy: &RetentionPolicy{
				RawDataDays:        30,
				AggregatedDataDays: 180,
				ConfigBackupDays:   14,
				AutoCleanupEnabled: false,
				CleanupTime:        "03:30",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGetCleanupTimeToday(t *testing.T) {
	policy := &RetentionPolicy{
		RawDataDays:        90,
		AggregatedDataDays: 365,
		ConfigBackupDays:   30,
		AutoCleanupEnabled: true,
		CleanupTime:        "14:30",
	}

	cleanupTime, err := policy.GetCleanupTimeToday()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	now := time.Now()

	// Check that it's today's date
	if cleanupTime.Year() != now.Year() ||
		cleanupTime.Month() != now.Month() ||
		cleanupTime.Day() != now.Day() {
		t.Errorf("Expected today's date, got %v", cleanupTime)
	}

	// Check that time is correct
	if cleanupTime.Hour() != 14 || cleanupTime.Minute() != 30 {
		t.Errorf("Expected time 14:30, got %02d:%02d", cleanupTime.Hour(), cleanupTime.Minute())
	}
}

func TestGetNextCleanupTime(t *testing.T) {
	// Test with a time that's already passed today
	policy := &RetentionPolicy{
		RawDataDays:        90,
		AggregatedDataDays: 365,
		ConfigBackupDays:   30,
		AutoCleanupEnabled: true,
		CleanupTime:        "00:01", // Just after midnight
	}

	nextCleanup, err := policy.GetNextCleanupTime()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	now := time.Now()

	// Should be tomorrow since 00:01 has already passed
	if nextCleanup.Before(now) {
		t.Errorf("Next cleanup time should be in the future, got %v", nextCleanup)
	}

	// Test with a time that might not have passed yet
	futurePolicy := &RetentionPolicy{
		RawDataDays:        90,
		AggregatedDataDays: 365,
		ConfigBackupDays:   30,
		AutoCleanupEnabled: true,
		CleanupTime:        "23:59",
	}

	nextCleanup2, err := futurePolicy.GetNextCleanupTime()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should be either today or tomorrow, but always in the future
	if nextCleanup2.Before(now) {
		t.Errorf("Next cleanup time should be in the future, got %v", nextCleanup2)
	}
}

func TestGetCleanupTimeWithInvalidFormat(t *testing.T) {
	policy := &RetentionPolicy{
		RawDataDays:        90,
		AggregatedDataDays: 365,
		ConfigBackupDays:   30,
		AutoCleanupEnabled: true,
		CleanupTime:        "invalid",
	}

	_, err := policy.GetCleanupTimeToday()
	if err == nil {
		t.Error("Expected error for invalid cleanup time format")
	}

	_, err = policy.GetNextCleanupTime()
	if err == nil {
		t.Error("Expected error for invalid cleanup time format")
	}
}
