# NetMonitor

## Project Overview

NetMonitor is a cross-platform desktop application built with Wails v2 that monitors network connectivity and performance over time. It uses Go for the backend (network testing, data storage, configuration management) and a web-based frontend (HTML/CSS/JavaScript with Vite).

## Development Commands

### Building and Running

- **Development mode**: `wails dev` - Runs with hot reload for frontend changes
  - Enable console logging on Windows: `$env:LOG_CONSOLE=1; wails dev`
- **Production build**: `wails build` - Creates redistributable package
- **Frontend only**: `cd frontend && npm run dev` - Run Vite dev server
- **Frontend build**: `cd frontend && npm run build` - Build frontend for production

### Testing

- **Run all tests**: `go test ./...`
- **Run specific package tests**: `go test ./internal/config` (or any other package)
- **Run with verbose output**: `go test -v ./...`
- **Run specific test**: `go test -run TestFunctionName ./internal/package`

## Architecture

### Application Structure

The application follows a manager-based architecture with three core components initialized at startup:

1. **Config Manager** ([internal/config/manager.go](internal/config/manager.go)): Handles configuration loading, validation, and file watching. Automatically reloads `config.json` when it changes on disk. Supports callbacks for configuration change notifications.

2. **Storage Manager** ([internal/storage/manager.go](internal/storage/manager.go)): Manages test result persistence using daily JSON files (`data/YYYY-MM-DD.json`). Each file contains all test results for that day plus metadata (version, timestamps, result count).

3. **Monitor Manager** ([internal/monitor/manager.go](internal/monitor/manager.go)): Orchestrates scheduled and manual network tests. Runs a monitoring loop based on the configured test interval. Currently uses mock implementations for actual network testing.

### Data Flow

- Application starts → Initialize logger → Create App → Wails calls `App.startup()` → Initialize managers in order (config, storage, monitor)
- Configuration changes → File watcher triggers reload → Callbacks notify dependent components
- Scheduled tests → Monitoring loop ticks → Execute tests for all endpoints → Store results via Storage Manager
- Manual tests → API call from frontend → Execute single test → Store result → Return result to frontend

### Configuration System

Configuration is managed through [internal/config/manager.go](internal/config/manager.go):
- Structured as regions containing endpoints with thresholds
- Each endpoint has: name, type (HTTP/TCP/UDP/ICMP), address, timeout
- Settings include: test interval, data retention days, notifications flag
- Validation enforces: test intervals (1-1440 min), timeouts (1000-60000 ms), retention (1-365 days)
- File watcher automatically reloads on external changes

### Network Testing Interfaces

The [internal/network/interfaces.go](internal/network/interfaces.go) file defines the contract for network testing implementations:
- `NetworkTest` interface: Execute(), GetProtocol(), Validate()
- `TestResult` struct: Common result format for all test types
- Protocol-specific configs: `HTTPConfig`, `TCPConfig`, `UDPConfig`, `ICMPConfig`
- **Note**: Actual protocol implementations (ICMP ping, HTTP, TCP, UDP) are not yet implemented

### Frontend-Backend Communication

The frontend communicates with Go backend through Wails bindings:
- Go methods on `App` struct are automatically exposed to JavaScript
- Available APIs: `GetSystemInfo()`, `GetConfiguration()`, `SetTheme()`, `GetMonitoringStatus()`, `StartMonitoring()`, `StopMonitoring()`, `RunManualTest(endpointID)`
- Frontend is embedded at compile time via `//go:embed all:frontend/dist`

### Logging

The application uses zerolog for structured logging:
- Initialized in [internal/logging/logger.go](internal/logging/logger.go)
- Context-aware logging throughout: `log.Ctx(ctx).Info().Msg("message")`
- Any non-trivial methods should receive a `ctx` (context.Context type) as the first attribute
- Console output enabled via `LOG_CONSOLE=1` environment variable

## Key Implementation Notes

### Adding New Network Test Types

1. Implement the `NetworkTest` interface from [internal/network/interfaces.go](internal/network/interfaces.go)
2. Add protocol-specific configuration struct if needed
3. Update validation in config manager to support the new type
4. Wire the implementation into monitor manager's test execution

### Working with Configuration

- Configuration file: `config.json` in project root
- Default configuration is auto-generated if file doesn't exist
- Always use `config.Manager.GetConfig()` to read (returns deep copy)
- Use `config.Manager.UpdateConfig()` to modify (validates before saving)
- File changes are detected automatically by fsnotify watcher

### Storage Patterns

- All test results go through `storage.Manager.StoreTestResult()`
- Results are automatically grouped by date into daily files
- File structure: date, results array, metadata (version, timestamps, count)
- Storage directory: `./data` relative to working directory
- Thread-safe with RWMutex protection

## Testing Standards

The project aims for 80% unit test coverage (excluding UI). When writing tests:
- Place test files alongside implementation: `manager_test.go` next to `manager.go`
- Use table-driven tests for multiple scenarios
- Example: [internal/network/interfaces_test.go](internal/network/interfaces_test.go)
- Mock external dependencies (filesystem, network, time)

## Coding Standards

- Follow Go conventions (gofmt, golint)
- Use context.Context for all long-running and/or non-trivial operations
- Use structured logging with zerolog
- Write clear, maintainable code with comments for complex logic
- Use dependency injection where appropriate for testability
- Use input/output objects to facilitate changes without changing methods signature
- Validate all external inputs (config files, user input)
- Imports must be separated by new line:
  - first all standard packages
  - then external dependencies
  - finally imports from the current project
- All public interfaces, structs, methods and attributes must have a comment explaining it's intent and usage examples, when applicable

## Configuration File Format

See [docs/specs/netmonitor.md](docs/specs/netmonitor.md) for detailed configuration examples. Key structure:
- `regions`: Map of region name to Region (endpoints + thresholds)
- `settings`: Global settings (test interval, retention, notifications)
- Endpoint types: "HTTP", "TCP", "UDP", "ICMP"
- Timeouts in milliseconds, intervals in minutes, retention in days

## Project Status

Active development tracked in [docs/tasks/](docs/tasks/). Completed tasks: T001-T006 (project setup, app structure, config system, frontend setup, Wails integration, network test interfaces).