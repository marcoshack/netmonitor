# NetMonitor

NetMonitor is a desktop application that monitors network connectivity and performance over time, providing detailed reports about Internet Service Provider and local network quality.

## Technical Architecture

- **Backend**: Go application handling network monitoring, data storage, and system tray integration
- **Frontend**: Web-based UI using HTML/CSS/JavaScript
- **Framework**: Wails v2 for cross-platform desktop app with web frontend
- **Data Storage**: JSON files for configuration and historical data storage
- **Supported Platforms**: Windows, macOS, Linux

## Core Features

### Network Monitoring
- **Protocols Supported**: HTTP, TCP, UDP, and ICMP endpoints
- **Metrics Collected**: 
  - Latency (response time in milliseconds)
  - Availability (success/failure status)
- **Regional Organization**: Endpoints grouped by geographical regions (e.g., NA-East, EU-West, Asia-Pacific)
- **Test Intervals**: Configurable from 1 minute to 24 hours (default: 5 minutes)

### User Interface
- **System Tray Integration**: Background operation with system tray icon and context menu
- **Main Dashboard**: 
  - Interactive graphs showing latency trends over time
  - Region selector dropdown for viewing different geographical areas
  - Time range selector (last 24 hours, week, month)
  - Real-time status indicators for each monitored endpoint
- **Manual Testing**: On-demand test execution button with immediate detailed results
- **Theme Support**: Light and dark mode with automatic system theme detection

### Data Management
- **Configuration Format**: JSON-based endpoint configuration
- **Data Storage**: Historical data stored in daily JSON files organized by date
- **Data Retention**: Configurable retention period (default: 90 days)
- **Export Capabilities**: CSV and JSON export for custom date ranges

### Notifications and Alerts
- **Threshold Monitoring**: Configurable latency and availability thresholds per region
- **Notification Types**: 
  - System notifications for failures and threshold breaches
  - Visual indicators in system tray icon
  - Optional email notifications (future enhancement)
- **Alert Conditions**:
  - Endpoint failure (connection timeout or error)
  - Latency above configured threshold
  - Availability below configured percentage

## Configuration Structure

### Endpoint Configuration Example
```json
{
  "regions": {
    "NA-East": {
      "endpoints": [
        {
          "name": "Google DNS",
          "type": "ICMP",
          "address": "8.8.8.8",
          "timeout": 5000
        },
        {
          "name": "Cloudflare HTTP",
          "type": "HTTP",
          "address": "https://1.1.1.1",
          "timeout": 10000
        }
      ],
      "thresholds": {
        "latency_ms": 100,
        "availability_percent": 99.0
      }
    }
  },
  "settings": {
    "test_interval_minutes": 5,
    "data_retention_days": 90,
    "notifications_enabled": true
  }
}
```

## Settings and Preferences

The settings screen provides configuration for:

### General Settings
- **Configuration File Location**: Custom path for configuration file
- **Test Interval**: Frequency of automated tests (1 min - 24 hours)
- **Data Retention**: How long to keep historical data (7-365 days)
- **Auto-start**: Launch application on system startup

### Notification Settings  
- **Enable/Disable Notifications**: Toggle system notifications
- **Latency Thresholds**: Per-region latency warning levels
- **Availability Thresholds**: Minimum acceptable uptime percentages
- **Notification Cooldown**: Minimum time between repeated alerts

### Appearance
- **Theme Selection**: Light, Dark, or System default
- **Graph Colors**: Customizable color scheme for data visualization
- **Language**: Interface localization (initially English only)

### Advanced Settings
- **Network Timeouts**: Connection and response timeout values
- **Concurrent Tests**: Maximum number of simultaneous endpoint tests
- **Logging Level**: Application log verbosity (Error, Warn, Info, Debug)

## Data Storage Format

### Historical Data Structure
- **File Organization**: `data/YYYY-MM-DD.json` for daily data files
- **Data Points**: Each test result includes timestamp, endpoint ID, latency, and status
- **Aggregation**: Hourly and daily statistics computed for efficient visualization
- **Retention**: Keep the last 3 months or historical data

### Example Data Entry
```json
{
  "timestamp": "2025-09-27T19:30:00Z",
  "endpoint_id": "na-east-google-dns",
  "latency_ms": 23,
  "status": "success"
}
```

## Implementation Considerations

### Performance Requirements
- **Memory Usage**: < 50MB typical, < 100MB maximum
- **CPU Usage**: < 1% during idle monitoring, < 5% during active testing
- **Disk I/O**: Minimal, append-only writes for data logging

### Security Considerations
- **Network Access**: Outbound connections only to configured endpoints
- **Data Privacy**: All data stored locally, no external telemetry
- **Configuration Security**: Validate all user-provided endpoints and settings

### Error Handling
- **Network Failures**: Graceful handling of connection timeouts and errors
- **Data Corruption**: Backup and recovery mechanisms for configuration and data
- **System Integration**: Robust system tray and notification system integration

### Coding Standards
- At least 80% unit test coverage (excluding UI)
- Follow Go and JavaScript best practices and style guides


## Future Enhancements

- Email notification support
- Custom alerting rules and scripting
- Network topology discovery
- Bandwidth utilization monitoring
- Multi-language internationalization
- Cloud backup and synchronization