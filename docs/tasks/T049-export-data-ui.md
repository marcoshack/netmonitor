# T049: Export Data UI

## Overview
Implement frontend user interface for the data export functionality, allowing users to create, monitor, and download network monitoring data exports through an intuitive web interface.

## Context
The backend export system (T018) is complete and provides comprehensive APIs for exporting test result data in CSV and JSON formats. This task implements the frontend UI components that allow users to interact with the export system, configure export parameters, monitor export progress, and download completed exports.

## Task Description
Create a complete frontend interface for data export operations, including export configuration forms, job monitoring displays, progress tracking, history views, and file download capabilities.

## Acceptance Criteria
- [ ] Export dialog with date range picker and format selection
- [ ] Endpoint and region filter selection UI
- [ ] CSV column customization interface
- [ ] Compression and advanced options toggles
- [ ] Real-time export job progress display with progress bars
- [ ] Active exports list showing running jobs with status
- [ ] Export history view with completed/failed job details
- [ ] Download buttons for completed exports
- [ ] Cancel buttons for active export jobs
- [ ] Error display and handling for failed exports
- [ ] File size display for completed exports
- [ ] Responsive design for mobile/tablet views

## UI Components

### 1. Export Dialog/Modal
Location: `frontend/src/components/ExportDialog.js` (or Vue/React equivalent)

Features:
- **Date Range Selector**
  - Start date/time picker
  - End date/time picker
  - Quick preset buttons (Last 24h, Last 7 days, Last 30 days, Custom)
  - Visual calendar with date range highlighting

- **Format Selection**
  - Radio buttons for CSV / JSON
  - Format description tooltips
  - File extension preview

- **Filter Options**
  - Multi-select dropdown for endpoints
  - Multi-select dropdown for regions
  - "Select All" / "Clear All" buttons

- **CSV Options** (shown when CSV format selected)
  - Column selector with checkboxes
  - Default columns highlighted
  - Column reordering (drag-and-drop)

- **Advanced Options**
  - Compression toggle (ZIP)
  - Include raw data checkbox
  - Include aggregated data checkbox

- **Actions**
  - "Start Export" button (primary)
  - "Cancel" button (secondary)
  - Validation error messages

### 2. Export Progress View
Location: `frontend/src/components/ExportProgress.js`

Features:
- **Active Jobs List**
  - Job ID and creation timestamp
  - Export format and date range
  - Progress bar (0-100%)
  - Current phase text ("Reading data", "Writing file", etc.)
  - Estimated time remaining
  - Cancel button per job

- **Progress Card Layout**
  ```
  ┌─────────────────────────────────────────────┐
  │ CSV Export - Jan 1-7, 2025          [Cancel]│
  │ Job ID: abc123...                           │
  │ ▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░ 45%             │
  │ Reading data... (Est. 2min remaining)       │
  └─────────────────────────────────────────────┘
  ```

- **Real-time Updates**
  - Poll status every 2 seconds for active jobs
  - Auto-refresh progress bar
  - Update phase text dynamically

### 3. Export History View
Location: `frontend/src/components/ExportHistory.js`

Features:
- **History Table/List**
  - Job ID (short form with tooltip for full ID)
  - Format (CSV/JSON badge)
  - Date range
  - Status (Completed/Failed/Cancelled badge)
  - File size (for completed)
  - Creation time
  - Completion time
  - Actions column (Download/Delete)

- **Filters**
  - Status filter (All/Completed/Failed/Cancelled)
  - Date range filter for job creation
  - Format filter

- **Sorting**
  - Sort by creation time, completion time, file size
  - Ascending/descending toggle

- **Actions**
  - Download button (opens file or triggers download)
  - Delete button with confirmation
  - Bulk cleanup button ("Clean up old exports")

### 4. Export Button/Trigger
Location: Main dashboard toolbar or data view

Features:
- "Export Data" button with icon
- Opens export dialog on click
- Shows badge if exports are in progress
- Tooltip with quick info

## API Integration

### Backend API Methods (Already Available)
```javascript
// From app.go - all automatically exposed via Wails
await window.go.main.App.CreateExport(request)
await window.go.main.App.GetExportStatus(jobID)
await window.go.main.App.CancelExport(jobID)
await window.go.main.App.GetExportHistory()
await window.go.main.App.GetActiveExports()
await window.go.main.App.CleanupOldExports(retentionDays)
```

### Request/Response Types
```typescript
// Request
interface ExportRequest {
  format: "csv" | "json";
  startDate: string;  // ISO 8601
  endDate: string;    // ISO 8601
  endpoints: string[];
  regions: string[];
  columns: string[];  // For CSV
  compressed: boolean;
  includeRaw: boolean;
  includeAgg: boolean;
}

// Response
interface ExportJob {
  id: string;
  request: ExportRequest;
  status: "pending" | "running" | "completed" | "failed" | "cancelled";
  progress: number;  // 0.0 to 1.0
  startTime: string;
  endTime?: string;
  filePath?: string;
  fileSize: number;
  error?: string;
}

interface ExportStatus {
  job: ExportJob;
  recordsProcessed: number;
  totalRecords: number;
  currentPhase: string;
  estimatedTimeLeft: string;
}
```

## State Management

