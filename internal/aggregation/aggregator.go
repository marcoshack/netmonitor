package aggregation

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marcoshack/netmonitor/internal/storage"
)

// AggregatedResult represents aggregated test results for a time period
type AggregatedResult struct {
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	Period          string    `json:"period"`          // "hourly" or "daily"
	EndpointID      string    `json:"endpointID"`
	RegionName      string    `json:"regionName"`
	TestCount       int       `json:"testCount"`
	SuccessCount    int       `json:"successCount"`
	FailureCount    int       `json:"failureCount"`
	AvgLatency      float64   `json:"avgLatency"`      // in milliseconds
	MinLatency      float64   `json:"minLatency"`      // in milliseconds
	MaxLatency      float64   `json:"maxLatency"`      // in milliseconds
	StdDevLatency   float64   `json:"stdDevLatency"`   // in milliseconds
	AvailabilityPct float64   `json:"availabilityPct"` // percentage
}

// Aggregator handles result aggregation
type Aggregator struct {
	storage *storage.Manager
	ctx     context.Context
	mutex   sync.RWMutex
	cache   map[string]*AggregatedResult // Cache for current period aggregations
}

// NewAggregator creates a new aggregator
func NewAggregator(ctx context.Context, storageMgr *storage.Manager) *Aggregator {
	return &Aggregator{
		storage: storageMgr,
		ctx:     ctx,
		cache:   make(map[string]*AggregatedResult),
	}
}

// AggregateHourly aggregates test results for a specific hour
func (a *Aggregator) AggregateHourly(endpointID, regionName string, startTime time.Time) (*AggregatedResult, error) {
	// Normalize to hour boundary
	hourStart := startTime.Truncate(time.Hour)
	hourEnd := hourStart.Add(time.Hour)

	log.Ctx(a.ctx).Debug().
		Str("endpoint_id", endpointID).
		Time("start", hourStart).
		Time("end", hourEnd).
		Msg("Aggregating hourly results")

	return a.aggregate(endpointID, regionName, hourStart, hourEnd, "hourly")
}

// AggregateDaily aggregates test results for a specific day
func (a *Aggregator) AggregateDaily(endpointID, regionName string, startTime time.Time) (*AggregatedResult, error) {
	// Normalize to day boundary
	dayStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	log.Ctx(a.ctx).Debug().
		Str("endpoint_id", endpointID).
		Time("start", dayStart).
		Time("end", dayEnd).
		Msg("Aggregating daily results")

	return a.aggregate(endpointID, regionName, dayStart, dayEnd, "daily")
}

// AggregateRange aggregates test results for a custom time range
func (a *Aggregator) AggregateRange(endpointID, regionName string, startTime, endTime time.Time, period string) (*AggregatedResult, error) {
	return a.aggregate(endpointID, regionName, startTime, endTime, period)
}

// aggregate performs the actual aggregation
func (a *Aggregator) aggregate(endpointID, regionName string, startTime, endTime time.Time, period string) (*AggregatedResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", endpointID, period, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	a.mutex.RLock()
	if cached, exists := a.cache[cacheKey]; exists {
		a.mutex.RUnlock()
		return cached, nil
	}
	a.mutex.RUnlock()

	// Fetch results for the time range
	results, err := a.getResultsInRange(endpointID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get results: %w", err)
	}

	if len(results) == 0 {
		// Return empty aggregation
		return &AggregatedResult{
			StartTime:       startTime,
			EndTime:         endTime,
			Period:          period,
			EndpointID:      endpointID,
			RegionName:      regionName,
			TestCount:       0,
			SuccessCount:    0,
			FailureCount:    0,
			AvgLatency:      0,
			MinLatency:      0,
			MaxLatency:      0,
			StdDevLatency:   0,
			AvailabilityPct: 0,
		}, nil
	}

	// Calculate statistics
	var (
		successCount  int
		failureCount  int
		totalLatency  float64
		minLatency    = math.MaxFloat64
		maxLatency    = 0.0
		latencies     []float64
	)

	for _, result := range results {
		latencyMs := float64(result.Latency.Nanoseconds()) / 1_000_000.0

		if result.Status == "success" {
			successCount++
			totalLatency += latencyMs
			latencies = append(latencies, latencyMs)

			if latencyMs < minLatency {
				minLatency = latencyMs
			}
			if latencyMs > maxLatency {
				maxLatency = latencyMs
			}
		} else {
			failureCount++
		}
	}

	// Calculate average
	avgLatency := 0.0
	if successCount > 0 {
		avgLatency = totalLatency / float64(successCount)
	}

	// Calculate standard deviation
	stdDev := 0.0
	if successCount > 1 {
		var sumSquaredDiff float64
		for _, latency := range latencies {
			diff := latency - avgLatency
			sumSquaredDiff += diff * diff
		}
		variance := sumSquaredDiff / float64(successCount)
		stdDev = math.Sqrt(variance)
	}

	// Calculate availability percentage
	availability := 0.0
	if len(results) > 0 {
		availability = (float64(successCount) / float64(len(results))) * 100.0
	}

	// Handle edge case where there are no successful tests
	if successCount == 0 {
		minLatency = 0
		maxLatency = 0
	}

	aggregated := &AggregatedResult{
		StartTime:       startTime,
		EndTime:         endTime,
		Period:          period,
		EndpointID:      endpointID,
		RegionName:      regionName,
		TestCount:       len(results),
		SuccessCount:    successCount,
		FailureCount:    failureCount,
		AvgLatency:      avgLatency,
		MinLatency:      minLatency,
		MaxLatency:      maxLatency,
		StdDevLatency:   stdDev,
		AvailabilityPct: availability,
	}

	// Cache the result
	a.mutex.Lock()
	a.cache[cacheKey] = aggregated
	a.mutex.Unlock()

	return aggregated, nil
}

