# T034: Responsive Design Implementation

## Overview
Implement comprehensive responsive design to ensure NetMonitor works seamlessly across desktop, tablet, and mobile devices with optimized layouts and interactions.

## Context
NetMonitor must be accessible and functional on various device sizes. The dashboard needs to adapt gracefully from large desktop monitors to small mobile screens while maintaining usability and visual hierarchy.

## Task Description
Create a fully responsive design system with adaptive layouts, touch-friendly interfaces, mobile-specific optimizations, and cross-device consistency.

## Acceptance Criteria
- [ ] Responsive grid system that adapts to all screen sizes
- [ ] Mobile-first CSS architecture with progressive enhancement
- [ ] Touch-friendly interface elements and interactions
- [ ] Optimized typography scaling across devices
- [ ] Mobile-specific navigation and menu systems
- [ ] Tablet-optimized layouts for medium screens
- [ ] Performance optimization for mobile devices
- [ ] Cross-browser compatibility on mobile platforms
- [ ] Accessibility compliance on all device types

## Responsive Breakpoint System
```css
/* Mobile-first breakpoints */
:root {
  --breakpoint-xs: 0px;      /* Extra small devices */
  --breakpoint-sm: 576px;    /* Small devices (phones) */
  --breakpoint-md: 768px;    /* Medium devices (tablets) */
  --breakpoint-lg: 992px;    /* Large devices (small laptops) */
  --breakpoint-xl: 1200px;   /* Extra large devices (desktops) */
  --breakpoint-xxl: 1400px;  /* Extra extra large devices */

  /* Container max-widths */
  --container-sm: 540px;
  --container-md: 720px;
  --container-lg: 960px;
  --container-xl: 1140px;
  --container-xxl: 1320px;

  /* Spacing scale */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 3rem;

  /* Typography scale */
  --font-size-xs: 0.75rem;
  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.25rem;
  --font-size-2xl: 1.5rem;
  --font-size-3xl: 1.875rem;
}

/* Responsive mixins using CSS custom properties */
@media (max-width: 575.98px) {
  :root {
    --container-width: 100%;
    --grid-columns: 1;
    --font-size-scale: 0.9;
  }
}

@media (min-width: 576px) and (max-width: 767.98px) {
  :root {
    --container-width: var(--container-sm);
    --grid-columns: 2;
    --font-size-scale: 0.95;
  }
}

@media (min-width: 768px) and (max-width: 991.98px) {
  :root {
    --container-width: var(--container-md);
    --grid-columns: 3;
    --font-size-scale: 1;
  }
}

@media (min-width: 992px) and (max-width: 1199.98px) {
  :root {
    --container-width: var(--container-lg);
    --grid-columns: 4;
    --font-size-scale: 1;
  }
}

@media (min-width: 1200px) {
  :root {
    --container-width: var(--container-xl);
    --grid-columns: 6;
    --font-size-scale: 1;
  }
}
```

## Mobile-First Layout System
```css
/* Base container */
.container {
  width: 100%;
  max-width: var(--container-width);
  margin: 0 auto;
  padding: 0 var(--spacing-md);
}

/* Responsive grid system */
.grid {
  display: grid;
  gap: var(--spacing-md);
  grid-template-columns: 1fr; /* Mobile: single column */
}

@media (min-width: 576px) {
  .grid {
    grid-template-columns: repeat(2, 1fr); /* Small: 2 columns */
  }
}

@media (min-width: 768px) {
  .grid {
    grid-template-columns: repeat(3, 1fr); /* Medium: 3 columns */
  }
}

@media (min-width: 992px) {
  .grid {
    grid-template-columns: repeat(4, 1fr); /* Large: 4 columns */
  }
}

@media (min-width: 1200px) {
  .grid {
    grid-template-columns: repeat(6, 1fr); /* XL: 6 columns */
  }
}

/* Flexible grid items */
.grid-item {
  grid-column: span 1;
}

.grid-item-2 { grid-column: span 2; }
.grid-item-3 { grid-column: span 3; }
.grid-item-4 { grid-column: span 4; }
.grid-item-full { grid-column: 1 / -1; }

/* Responsive utilities */
.hidden-xs { display: none; }
.hidden-sm { display: none; }

@media (min-width: 576px) {
  .hidden-xs { display: block; }
  .visible-xs { display: none; }
}

@media (min-width: 768px) {
  .hidden-sm { display: block; }
  .visible-sm { display: none; }
}

@media (min-width: 992px) {
  .hidden-md { display: block; }
  .visible-md { display: none; }
}
```

