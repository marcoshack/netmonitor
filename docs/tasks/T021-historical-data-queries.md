# T021: Historical Data Queries

## Overview
Implement efficient historical data query system that supports dashboard requirements for time series graphs, trend analysis, and statistical reporting.

## Context
The dashboard needs to display interactive graphs showing latency trends over various time periods (24 hours, week, month). The query system must efficiently retrieve and aggregate historical data for visualization.

## Task Description
Create a comprehensive query system that can efficiently retrieve historical data for different time ranges, endpoints, and aggregation levels with support for dashboard requirements.

## Acceptance Criteria
- [ ] Time range queries (last 24 hours, week, month, custom)
- [ ] Endpoint-specific data queries
- [ ] Region-wide data aggregation queries
- [ ] Statistical queries (min, max, average, percentiles)
- [ ] Data point sampling for large time ranges
- [ ] Query result caching for performance
- [ ] Pagination support for large result sets
- [ ] Query optimization for common dashboard patterns
- [ ] Real-time data integration with historical queries

## Query Types
```go
type QueryRequest struct {
    Type        string    `json:"type"`        // "timeseries", "statistics", "availability"
    StartTime   time.Time `json:"startTime"`
    EndTime     time.Time `json:"endTime"`
    Endpoints   []string  `json:"endpoints"`   // Specific endpoints or empty for all
    Regions     []string  `json:"regions"`     // Specific regions or empty for all
    Granularity string    `json:"granularity"` // "raw", "hourly", "daily"
    MaxPoints   int       `json:"maxPoints"`   // Limit result size for graphs
    Metrics     []string  `json:"metrics"`     // "latency", "availability", "success_rate"
}
```

## API Methods
```go
func (a *App) QueryTimeSeries(request QueryRequest) (*TimeSeriesResult, error)
func (a *App) QueryStatistics(request QueryRequest) (*StatisticsResult, error)
func (a *App) QueryAvailability(request QueryRequest) (*AvailabilityResult, error)
func (a *App) QueryTrends(request QueryRequest) (*TrendResult, error)
func (a *App) QueryRecentData(endpointID string, minutes int) ([]*TestResult, error)
```

## Result Structures
```go
type TimeSeriesResult struct {
    StartTime   time.Time           `json:"startTime"`
    EndTime     time.Time           `json:"endTime"`
    Granularity string              `json:"granularity"`
    Series      map[string][]DataPoint `json:"series"` // endpoint_id -> data points
    Metadata    QueryMetadata       `json:"metadata"`
}

type DataPoint struct {
    Timestamp time.Time `json:"timestamp"`
    Value     float64   `json:"value"`
    Status    string    `json:"status"`
    Count     int       `json:"count"`    // Number of samples in this point
}

type StatisticsResult struct {
    StartTime   time.Time                    `json:"startTime"`
    EndTime     time.Time                    `json:"endTime"`
    Statistics  map[string]EndpointStats     `json:"statistics"` // endpoint_id -> stats
}

type EndpointStats struct {
    TestCount       int           `json:"testCount"`
    SuccessCount    int           `json:"successCount"`
    FailureCount    int           `json:"failureCount"`
    AvgLatency      float64       `json:"avgLatency"`
    MinLatency      time.Duration `json:"minLatency"`
    MaxLatency      time.Duration `json:"maxLatency"`
    P50Latency      time.Duration `json:"p50Latency"`
    P95Latency      time.Duration `json:"p95Latency"`
    P99Latency      time.Duration `json:"p99Latency"`
    AvailabilityPct float64       `json:"availabilityPct"`
}
```

## Query Optimizations
- **Data Sampling**: Reduce data points for large time ranges
- **Aggregation**: Use pre-computed hourly/daily aggregations
- **Caching**: Cache frequently requested queries
- **Indexing**: Fast lookup by timestamp and endpoint
- **Streaming**: Handle large result sets efficiently

## Dashboard Query Patterns
- **Live Updates**: Last 30 minutes with real-time updates
- **Daily View**: Last 24 hours with hourly granularity
- **Weekly View**: Last 7 days with daily granularity
- **Monthly View**: Last 30 days with daily granularity
- **Comparison**: Compare multiple endpoints or time periods

## Verification Steps
1. Query last 24 hours - should return hourly aggregated data
2. Query specific endpoint - should filter results correctly
3. Query large time range - should sample data points appropriately
4. Query with multiple endpoints - should return data for all requested endpoints
5. Test query caching - should return cached results for repeat queries
6. Test statistical queries - should calculate correct statistics
7. Verify pagination - should handle large result sets
8. Test real-time integration - should include latest test results

## Dependencies
- T016: JSON Storage System
- T013: Test Result Aggregation
- T020: Storage Performance Optimization

## Notes
- Optimize for common dashboard query patterns
- Consider using background query pre-computation
- Implement proper error handling for missing data
- Support timezone handling for queries
- Plan for future query complexity (filters, sorting)
- Consider implementing GraphQL-style flexible queries