// getResultsInRange retrieves test results within a time range
func (a *Aggregator) getResultsInRange(endpointID string, startTime, endTime time.Time) ([]*storage.TestResult, error) {
	var allResults []*storage.TestResult

	// Iterate through each day in the range
	currentDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	endDate := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, endTime.Location())

	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		results, err := a.storage.GetResults(currentDate)
		if err != nil {
			// Log error but continue with other dates
			log.Ctx(a.ctx).Warn().
				Err(err).
				Time("date", currentDate).
				Msg("Failed to get results for date")
		} else {
			// Filter results for this endpoint and time range
			for _, result := range results {
				if result.EndpointID == endpointID &&
					!result.Timestamp.Before(startTime) &&
					result.Timestamp.Before(endTime) {
					allResults = append(allResults, result)
				}
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return allResults, nil
}

// GetHourlyAggregations returns hourly aggregations for a time range
func (a *Aggregator) GetHourlyAggregations(endpointID, regionName string, startTime, endTime time.Time) ([]*AggregatedResult, error) {
	var aggregations []*AggregatedResult

	// Normalize to hour boundaries
	currentHour := startTime.Truncate(time.Hour)
	endHour := endTime.Truncate(time.Hour)

	for currentHour.Before(endHour) || currentHour.Equal(endHour) {
		aggregated, err := a.AggregateHourly(endpointID, regionName, currentHour)
		if err != nil {
			log.Ctx(a.ctx).Warn().
				Err(err).
				Time("hour", currentHour).
				Msg("Failed to aggregate hourly results")
		} else {
			aggregations = append(aggregations, aggregated)
		}

		currentHour = currentHour.Add(time.Hour)
	}

	return aggregations, nil
}

// GetDailyAggregations returns daily aggregations for a time range
func (a *Aggregator) GetDailyAggregations(endpointID, regionName string, startTime, endTime time.Time) ([]*AggregatedResult, error) {
	var aggregations []*AggregatedResult

	// Normalize to day boundaries
	currentDay := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	endDay := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, endTime.Location())

	for currentDay.Before(endDay) || currentDay.Equal(endDay) {
		aggregated, err := a.AggregateDaily(endpointID, regionName, currentDay)
		if err != nil {
			log.Ctx(a.ctx).Warn().
				Err(err).
				Time("day", currentDay).
				Msg("Failed to aggregate daily results")
		} else {
			aggregations = append(aggregations, aggregated)
		}

		currentDay = currentDay.AddDate(0, 0, 1)
	}

	return aggregations, nil
}

// ClearCache clears the aggregation cache
func (a *Aggregator) ClearCache() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.cache = make(map[string]*AggregatedResult)
	log.Ctx(a.ctx).Debug().Msg("Aggregation cache cleared")
}

// InvalidateEndpoint removes cached aggregations for a specific endpoint
func (a *Aggregator) InvalidateEndpoint(endpointID string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for key := range a.cache {
		if len(key) > len(endpointID) && key[:len(endpointID)] == endpointID {
			delete(a.cache, key)
		}
	}

	log.Ctx(a.ctx).Debug().
		Str("endpoint_id", endpointID).
		Msg("Invalidated cached aggregations for endpoint")
}