## Mobile Navigation System
```html
<nav class="mobile-nav">
  <div class="nav-header">
    <div class="nav-brand">
      <h1 class="brand-title">NetMonitor</h1>
    </div>
    <button class="nav-toggle" aria-label="Toggle navigation" aria-expanded="false">
      <span class="hamburger">
        <span class="hamburger-line"></span>
        <span class="hamburger-line"></span>
        <span class="hamburger-line"></span>
      </span>
    </button>
  </div>

  <div class="nav-menu" aria-hidden="true">
    <div class="nav-overlay"></div>
    <div class="nav-content">
      <div class="nav-header-mobile">
        <h2>Menu</h2>
        <button class="nav-close" aria-label="Close menu">×</button>
      </div>

      <ul class="nav-links">
        <li class="nav-item">
          <a href="#overview" class="nav-link active">
            <span class="nav-icon">📊</span>
            <span class="nav-text">Overview</span>
          </a>
        </li>
        <li class="nav-item">
          <a href="#regions" class="nav-link">
            <span class="nav-icon">🌍</span>
            <span class="nav-text">Regions</span>
          </a>
        </li>
        <li class="nav-item">
          <a href="#endpoints" class="nav-link">
            <span class="nav-icon">🎯</span>
            <span class="nav-text">Endpoints</span>
          </a>
        </li>
        <li class="nav-item">
          <a href="#manual" class="nav-link">
            <span class="nav-icon">⚡</span>
            <span class="nav-text">Manual Tests</span>
          </a>
        </li>
        <li class="nav-item">
          <a href="#settings" class="nav-link">
            <span class="nav-icon">⚙️</span>
            <span class="nav-text">Settings</span>
          </a>
        </li>
      </ul>

      <div class="nav-footer">
        <div class="connection-status">
          <span class="status-indicator connected"></span>
          <span class="status-text">Connected</span>
        </div>
        <div class="theme-toggle-mobile">
          <button class="theme-btn">🌙 Dark Mode</button>
        </div>
      </div>
    </div>
  </div>
</nav>
```

## Responsive Dashboard Layout
```css
/* Dashboard layout adaptation */
.dashboard {
  display: grid;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
}

/* Mobile layout: single column */
.dashboard {
  grid-template-areas:
    "status"
    "charts"
    "endpoints";
  grid-template-columns: 1fr;
}

/* Tablet layout: 2 columns */
@media (min-width: 768px) {
  .dashboard {
    grid-template-areas:
      "status status"
      "charts charts"
      "endpoints endpoints";
    grid-template-columns: 1fr 1fr;
  }
}

/* Desktop layout: sidebar + content */
@media (min-width: 992px) {
  .dashboard {
    grid-template-areas:
      "sidebar status status"
      "sidebar charts charts"
      "sidebar endpoints endpoints";
    grid-template-columns: 250px 1fr 1fr;
  }

  .sidebar {
    position: sticky;
    top: var(--spacing-md);
    height: fit-content;
  }
}

/* Component responsive behavior */
.status-widgets {
  display: grid;
  gap: var(--spacing-sm);
  grid-template-columns: 1fr; /* Mobile: stack widgets */
}

@media (min-width: 576px) {
  .status-widgets {
    grid-template-columns: repeat(2, 1fr); /* Small: 2 per row */
  }
}

@media (min-width: 992px) {
  .status-widgets {
    grid-template-columns: repeat(4, 1fr); /* Desktop: 4 per row */
  }
}
```

## Touch-Friendly Interface Elements
```css
/* Touch targets - minimum 44px */
.touch-target {
  min-height: 44px;
  min-width: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Buttons */
.btn {
  padding: var(--spacing-sm) var(--spacing-md);
  min-height: 44px;
  border-radius: 8px;
  font-size: var(--font-size-base);
  touch-action: manipulation; /* Disable double-tap zoom */
}

/* Form inputs */
.form-input {
  padding: var(--spacing-sm) var(--spacing-md);
  min-height: 44px;
  font-size: 16px; /* Prevent iOS zoom */
  border-radius: 8px;
}

/* Interactive elements */
.interactive {
  cursor: pointer;
  transition: transform var(--transition-fast);
}

.interactive:hover {
  transform: translateY(-1px);
}

.interactive:active {
  transform: translateY(0);
}

/* Touch feedback */
@media (hover: none) and (pointer: coarse) {
  .interactive:active {
    background-color: var(--color-surface-hover);
  }
}
```

