package aggregation

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/marcoshack/netmonitor/internal/storage"
)

func setupTestAggregator(t *testing.T) (*Aggregator, *storage.Manager, context.Context) {
	ctx := context.Background()

	// Create temporary directory for storage
	storageMgr, err := storage.NewManager(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	aggregator := NewAggregator(ctx, storageMgr)

	return aggregator, storageMgr, ctx
}

func TestNewAggregator(t *testing.T) {
	aggregator, _, _ := setupTestAggregator(t)

	if aggregator == nil {
		t.Fatal("Expected aggregator to be created")
	}

	if aggregator.cache == nil {
		t.Fatal("Expected cache to be initialized")
	}
}

func TestAggregateHourly(t *testing.T) {
	aggregator, storageMgr, _ := setupTestAggregator(t)

	// Create test results
	now := time.Now().Truncate(time.Hour)
	endpointID := "test-endpoint"
	regionName := "test-region"

	// Add successful test results
	for i := 0; i < 5; i++ {
		result := &storage.TestResult{
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
			EndpointID: endpointID,
			Protocol:   "ICMP",
			Latency:    time.Duration(10+i) * time.Millisecond,
			Status:     "success",
		}
		if err := storageMgr.StoreTestResult(result); err != nil {
			t.Fatalf("Failed to store test result: %v", err)
		}
	}

	// Add failed test result
	failedResult := &storage.TestResult{
		Timestamp:  now.Add(30 * time.Minute),
		EndpointID: endpointID,
		Protocol:   "ICMP",
		Latency:    0,
		Status:     "failed",
		Error:      "timeout",
	}
	if err := storageMgr.StoreTestResult(failedResult); err != nil {
		t.Fatalf("Failed to store failed test result: %v", err)
	}

	// Aggregate hourly
	aggregated, err := aggregator.AggregateHourly(endpointID, regionName, now)
	if err != nil {
		t.Fatalf("Failed to aggregate hourly: %v", err)
	}

	// Verify aggregation
	if aggregated.TestCount != 6 {
		t.Errorf("Expected 6 tests, got %d", aggregated.TestCount)
	}

	if aggregated.SuccessCount != 5 {
		t.Errorf("Expected 5 successful tests, got %d", aggregated.SuccessCount)
	}

	if aggregated.FailureCount != 1 {
		t.Errorf("Expected 1 failed test, got %d", aggregated.FailureCount)
	}

	expectedAvg := 12.0 // (10+11+12+13+14)/5
	if math.Abs(aggregated.AvgLatency-expectedAvg) > 0.1 {
		t.Errorf("Expected average latency %.2f, got %.2f", expectedAvg, aggregated.AvgLatency)
	}

	if aggregated.MinLatency != 10.0 {
		t.Errorf("Expected min latency 10.0, got %.2f", aggregated.MinLatency)
	}

	if aggregated.MaxLatency != 14.0 {
		t.Errorf("Expected max latency 14.0, got %.2f", aggregated.MaxLatency)
	}

	expectedAvailability := (5.0 / 6.0) * 100.0
	if math.Abs(aggregated.AvailabilityPct-expectedAvailability) > 0.1 {
		t.Errorf("Expected availability %.2f%%, got %.2f%%", expectedAvailability, aggregated.AvailabilityPct)
	}

	if aggregated.Period != "hourly" {
		t.Errorf("Expected period 'hourly', got '%s'", aggregated.Period)
	}
}

func TestAggregateDaily(t *testing.T) {
	aggregator, storageMgr, _ := setupTestAggregator(t)

	// Create test results for a day
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endpointID := "test-endpoint"
	regionName := "test-region"

	// Add test results throughout the day
	for i := 0; i < 24; i++ {
		result := &storage.TestResult{
			Timestamp:  dayStart.Add(time.Duration(i) * time.Hour),
			EndpointID: endpointID,
			Protocol:   "HTTP",
			Latency:    time.Duration(20+i) * time.Millisecond,
			Status:     "success",
		}
		if err := storageMgr.StoreTestResult(result); err != nil {
			t.Fatalf("Failed to store test result: %v", err)
		}
	}

	// Aggregate daily
	aggregated, err := aggregator.AggregateDaily(endpointID, regionName, dayStart)
	if err != nil {
		t.Fatalf("Failed to aggregate daily: %v", err)
	}

	// Verify aggregation
	if aggregated.TestCount != 24 {
		t.Errorf("Expected 24 tests, got %d", aggregated.TestCount)
	}

	if aggregated.SuccessCount != 24 {
		t.Errorf("Expected 24 successful tests, got %d", aggregated.SuccessCount)
	}

	if aggregated.FailureCount != 0 {
		t.Errorf("Expected 0 failed tests, got %d", aggregated.FailureCount)
	}

	if aggregated.AvailabilityPct != 100.0 {
		t.Errorf("Expected 100%% availability, got %.2f%%", aggregated.AvailabilityPct)
	}

	if aggregated.Period != "daily" {
		t.Errorf("Expected period 'daily', got '%s'", aggregated.Period)
	}
}

func TestAggregateEmptyResults(t *testing.T) {
	aggregator, _, _ := setupTestAggregator(t)

	now := time.Now()
	endpointID := "nonexistent-endpoint"
	regionName := "test-region"

	// Aggregate hourly with no results
	aggregated, err := aggregator.AggregateHourly(endpointID, regionName, now)
	if err != nil {
		t.Fatalf("Failed to aggregate empty results: %v", err)
	}

	if aggregated.TestCount != 0 {
		t.Errorf("Expected 0 tests, got %d", aggregated.TestCount)
	}

	if aggregated.SuccessCount != 0 {
		t.Errorf("Expected 0 successful tests, got %d", aggregated.SuccessCount)
	}

	if aggregated.AvgLatency != 0 {
		t.Errorf("Expected 0 average latency, got %.2f", aggregated.AvgLatency)
	}

	if aggregated.AvailabilityPct != 0 {
		t.Errorf("Expected 0%% availability, got %.2f%%", aggregated.AvailabilityPct)
	}
}

func TestAggregateStandardDeviation(t *testing.T) {
	aggregator, storageMgr, _ := setupTestAggregator(t)

	now := time.Now().Truncate(time.Hour)
	endpointID := "test-endpoint"
	regionName := "test-region"

	// Add test results with known latencies
	latencies := []int{10, 20, 30, 40, 50} // ms
	for i, latency := range latencies {
		result := &storage.TestResult{
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
			EndpointID: endpointID,
			Protocol:   "ICMP",
			Latency:    time.Duration(latency) * time.Millisecond,
			Status:     "success",
		}
		if err := storageMgr.StoreTestResult(result); err != nil {
			t.Fatalf("Failed to store test result: %v", err)
		}
	}

	aggregated, err := aggregator.AggregateHourly(endpointID, regionName, now)
	if err != nil {
		t.Fatalf("Failed to aggregate: %v", err)
	}

	// Calculate expected standard deviation
	// Mean = 30
	// Variance = ((20^2 + 10^2 + 0 + 10^2 + 20^2) / 5) = (400 + 100 + 0 + 100 + 400) / 5 = 200
	// StdDev = sqrt(200) ≈ 14.14
	expectedStdDev := math.Sqrt(200.0)

	if math.Abs(aggregated.StdDevLatency-expectedStdDev) > 0.1 {
		t.Errorf("Expected standard deviation %.2f, got %.2f", expectedStdDev, aggregated.StdDevLatency)
	}
}

func TestGetHourlyAggregations(t *testing.T) {
	aggregator, storageMgr, _ := setupTestAggregator(t)

	now := time.Now().Truncate(time.Hour)
	endpointID := "test-endpoint"
	regionName := "test-region"

	// Add test results for 3 hours
	for hour := 0; hour < 3; hour++ {
		for i := 0; i < 5; i++ {
			result := &storage.TestResult{
				Timestamp:  now.Add(time.Duration(hour)*time.Hour + time.Duration(i)*time.Minute),
				EndpointID: endpointID,
				Protocol:   "ICMP",
				Latency:    time.Duration(10+i) * time.Millisecond,
				Status:     "success",
			}
			if err := storageMgr.StoreTestResult(result); err != nil {
				t.Fatalf("Failed to store test result: %v", err)
			}
		}
	}

	// Get hourly aggregations (exclusive end)
	endTime := now.Add(3 * time.Hour).Add(-1 * time.Second)
	aggregations, err := aggregator.GetHourlyAggregations(endpointID, regionName, now, endTime)
	if err != nil {
		t.Fatalf("Failed to get hourly aggregations: %v", err)
	}

	if len(aggregations) != 3 {
		t.Errorf("Expected 3 hourly aggregations, got %d", len(aggregations))
	}

	for i, agg := range aggregations {
		if agg.TestCount != 5 {
			t.Errorf("Hour %d: Expected 5 tests per hour, got %d", i, agg.TestCount)
		}
		if agg.Period != "hourly" {
			t.Errorf("Hour %d: Expected period 'hourly', got '%s'", i, agg.Period)
		}
	}
}

func TestAggregationCache(t *testing.T) {
	aggregator, storageMgr, _ := setupTestAggregator(t)

	now := time.Now().Truncate(time.Hour)
	endpointID := "test-endpoint"
	regionName := "test-region"

	// Add test result
	result := &storage.TestResult{
		Timestamp:  now,
		EndpointID: endpointID,
		Protocol:   "ICMP",
		Latency:    10 * time.Millisecond,
		Status:     "success",
	}
	if err := storageMgr.StoreTestResult(result); err != nil {
		t.Fatalf("Failed to store test result: %v", err)
	}

	// First aggregation
	agg1, err := aggregator.AggregateHourly(endpointID, regionName, now)
	if err != nil {
		t.Fatalf("Failed to aggregate: %v", err)
	}

	// Second aggregation should come from cache
	agg2, err := aggregator.AggregateHourly(endpointID, regionName, now)
	if err != nil {
		t.Fatalf("Failed to aggregate: %v", err)
	}

	// Should be the same instance (from cache)
	if agg1 != agg2 {
		t.Error("Expected second aggregation to come from cache")
	}

	// Clear cache
	aggregator.ClearCache()

	// Third aggregation should be recalculated
	agg3, err := aggregator.AggregateHourly(endpointID, regionName, now)
	if err != nil {
		t.Fatalf("Failed to aggregate: %v", err)
	}

	// Should be a different instance (cache was cleared)
	if agg1 == agg3 {
		t.Error("Expected third aggregation to be recalculated after cache clear")
	}
}

func TestInvalidateEndpoint(t *testing.T) {
	aggregator, storageMgr, _ := setupTestAggregator(t)

	now := time.Now().Truncate(time.Hour)
	endpointID1 := "endpoint-1"
	endpointID2 := "endpoint-2"
	regionName := "test-region"

	// Add test results for two endpoints
	for _, epID := range []string{endpointID1, endpointID2} {
		result := &storage.TestResult{
			Timestamp:  now,
			EndpointID: epID,
			Protocol:   "ICMP",
			Latency:    10 * time.Millisecond,
			Status:     "success",
		}
		if err := storageMgr.StoreTestResult(result); err != nil {
			t.Fatalf("Failed to store test result: %v", err)
		}

		// Aggregate to populate cache
		if _, err := aggregator.AggregateHourly(epID, regionName, now); err != nil {
			t.Fatalf("Failed to aggregate: %v", err)
		}
	}

	// Cache should have entries for both endpoints
	if len(aggregator.cache) != 2 {
		t.Errorf("Expected 2 cache entries, got %d", len(aggregator.cache))
	}

	// Invalidate endpoint 1
	aggregator.InvalidateEndpoint(endpointID1)

	// Cache should only have entry for endpoint 2
	if len(aggregator.cache) != 1 {
		t.Errorf("Expected 1 cache entry after invalidation, got %d", len(aggregator.cache))
	}

	// Verify endpoint 2 is still in cache
	_, err := aggregator.AggregateHourly(endpointID2, regionName, now)
	if err != nil {
		t.Fatalf("Failed to aggregate endpoint 2: %v", err)
	}
}
