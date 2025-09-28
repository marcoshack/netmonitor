# T001: Project Setup and Initialization

## Overview
Set up the basic NetMonitor project structure with Wails v2 framework for a cross-platform desktop application with web frontend.

## Context
NetMonitor is a desktop application that monitors network connectivity and performance over time. It uses:
- **Backend**: Go application handling network monitoring, data storage, and system tray integration
- **Frontend**: Web-based UI using HTML/CSS/JavaScript
- **Framework**: Wails v2 for cross-platform desktop app with web frontend
- **Supported Platforms**: Windows, macOS, Linux

## Task Description
Initialize a new Wails v2 project with the proper directory structure and basic configuration.

## Acceptance Criteria
- [X] Wails v2 project successfully created
- [X] Project builds without errors (`wails build`)
- [X] Application launches and displays a basic window
- [X] Directory structure follows Wails conventions:
  - `app/` - Go backend code
  - `frontend/` - Web frontend code
  - `build/` - Build output
  - `wails.json` - Wails configuration

## Verification Steps
1. Run `wails build` - should complete without errors
2. Run the generated executable - should launch a window
3. Verify all required directories are present
4. Check that `wails.json` contains proper project configuration

## Dependencies
- None (this is the first task)

## Notes
- Use Wails v2 latest stable version
- Configure for cross-platform builds (Windows, macOS, Linux)
- Set application name to "NetMonitor"