### Export State
```javascript
const exportState = {
  // Dialog state
  dialogOpen: false,
  dialogConfig: {
    format: "csv",
    startDate: null,
    endDate: null,
    endpoints: [],
    regions: [],
    columns: ["timestamp", "endpoint_id", "protocol", "status", "latency_ms"],
    compressed: false,
    includeRaw: true,
    includeAgg: false
  },

  // Active jobs
  activeJobs: [],

  // History
  history: [],
  historyFilter: "all",

  // UI state
  loading: false,
  error: null,
  lastRefresh: null
};
```

### Actions
- `openExportDialog()` - Open dialog with default settings
- `closeExportDialog()` - Close and reset dialog
- `updateExportConfig(field, value)` - Update configuration
- `startExport()` - Submit export request
- `cancelExport(jobID)` - Cancel running export
- `refreshActiveJobs()` - Poll for updates
- `refreshHistory()` - Reload history
- `downloadExport(jobID)` - Download completed export
- `deleteExport(jobID)` - Delete export file
- `cleanupOldExports(days)` - Bulk cleanup

## User Flows

### Flow 1: Create Export
1. User clicks "Export Data" button
2. Export dialog opens with sensible defaults
3. User selects date range (or uses preset)
4. User optionally filters endpoints/regions
5. User selects format (CSV/JSON)
6. If CSV, user customizes columns
7. User toggles compression if needed
8. User clicks "Start Export"
9. Dialog shows "Creating export..." spinner
10. Dialog closes, success notification appears
11. Export appears in active jobs list
12. Progress updates in real-time

### Flow 2: Monitor Active Export
1. User navigates to exports section
2. Active exports show with progress bars
3. Progress updates every 2 seconds
4. User sees current phase and estimated time
5. When complete, job moves to history
6. Notification: "Export complete - Ready to download"

### Flow 3: Download Export
1. User views export history
2. User finds completed export
3. User clicks "Download" button
4. Browser downloads the file
5. Optional: Auto-open file location

### Flow 4: Cancel Export
1. User sees long-running export
2. User clicks "Cancel" button
3. Confirmation dialog: "Cancel this export?"
4. User confirms
5. Export status changes to "Cancelled"
6. Job moves to history

## Implementation Guidelines

### Technology Stack
- Use project's existing frontend framework (Wails + vanilla JS/Vue/React)
- Integrate with existing state management
- Follow established UI component patterns
- Use existing design system/CSS framework

### Performance Considerations
- Debounce status polling (2-5 second intervals)
- Limit history to last 100 exports (with pagination if needed)
- Cancel polling when exports complete
- Lazy load history on tab/page open

### Error Handling
- Display validation errors inline in dialog
- Show error notifications for API failures
- Provide retry option for failed exports
- Clear error states on retry

### Accessibility
- Proper ARIA labels for all controls
- Keyboard navigation support
- Focus management in modal dialogs
- Screen reader announcements for progress updates

### Responsive Design
- Mobile: Stack form fields vertically, full-width buttons
- Tablet: 2-column layout where appropriate
- Desktop: Optimal spacing and multi-column layouts
- Ensure touch-friendly button sizes (min 44x44px)

## Verification Steps
1. Open export dialog and verify all controls work
2. Create CSV export with custom columns - should complete successfully
3. Create JSON export with compression - should create .zip file
4. Monitor active export - should show real-time progress
5. Cancel running export - should stop and mark as cancelled
6. View export history - should show all past exports
7. Download completed export - should download correct file
8. Test date range presets - should set correct date ranges
9. Test endpoint filtering - should only export selected endpoints
10. Test on mobile device - should be fully functional and touch-friendly

## Dependencies
- T018: Data Export Functionality (Backend) - **COMPLETED**
- T035: Frontend API Integration - Required for state management framework
- T026: Dashboard Layout - For embedding export UI
- T032: Time Range Selector Component - Can reuse date picker components

## Notes
- Backend APIs are already complete and tested
- All export business logic is handled by backend
- Frontend is purely presentational and orchestration
- Consider using Wails' file dialog API for download location selection
- Plan for future features: scheduled exports, export templates
- Export files are stored in `./exports` directory on backend
- Maximum file download size considerations for browser limits
- Consider adding export preview (first 10 rows) before download

## Example UI Mockup

```
┌─────────────────────────────────────────────────────────┐
│ Export Data                                      [Close] │
├─────────────────────────────────────────────────────────┤
│ Date Range                                              │
│ ┌─────────────┐  to  ┌─────────────┐                   │
│ │ 2025-01-01  │      │ 2025-01-07  │                   │
│ └─────────────┘      └─────────────┘                   │
│ [Last 24h] [Last 7d] [Last 30d] [Custom]               │
│                                                         │
│ Format                                                  │
│ ○ CSV  ● JSON                                          │
│                                                         │
│ Filters (Optional)                                      │
│ Endpoints: [Select endpoints...        ▼]              │
│ Regions:   [Select regions...          ▼]              │
│                                                         │
│ Options                                                 │
│ ☑ Compress (ZIP)                                       │
│ ☑ Include raw test results                             │
│ ☐ Include aggregated data                              │
│                                                         │
│                              [Cancel] [Start Export]    │
└─────────────────────────────────────────────────────────┘
```

## Future Enhancements (Post-MVP)
- Drag-and-drop date range selection on graphs
- Export templates (save common export configurations)
- Scheduled/recurring exports
- Email notification when export completes
- Export directly to cloud storage (S3, Drive, Dropbox)
- Bulk export operations (export multiple date ranges)
- Export data preview before download
- Chart/visualization export (export as image)
