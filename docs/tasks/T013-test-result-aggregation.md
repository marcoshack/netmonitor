# T013: Test Result Aggregation

## Overview
Implement result aggregation system that computes hourly and daily statistics from individual test results for efficient data visualization and reporting.

## Context
NetMonitor needs to efficiently display trends and statistics over time. Raw test results need to be aggregated into hourly and daily summaries to support the dashboard graphs and historical analysis.

## Task Description
Create an aggregation system that processes individual test results and generates statistical summaries for different time periods.

## Acceptance Criteria
- [ ] Hourly aggregation of test results by endpoint
- [ ] Daily aggregation of test results by endpoint
- [ ] Statistical calculations:
  - Average latency
  - Minimum/Maximum latency
  - Success rate (availability percentage)
  - Test count
  - Standard deviation
- [ ] Real-time aggregation as new results arrive
- [ ] Aggregation data storage alongside raw results
- [ ] Efficient queries for dashboard data
- [ ] Background aggregation processing
- [ ] Unit tests for aggregation calculations

## Aggregation Data Structure
```go
type AggregatedResult struct {
    StartTime       time.Time
    EndTime         time.Time
    Period          string    // "hourly" or "daily"
    EndpointID      string
    RegionName      string
    TestCount       int
    SuccessCount    int
    FailureCount    int
    AvgLatency      float64
    MinLatency      time.Duration
    MaxLatency      time.Duration
    StdDevLatency   float64
    AvailabilityPct float64
}
```

## Implementation Requirements
- Process test results in batches for efficiency
- Handle missing data points appropriately
- Support partial aggregations for current time periods
- Implement efficient storage and retrieval
- Handle timezone considerations

## Aggregation Scenarios
- **Hourly**: Aggregate all tests within each hour
- **Daily**: Aggregate all tests within each day
- **Region-wide**: Aggregate across all endpoints in a region
- **Real-time**: Update current period aggregations as tests complete

## Verification Steps
1. Generate test results for 24 hours - should create 24 hourly aggregations
2. Verify statistical calculations accuracy - averages, min/max should be correct
3. Test partial hour aggregation - should handle incomplete periods
4. Test with missing data - should handle gaps appropriately
5. Verify real-time updates - should update current aggregations
6. Test region-wide aggregations - should combine multiple endpoints
7. Performance test with large datasets - should aggregate efficiently

## Dependencies
- T006: Network Test Interfaces
- T011: Test Scheduler (for generating results to aggregate)

## Notes
- Consider using sliding window aggregations for real-time updates
- Implement efficient storage queries for dashboard needs
- Handle leap seconds and daylight saving time changes
- Consider pre-computing common dashboard queries
- Plan for future aggregation periods (weekly, monthly)