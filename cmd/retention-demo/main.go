package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/marcoshack/netmonitor/internal/retention"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx := log.Logger.WithContext(context.Background())

	fmt.Println("=== Data Retention Management Demo ===\n")

	// Create temporary directory for demo
	tempDir, err := os.MkdirTemp("", "retention-demo-*")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create temp directory")
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Demo directory: %s\n\n", tempDir)

	// 1. Create retention manager with custom policy
	fmt.Println("1. Creating retention manager with 30-day retention policy...")
	policy := &retention.RetentionPolicy{
		RawDataDays:        30,
		AggregatedDataDays: 180,
		ConfigBackupDays:   14,
		AutoCleanupEnabled: false, // Disable for demo
		CleanupTime:        "02:00",
	}

	manager, err := retention.NewManager(ctx, tempDir, policy)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create retention manager")
	}
	defer manager.Close()

	fmt.Printf("   ✓ Retention policy: %d days for raw data, %d days for aggregated data\n\n",
		policy.RawDataDays, policy.AggregatedDataDays)

	// 2. Create test data files with various ages
	fmt.Println("2. Creating test data files...")
	now := time.Now()

	testDates := []struct {
		daysAgo int
		desc    string
	}{
		{5, "Recent (5 days old)"},
		{20, "Recent (20 days old)"},
		{35, "Old (35 days old - should be deleted)"},
		{50, "Very old (50 days old - should be deleted)"},
		{90, "Ancient (90 days old - should be deleted)"},
	}

	for _, td := range testDates {
		date := now.AddDate(0, 0, -td.daysAgo).Format("2006-01-02")
		filename := filepath.Join(tempDir, date+".json")
		content := fmt.Sprintf(`{"date":"%s","results":[]}`, date)

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Fatal().Err(err).Str("file", filename).Msg("Failed to create test file")
		}

		fmt.Printf("   - Created: %s (%s)\n", date+".json", td.desc)
	}

	fmt.Println()

	// 3. Get storage stats before cleanup
	fmt.Println("3. Storage statistics before cleanup:")
	stats, err := manager.GetStorageStats()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get storage stats")
	}

	fmt.Printf("   Total files: %d\n", stats.TotalFiles)
	fmt.Printf("   Total size: %d bytes\n", stats.TotalSizeBytes)
	fmt.Printf("   Oldest data: %s\n", stats.OldestDataDate.Format("2006-01-02"))
	fmt.Printf("   Newest data: %s\n", stats.NewestDataDate.Format("2006-01-02"))
	fmt.Printf("   Days of data: %d\n\n", stats.DaysOfData)

	// 4. Trigger manual cleanup
	fmt.Println("4. Triggering manual cleanup (retaining last 30 days)...")
	report, err := manager.TriggerManualCleanup()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to trigger cleanup")
	}

	fmt.Printf("   Files deleted: %d\n", report.FilesDeleted)
	fmt.Printf("   Space freed: %d bytes\n", report.SpaceFreed)
	fmt.Printf("   Duration: %v\n", report.EndTime.Sub(report.StartTime))
	fmt.Printf("   Errors: %d\n\n", report.ErrorCount)

	// 5. Get storage stats after cleanup
	fmt.Println("5. Storage statistics after cleanup:")
	stats, err = manager.GetStorageStats()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get storage stats")
	}

	fmt.Printf("   Total files: %d\n", stats.TotalFiles)
	fmt.Printf("   Total size: %d bytes\n", stats.TotalSizeBytes)
	fmt.Printf("   Oldest data: %s\n", stats.OldestDataDate.Format("2006-01-02"))
	fmt.Printf("   Newest data: %s\n", stats.NewestDataDate.Format("2006-01-02"))
	fmt.Printf("   Days of data: %d\n\n", stats.DaysOfData)

	// 6. Test policy update
	fmt.Println("6. Updating retention policy to 15 days...")
	newPolicy := &retention.RetentionPolicy{
		RawDataDays:        15,
		AggregatedDataDays: 90,
		ConfigBackupDays:   7,
		AutoCleanupEnabled: false,
		CleanupTime:        "03:00",
	}

	if err := manager.UpdatePolicy(newPolicy); err != nil {
		log.Fatal().Err(err).Msg("Failed to update policy")
	}

	fmt.Printf("   ✓ Policy updated to %d days retention\n\n", newPolicy.RawDataDays)

	// 7. Trigger another cleanup with new policy
	fmt.Println("7. Triggering cleanup with new 15-day policy...")
	report, err = manager.TriggerManualCleanup()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to trigger cleanup")
	}

	fmt.Printf("   Files deleted: %d\n", report.FilesDeleted)
	fmt.Printf("   Space freed: %d bytes\n\n", report.SpaceFreed)

	// 8. Get final storage stats
	fmt.Println("8. Final storage statistics:")
	stats, err = manager.GetStorageStats()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get storage stats")
	}

	fmt.Printf("   Total files: %d\n", stats.TotalFiles)
	fmt.Printf("   Total size: %d bytes\n", stats.TotalSizeBytes)
	if stats.TotalFiles > 0 {
		fmt.Printf("   Oldest data: %s\n", stats.OldestDataDate.Format("2006-01-02"))
		fmt.Printf("   Newest data: %s\n", stats.NewestDataDate.Format("2006-01-02"))
		fmt.Printf("   Days of data: %d\n", stats.DaysOfData)
	}
	fmt.Println()

	// 9. Show cleanup history
	fmt.Println("9. Cleanup operation history:")
	history := manager.GetCleanupHistory()
	for i, op := range history {
		fmt.Printf("   Operation %d:\n", i+1)
		fmt.Printf("     Timestamp: %s\n", op.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("     Files deleted: %d\n", op.FilesDeleted)
		fmt.Printf("     Space freed: %d bytes\n", op.SpaceFreed)
		fmt.Printf("     Duration: %d ms\n", op.Duration)
		fmt.Printf("     Success: %v\n", op.Success)
		if op.ErrorMessage != "" {
			fmt.Printf("     Error: %s\n", op.ErrorMessage)
		}
		fmt.Println()
	}

	fmt.Println("=== Demo completed successfully! ===")
}
