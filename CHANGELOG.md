# Changelog

All notable changes to this project will be documented in this file.

## [v0.3] - 2025-12-14

### Features
- **Start on Boot**: Added option to automatically start the application on system boot (Windows, Linux, macOS).
- **Data Visualization**: Improved graph to show data point gaps and distinct failure indicators (red dots).
- **Details View**: Added period selector to endpoint details modal for granular history views.

### Improvements
- **UX**: Pressing `ESC` now closes open modals.
- **UX**: Endpoint details modal now always opens at the top.
- **UX**: Improved endpoint delete confirmation with clearer prompt.
- **Documentation**: Cleaned up README.

### Internals
- Refactored startup logic to cross-platform `internal/startup` package.

## [v0.2] - 2025-12-13

### Features
- **Monitor Management**: Added ability to Add, Edit, and Delete monitors.
- **Drag & Drop**: Added drag-and-drop support to reorder monitor cards.
- **System Integ**: Added System Tray support for background operation.
- **Windows Installer**: Setup logic for creating a Windows installer.

### Improvements
- **Storage**: Migrated application data to `%AppData%`.
- **Concurrency**: Enforced single instance lock using Wails.
- **Performance**: Native Go ping implementation replacing system ping.
- **Logs**: Added application logs.
- **Data**: Changed timestamp format to UnixMilli.

### Initial
- Initial release with basic monitoring capabilities.
