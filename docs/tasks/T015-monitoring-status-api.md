# T015: Monitoring Status API

## Overview
Implement comprehensive status monitoring API that provides real-time information about monitoring state, test results, and system health.

## Context
The frontend dashboard needs access to real-time monitoring status, recent test results, and system health information to display current network status and update the user interface.

## Task Description
Create a robust status API that exposes monitoring state, recent results, system health, and real-time updates for the frontend dashboard.

## Acceptance Criteria
- [ ] Real-time monitoring status for all endpoints
- [ ] Recent test results retrieval (last N tests)
- [ ] System health and performance metrics
- [ ] Live status updates (WebSocket or polling)
- [ ] Historical data queries for dashboard graphs
- [ ] Regional status summaries
- [ ] Alert and threshold status reporting
- [ ] Performance metrics (memory, CPU usage)
- [ ] WebSocket integration for real-time updates

## API Methods to Implement
```go
// Status and monitoring state
func (a *App) GetMonitoringStatus() (*MonitoringStatus, error)
func (a *App) GetEndpointStatus(endpointID string) (*EndpointStatus, error)
func (a *App) GetRegionStatus(regionName string) (*RegionStatus, error)

// Historical data for graphs
func (a *App) GetRecentResults(endpointID string, hours int) ([]*TestResult, error)
func (a *App) GetAggregatedData(endpointID string, period string, hours int) ([]*AggregatedResult, error)

// System health
func (a *App) GetSystemHealth() (*SystemHealth, error)
func (a *App) GetPerformanceMetrics() (*PerformanceMetrics, error)
```

## Status Data Structures
```go
type MonitoringStatus struct {
    Running         bool               `json:"running"`
    StartTime       time.Time          `json:"startTime"`
    TotalEndpoints  int                `json:"totalEndpoints"`
    ActiveEndpoints int                `json:"activeEndpoints"`
    LastTestTime    time.Time          `json:"lastTestTime"`
    NextTestTime    time.Time          `json:"nextTestTime"`
    RegionStatus    map[string]*RegionStatus `json:"regionStatus"`
}

type EndpointStatus struct {
    ID           string        `json:"id"`
    Name         string        `json:"name"`
    Status       string        `json:"status"`      // "up", "down", "warning"
    LastLatency  time.Duration `json:"lastLatency"`
    LastTest     time.Time     `json:"lastTest"`
    Uptime       float64       `json:"uptime"`      // Percentage
    ConsecutiveFails int       `json:"consecutiveFails"`
}

type RegionStatus struct {
    Name            string            `json:"name"`
    EndpointCount   int               `json:"endpointCount"`
    HealthyCount    int               `json:"healthyCount"`
    WarningCount    int               `json:"warningCount"`
    DownCount       int               `json:"downCount"`
    AverageLatency  float64           `json:"averageLatency"`
    OverallHealth   string            `json:"overallHealth"`
}
```

## Real-time Updates
- WebSocket connection for live status updates
- Event-driven updates when test results arrive
- Efficient delta updates to minimize bandwidth
- Connection management and reconnection logic

## Verification Steps
1. Get monitoring status - should return current state
2. Get endpoint status - should return individual endpoint health
3. Get region status - should aggregate endpoint statuses
4. Query recent results - should return last N test results
5. Test real-time updates - should push updates via WebSocket
6. Verify system health metrics - should report accurate resource usage
7. Test with high test frequency - should handle rapid updates
8. Verify historical data queries - should return correct time ranges

## Dependencies
- T011: Test Scheduler
- T013: Test Result Aggregation
- T005: Wails Frontend-Backend Integration
- T006: Network Test Interfaces

## Notes
- Implement efficient caching for frequently requested data
- Consider rate limiting for status queries
- Use appropriate data structures for fast lookups
- Plan for future dashboard widgets and requirements
- Implement proper error handling for missing data