# T014: Endpoint Management

## Overview
Implement endpoint management functionality for adding, updating, and removing network monitoring endpoints through the application interface.

## Context
Users need to be able to manage their monitoring endpoints without manually editing JSON configuration files. This includes adding new endpoints, modifying existing ones, and organizing them by regions.

## Task Description
Create a comprehensive endpoint management system that allows full CRUD operations on monitoring endpoints with validation and real-time configuration updates.

## Acceptance Criteria
- [ ] Add new endpoints with validation
- [ ] Update existing endpoint configurations
- [ ] Remove endpoints from monitoring
- [ ] Move endpoints between regions
- [ ] Create and manage regions
- [ ] Endpoint configuration validation
- [ ] Real-time configuration persistence
- [ ] Frontend integration for endpoint management
- [ ] Undo/redo functionality for changes
- [ ] Import/export endpoint configurations

## API Methods to Implement
```go
func (a *App) AddEndpoint(regionName string, endpoint EndpointConfig) error
func (a *App) UpdateEndpoint(endpointID string, endpoint EndpointConfig) error
func (a *App) RemoveEndpoint(endpointID string) error
func (a *App) MoveEndpoint(endpointID, newRegionName string) error
func (a *App) CreateRegion(regionName string, thresholds RegionThresholds) error
func (a *App) UpdateRegion(regionName string, thresholds RegionThresholds) error
func (a *App) RemoveRegion(regionName string) error
func (a *App) ValidateEndpoint(endpoint EndpointConfig) (*ValidationResult, error)
```

## Endpoint Configuration Structure
```go
type EndpointConfig struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Type     string `json:"type"`     // HTTP, TCP, UDP, ICMP
    Address  string `json:"address"`
    Timeout  int    `json:"timeout"`  // milliseconds
    Enabled  bool   `json:"enabled"`
}

type ValidationResult struct {
    Valid    bool     `json:"valid"`
    Errors   []string `json:"errors"`
    Warnings []string `json:"warnings"`
}
```

## Validation Rules
- **Name**: Required, 1-100 characters, unique within region
- **Type**: Must be one of HTTP, TCP, UDP, ICMP
- **Address**: Valid URL for HTTP, host:port for TCP/UDP, IP for ICMP
- **Timeout**: 1-60 seconds
- **Duplicates**: Prevent duplicate addresses within same region

## Verification Steps
1. Add valid endpoint - should save to configuration
2. Add invalid endpoint - should return validation errors
3. Update endpoint address - should update configuration and restart monitoring
4. Remove endpoint - should stop monitoring and remove from config
5. Move endpoint between regions - should update region assignment
6. Create new region - should add to configuration
7. Validate duplicate endpoint - should prevent creation
8. Test configuration persistence - should survive application restart

## Dependencies
- T003: Configuration System
- T005: Wails Frontend-Backend Integration
- T011: Test Scheduler (for restarting monitoring)

## Notes
- Changes should trigger configuration file updates immediately
- Provide preview functionality before saving changes
- Consider backup/restore functionality for configurations
- Implement proper error handling for file system errors
- Support batch operations for efficiency