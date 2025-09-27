# T004: Basic Frontend Setup

## Overview
Set up the basic web frontend structure with HTML, CSS, and JavaScript foundation for the NetMonitor dashboard.

## Context
NetMonitor uses a web-based frontend that will display network monitoring data, graphs, and controls. The frontend needs to be responsive and support both light and dark themes.

## Task Description
Create the basic frontend structure with HTML layout, CSS styling foundation, and JavaScript framework setup.

## Acceptance Criteria
- [ ] Basic HTML structure with semantic layout
- [ ] CSS foundation with CSS variables for theming
- [ ] JavaScript module structure established
- [ ] Basic theme switching capability (light/dark)
- [ ] Responsive design foundations
- [ ] Frontend builds and serves correctly through Wails

## Frontend Structure
- `frontend/index.html` - Main application page
- `frontend/css/` - Stylesheets
  - `main.css` - Main styles
  - `themes.css` - Theme definitions
- `frontend/js/` - JavaScript modules
  - `main.js` - Application entry point
  - `api.js` - Backend communication
  - `theme.js` - Theme management

## Verification Steps
1. Open application - should display basic HTML layout
2. Toggle theme - should switch between light and dark
3. Resize window - should be responsive
4. Check browser console - should be free of errors
5. Verify CSS variables are working for theming

## Dependencies
- T001: Project Setup and Initialization

## Notes
- Use modern CSS (Grid, Flexbox, CSS variables)
- Prepare for Chart.js integration for graphs
- Follow semantic HTML practices
- Ensure accessibility considerations