# T005: Wails Frontend-Backend Integration

## Overview
Establish the communication bridge between the Go backend and web frontend using Wails context and methods.

## Context
Wails v2 provides a context system for frontend-backend communication. The frontend needs to be able to call Go methods and receive data from the backend.

## Task Description
Set up Wails context binding, create basic API methods for frontend-backend communication, and establish the foundation for data exchange.

## Acceptance Criteria
- [X] Wails context properly configured in Go backend
- [X] Basic API methods exposed to frontend:
  - `GetConfiguration()` - Retrieve current config
  - `GetSystemInfo()` - Get basic system information
  - `SetTheme(theme string)` - Set application theme
- [X] Frontend can successfully call backend methods
- [X] Error handling for frontend-backend communication
- [X] Frontend displays data received from backend

## API Methods to Implement
```go
type App struct {
    ctx context.Context
}

func (a *App) GetConfiguration() (*config.Config, error)
func (a *App) GetSystemInfo() (*SystemInfo, error)
func (a *App) SetTheme(theme string) error
```

## Verification Steps
1. Call `GetConfiguration()` from frontend - should return config data
2. Call `GetSystemInfo()` from frontend - should return system information
3. Call `SetTheme()` with "dark" - should update theme
4. Test with invalid parameters - should return appropriate errors
5. Verify all calls work without blocking the UI

## Dependencies
- T002: Basic Application Structure
- T003: Configuration System
- T004: Basic Frontend Setup

## Notes
- Use Wails context for method binding
- Implement proper error handling and return types
- Consider using JSON for complex data structures
- Prepare for more complex API methods in later tasks