## Mobile-Optimized Components
```javascript
class ResponsiveComponent {
  constructor(element) {
    this.element = element;
    this.breakpoints = {
      xs: 0,
      sm: 576,
      md: 768,
      lg: 992,
      xl: 1200
    };
    this.currentBreakpoint = this.getCurrentBreakpoint();

    this.init();
  }

  init() {
    this.setupResizeListener();
    this.adaptToBreakpoint();
  }

  setupResizeListener() {
    let resizeTimer;
    window.addEventListener('resize', () => {
      clearTimeout(resizeTimer);
      resizeTimer = setTimeout(() => {
        const newBreakpoint = this.getCurrentBreakpoint();
        if (newBreakpoint !== this.currentBreakpoint) {
          this.currentBreakpoint = newBreakpoint;
          this.adaptToBreakpoint();
        }
      }, 150);
    });
  }

  getCurrentBreakpoint() {
    const width = window.innerWidth;
    if (width >= this.breakpoints.xl) return 'xl';
    if (width >= this.breakpoints.lg) return 'lg';
    if (width >= this.breakpoints.md) return 'md';
    if (width >= this.breakpoints.sm) return 'sm';
    return 'xs';
  }

  adaptToBreakpoint() {
    // Remove existing breakpoint classes
    Object.keys(this.breakpoints).forEach(bp => {
      this.element.classList.remove(`breakpoint-${bp}`);
    });

    // Add current breakpoint class
    this.element.classList.add(`breakpoint-${this.currentBreakpoint}`);

    // Trigger breakpoint-specific adaptations
    this.onBreakpointChange(this.currentBreakpoint);
  }

  onBreakpointChange(breakpoint) {
    switch (breakpoint) {
      case 'xs':
      case 'sm':
        this.configureMobile();
        break;
      case 'md':
        this.configureTablet();
        break;
      case 'lg':
      case 'xl':
        this.configureDesktop();
        break;
    }
  }

  configureMobile() {
    // Mobile-specific configurations
    this.element.setAttribute('data-mobile', 'true');
    this.enableSwipeGestures();
    this.optimizeForTouch();
  }

  configureTablet() {
    // Tablet-specific configurations
    this.element.removeAttribute('data-mobile');
    this.enableHybridInteractions();
  }

  configureDesktop() {
    // Desktop-specific configurations
    this.element.removeAttribute('data-mobile');
    this.disableSwipeGestures();
    this.optimizeForMouse();
  }

  enableSwipeGestures() {
    // Implement swipe gesture handling for mobile
    let startX, startY, startTime;

    this.element.addEventListener('touchstart', (e) => {
      const touch = e.touches[0];
      startX = touch.clientX;
      startY = touch.clientY;
      startTime = Date.now();
    });

    this.element.addEventListener('touchend', (e) => {
      if (!startX || !startY) return;

      const touch = e.changedTouches[0];
      const deltaX = touch.clientX - startX;
      const deltaY = touch.clientY - startY;
      const deltaTime = Date.now() - startTime;

      // Detect swipe gestures
      if (Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > 50 && deltaTime < 300) {
        if (deltaX > 0) {
          this.onSwipeRight();
        } else {
          this.onSwipeLeft();
        }
      }

      startX = startY = startTime = null;
    });
  }

  optimizeForTouch() {
    // Increase touch target sizes
    const interactiveElements = this.element.querySelectorAll('button, a, input, select');
    interactiveElements.forEach(el => {
      el.classList.add('touch-optimized');
    });
  }
}
```

## Performance Optimizations
```css
/* Optimize for mobile performance */
.performance-optimized {
  /* Use transforms for animations (GPU accelerated) */
  transform: translateZ(0);

  /* Optimize repaints */
  will-change: transform;

  /* Reduce layout thrashing */
  contain: layout style paint;
}

/* Optimize images for different screen densities */
.responsive-image {
  width: 100%;
  height: auto;
  object-fit: cover;
}

@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
  .responsive-image {
    /* High DPI optimizations */
  }
}

/* Reduce motion for users who prefer it */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* Dark mode optimizations for OLED screens */
@media (prefers-color-scheme: dark) {
  :root {
    /* Use true black for OLED power savings */
    --color-background: #000000;
  }
}
```

## Verification Steps
1. Test on various device sizes - should adapt layout appropriately
2. Verify touch interactions - should work smoothly on mobile devices
3. Test navigation - should provide mobile-friendly menu system
4. Verify typography scaling - should be readable on all screen sizes
5. Test performance - should load quickly on mobile networks
6. Verify accessibility - should support screen readers and keyboard navigation
7. Test cross-browser compatibility - should work on mobile browsers
8. Verify landscape/portrait orientation handling

## Dependencies
- T026: Dashboard Layout and Structure
- T027: Status Overview Widgets
- T028: Interactive Latency Graphs
- T029: Endpoint Status Grid

## Notes
- Use progressive enhancement approach
- Test on real devices, not just browser dev tools
- Consider touch gesture libraries for advanced interactions
- Optimize images and assets for mobile
- Plan for future PWA implementation
- Consider implementing offline functionality for mobile users