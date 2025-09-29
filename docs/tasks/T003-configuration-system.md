# T003: Configuration System

## Overview
Implement the JSON-based configuration system for endpoint definitions, regional organization, and application settings.

## Context
NetMonitor uses JSON configuration files to define:
- Network endpoints organized by geographical regions
- Monitoring thresholds per region
- Application settings (test intervals, data retention, notifications)

## Task Description
Create a robust configuration system that loads, validates, and manages the JSON configuration structure.

## Acceptance Criteria
- [X] Configuration struct definitions matching the spec format
- [X] JSON configuration file loading and parsing
- [X] Configuration validation (required fields, value ranges)
- [X] Default configuration generation
- [X] Configuration file watching for live updates
- [X] Error handling for malformed configurations

## Expected Configuration Structure
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

## Verification Steps
1. Load valid configuration file - should parse successfully
2. Load malformed JSON - should return appropriate error
3. Validate configuration with missing required fields - should fail validation
4. Generate default config - should create valid configuration
5. Test configuration file watching - should detect changes

## Dependencies
- T002: Basic Application Structure

## Notes
- Use Go's `encoding/json` package
- Implement proper validation for all fields
- Support both absolute and relative file paths
- Consider using `fsnotify` for file watching