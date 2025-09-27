# T026: Dashboard Layout and Structure

## Overview
Create the main dashboard layout with responsive design, navigation structure, and component organization for the NetMonitor web interface.

## Context
NetMonitor's main interface is a web-based dashboard that displays network monitoring data, graphs, and controls. The layout needs to be intuitive, responsive, and efficiently organize complex monitoring information.

## Task Description
Design and implement the core dashboard layout with navigation, content areas, responsive behavior, and the foundational structure for all dashboard components.

## Acceptance Criteria
- [ ] Responsive grid-based layout system
- [ ] Main navigation with clear sections
- [ ] Content areas for graphs, status, and controls
- [ ] Mobile-friendly responsive behavior
- [ ] Consistent spacing and typography
- [ ] Theme-aware styling system
- [ ] Loading states and skeleton screens
- [ ] Accessibility compliance (ARIA, keyboard navigation)
- [ ] Cross-browser compatibility

## Layout Structure
```
┌─────────────────────────────────────────────────────┐
│ Header: Title, Theme Toggle, Settings              │
├─────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────────────────────────┐ │
│ │   Sidebar   │ │        Main Content Area        │ │
│ │             │ │                                 │ │
│ │ - Overview  │ │ ┌─────────────────────────────┐ │ │
│ │ - Regions   │ │ │      Status Overview        │ │ │
│ │ - Endpoints │ │ │                             │ │ │
│ │ - Settings  │ │ └─────────────────────────────┘ │ │
│ │ - Manual    │ │ ┌─────────────────────────────┐ │ │
│ │   Tests     │ │ │      Latency Graphs         │ │ │
│ │             │ │ │                             │ │ │
│ └─────────────┘ │ └─────────────────────────────┘ │ │
│                 │ ┌─────────────────────────────┐ │ │
│                 │ │    Endpoint Status Grid     │ │ │
│                 │ │                             │ │ │
│                 │ └─────────────────────────────┘ │ │
│                 └─────────────────────────────────┘ │
├─────────────────────────────────────────────────────┤
│ Footer: Status Bar, Connection Info                │
└─────────────────────────────────────────────────────┘
```

## CSS Architecture
- **CSS Variables**: Theme-aware color system
- **CSS Grid**: Main layout structure
- **Flexbox**: Component-level layouts
- **Media Queries**: Responsive breakpoints
- **CSS Modules**: Component-scoped styling

## Component Structure
```html
<div class="app-container">
  <header class="app-header">
    <div class="header-content">
      <h1 class="app-title">NetMonitor</h1>
      <div class="header-controls">
        <button class="theme-toggle" aria-label="Toggle theme">🌙</button>
        <button class="settings-btn" aria-label="Settings">⚙️</button>
      </div>
    </div>
  </header>

  <div class="app-body">
    <nav class="sidebar">
      <ul class="nav-menu">
        <li><a href="#overview" class="nav-link active">Overview</a></li>
        <li><a href="#regions" class="nav-link">Regions</a></li>
        <li><a href="#endpoints" class="nav-link">Endpoints</a></li>
        <li><a href="#manual" class="nav-link">Manual Tests</a></li>
        <li><a href="#settings" class="nav-link">Settings</a></li>
      </ul>
    </nav>

    <main class="main-content">
      <div class="content-grid">
        <section class="status-overview">
          <!-- Status widgets -->
        </section>
        <section class="graphs-section">
          <!-- Charts and graphs -->
        </section>
        <section class="endpoints-section">
          <!-- Endpoint status grid -->
        </section>
      </div>
    </main>
  </div>

  <footer class="app-footer">
    <div class="status-bar">
      <span class="connection-status">Connected</span>
      <span class="last-update">Last update: 2 minutes ago</span>
    </div>
  </footer>
</div>
```

## Responsive Breakpoints
- **Mobile**: < 768px (stacked layout, collapsible sidebar)
- **Tablet**: 768px - 1024px (reduced columns)
- **Desktop**: > 1024px (full layout)

## Theme System
```css
:root {
  /* Light theme */
  --color-primary: #007bff;
  --color-secondary: #6c757d;
  --color-success: #28a745;
  --color-warning: #ffc107;
  --color-danger: #dc3545;
  --color-background: #ffffff;
  --color-surface: #f8f9fa;
  --color-text-primary: #212529;
  --color-text-secondary: #6c757d;
  --color-border: #dee2e6;
}

[data-theme="dark"] {
  /* Dark theme overrides */
  --color-background: #1a1a1a;
  --color-surface: #2d2d2d;
  --color-text-primary: #ffffff;
  --color-text-secondary: #b0b0b0;
  --color-border: #404040;
}
```

## Verification Steps
1. Test responsive behavior - should adapt to different screen sizes
2. Verify theme switching - should apply light/dark themes correctly
3. Test navigation - should highlight active sections
4. Verify accessibility - should support keyboard navigation and screen readers
5. Test on different browsers - should display consistently
6. Verify loading states - should show appropriate placeholders
7. Test with long content - should handle overflow properly
8. Verify mobile usability - should be touch-friendly

## Dependencies
- T004: Basic Frontend Setup
- T005: Wails Frontend-Backend Integration

## Notes
- Follow modern CSS best practices
- Ensure good performance on lower-end devices
- Consider future components when designing layout
- Implement proper semantic HTML structure
- Plan for internationalization (text expansion)
- Consider print styles for reporting features