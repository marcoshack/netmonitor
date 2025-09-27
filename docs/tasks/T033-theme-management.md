# T033: Theme Management System

## Overview
Implement comprehensive theme management with light/dark modes, automatic system theme detection, custom theme creation, and persistent theme preferences.

## Context
NetMonitor needs to support both light and dark themes with automatic system detection. Users should be able to manually override theme selection and potentially customize themes for better accessibility and personal preference.

## Task Description
Create a robust theme management system that handles theme switching, persistence, system integration, and provides a foundation for future theme customization.

## Acceptance Criteria
- [ ] Light and dark theme implementations
- [ ] Automatic system theme detection
- [ ] Manual theme override capability
- [ ] Theme persistence across sessions
- [ ] Smooth theme transition animations
- [ ] System theme change detection and response
- [ ] High contrast mode support for accessibility
- [ ] Theme-aware component styling
- [ ] Theme export/import for sharing

## Theme System Architecture
```css
/* CSS Custom Properties for theming */
:root {
  /* Light theme (default) */
  --theme-name: 'light';

  /* Primary colors */
  --color-primary: #007bff;
  --color-primary-hover: #0056b3;
  --color-primary-light: #e3f2fd;

  /* Status colors */
  --color-success: #28a745;
  --color-warning: #ffc107;
  --color-danger: #dc3545;
  --color-info: #17a2b8;

  /* Background colors */
  --color-background: #ffffff;
  --color-surface: #f8f9fa;
  --color-surface-hover: #e9ecef;
  --color-card: #ffffff;

  /* Text colors */
  --color-text-primary: #212529;
  --color-text-secondary: #6c757d;
  --color-text-muted: #adb5bd;
  --color-text-inverse: #ffffff;

  /* Border colors */
  --color-border: #dee2e6;
  --color-border-light: #f1f3f4;
  --color-divider: #e9ecef;

  /* Shadow */
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.1);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.1);
  --shadow-lg: 0 10px 15px rgba(0,0,0,0.1);

  /* Transitions */
  --transition-fast: 150ms ease;
  --transition-normal: 250ms ease;
  --transition-slow: 400ms ease;
}

[data-theme="dark"] {
  --theme-name: 'dark';

  /* Primary colors */
  --color-primary: #4dabf7;
  --color-primary-hover: #339af0;
  --color-primary-light: #1c2d41;

  /* Background colors */
  --color-background: #1a1a1a;
  --color-surface: #2d2d2d;
  --color-surface-hover: #404040;
  --color-card: #333333;

  /* Text colors */
  --color-text-primary: #ffffff;
  --color-text-secondary: #b0b0b0;
  --color-text-muted: #808080;
  --color-text-inverse: #000000;

  /* Border colors */
  --color-border: #404040;
  --color-border-light: #333333;
  --color-divider: #2d2d2d;

  /* Darker shadows for dark theme */
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.3);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.3);
  --shadow-lg: 0 10px 15px rgba(0,0,0,0.3);
}

[data-theme="high-contrast"] {
  --theme-name: 'high-contrast';

  /* High contrast colors */
  --color-primary: #0000ff;
  --color-background: #ffffff;
  --color-surface: #ffffff;
  --color-text-primary: #000000;
  --color-border: #000000;

  /* Remove subtle variations for high contrast */
  --color-text-secondary: #000000;
  --color-surface-hover: #f0f0f0;
  --shadow-sm: none;
  --shadow-md: 0 2px 4px #000000;
  --shadow-lg: 0 4px 8px #000000;
}

/* Smooth theme transitions */
* {
  transition:
    background-color var(--transition-normal),
    border-color var(--transition-normal),
    color var(--transition-normal),
    box-shadow var(--transition-normal);
}
```

## Theme Toggle Component
```html
<div class="theme-selector">
  <button class="theme-toggle" aria-label="Toggle theme" title="Toggle theme">
    <span class="theme-icon">
      <span class="sun-icon">☀️</span>
      <span class="moon-icon">🌙</span>
      <span class="auto-icon">🔄</span>
    </span>
  </button>

  <div class="theme-menu" style="display: none;">
    <div class="theme-menu-header">
      <h3>Theme Settings</h3>
    </div>

    <div class="theme-options">
      <label class="theme-option">
        <input type="radio" name="theme" value="light">
        <div class="option-content">
          <span class="option-icon">☀️</span>
          <span class="option-label">Light</span>
        </div>
      </label>

      <label class="theme-option">
        <input type="radio" name="theme" value="dark">
        <div class="option-content">
          <span class="option-icon">🌙</span>
          <span class="option-label">Dark</span>
        </div>
      </label>

      <label class="theme-option">
        <input type="radio" name="theme" value="auto" checked>
        <div class="option-content">
          <span class="option-icon">🔄</span>
          <span class="option-label">Auto</span>
        </div>
      </label>

      <label class="theme-option">
        <input type="radio" name="theme" value="high-contrast">
        <div class="option-content">
          <span class="option-icon">🔳</span>
          <span class="option-label">High Contrast</span>
        </div>
      </label>
    </div>

    <div class="theme-preview">
      <div class="preview-card">
        <div class="preview-header">Preview</div>
        <div class="preview-content">
          <div class="preview-text">Sample text</div>
          <div class="preview-button">Button</div>
        </div>
      </div>
    </div>

    <div class="theme-advanced">
      <button class="theme-customize-btn">Customize Colors</button>
      <button class="theme-export-btn">Export Theme</button>
      <button class="theme-import-btn">Import Theme</button>
    </div>
  </div>
</div>
```

## Theme Manager Class
```javascript
class ThemeManager {
  constructor(options = {}) {
    this.options = {
      storageKey: 'netmonitor-theme',
      defaultTheme: 'auto',
      enableCustomThemes: true,
      ...options
    };

    this.currentTheme = null;
    this.systemTheme = null;
    this.mediaQuery = null;
    this.callbacks = {
      onThemeChange: options.onThemeChange || (() => {})
    };

    this.themes = {
      light: { name: 'Light', icon: '☀️' },
      dark: { name: 'Dark', icon: '🌙' },
      auto: { name: 'Auto', icon: '🔄' },
      'high-contrast': { name: 'High Contrast', icon: '🔳' }
    };

    this.init();
  }

  init() {
    this.detectSystemTheme();
    this.loadSavedTheme();
    this.setupSystemThemeDetection();
    this.bindEvents();
    this.applyTheme();
  }

  detectSystemTheme() {
    if (window.matchMedia) {
      this.mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      this.systemTheme = this.mediaQuery.matches ? 'dark' : 'light';
    } else {
      this.systemTheme = 'light';
    }
  }

  setupSystemThemeDetection() {
    if (this.mediaQuery) {
      this.mediaQuery.addListener((e) => {
        this.systemTheme = e.matches ? 'dark' : 'light';
        if (this.currentTheme === 'auto') {
          this.applyTheme();
        }
      });
    }

    // Detect high contrast mode
    if (window.matchMedia('(prefers-contrast: high)').matches) {
      this.systemTheme = 'high-contrast';
    }
  }

  loadSavedTheme() {
    try {
      const saved = localStorage.getItem(this.options.storageKey);
      this.currentTheme = saved || this.options.defaultTheme;
    } catch (error) {
      console.warn('Failed to load saved theme:', error);
      this.currentTheme = this.options.defaultTheme;
    }
  }

  saveTheme(theme) {
    try {
      localStorage.setItem(this.options.storageKey, theme);
    } catch (error) {
      console.warn('Failed to save theme:', error);
    }
  }

  setTheme(theme) {
    if (!this.themes[theme]) {
      console.warn(`Unknown theme: ${theme}`);
      return;
    }

    this.currentTheme = theme;
    this.saveTheme(theme);
    this.applyTheme();
    this.updateThemeToggle();
    this.callbacks.onThemeChange(this.getEffectiveTheme());
  }

  applyTheme() {
    const effectiveTheme = this.getEffectiveTheme();

    // Remove all theme classes
    document.documentElement.removeAttribute('data-theme');

    // Apply new theme
    if (effectiveTheme !== 'light') {
      document.documentElement.setAttribute('data-theme', effectiveTheme);
    }

    // Update meta theme-color for mobile browsers
    this.updateMetaThemeColor(effectiveTheme);

    // Update CSS custom property for JavaScript access
    document.documentElement.style.setProperty('--current-theme', effectiveTheme);
  }

  getEffectiveTheme() {
    if (this.currentTheme === 'auto') {
      return this.systemTheme;
    }
    return this.currentTheme;
  }

  updateMetaThemeColor(theme) {
    let metaThemeColor = document.querySelector('meta[name="theme-color"]');
    if (!metaThemeColor) {
      metaThemeColor = document.createElement('meta');
      metaThemeColor.name = 'theme-color';
      document.head.appendChild(metaThemeColor);
    }

    const colors = {
      light: '#ffffff',
      dark: '#1a1a1a',
      'high-contrast': '#ffffff'
    };

    metaThemeColor.content = colors[theme] || colors.light;
  }

  bindEvents() {
    // Theme toggle click
    document.addEventListener('click', (e) => {
      if (e.target.closest('.theme-toggle')) {
        this.toggleThemeMenu();
      }
    });

    // Theme option selection
    document.addEventListener('change', (e) => {
      if (e.target.name === 'theme') {
        this.setTheme(e.target.value);
      }
    });

    // Close menu on outside click
    document.addEventListener('click', (e) => {
      if (!e.target.closest('.theme-selector')) {
        this.closeThemeMenu();
      }
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
          case 't':
            this.cycleTheme();
            e.preventDefault();
            break;
        }
      }
    });
  }

  toggleThemeMenu() {
    const menu = document.querySelector('.theme-menu');
    const isVisible = menu.style.display !== 'none';

    if (isVisible) {
      this.closeThemeMenu();
    } else {
      this.openThemeMenu();
    }
  }

  openThemeMenu() {
    const menu = document.querySelector('.theme-menu');
    menu.style.display = 'block';

    // Update radio button selection
    const radio = menu.querySelector(`input[value="${this.currentTheme}"]`);
    if (radio) {
      radio.checked = true;
    }
  }

  closeThemeMenu() {
    document.querySelector('.theme-menu').style.display = 'none';
  }

  cycleTheme() {
    const themes = Object.keys(this.themes);
    const currentIndex = themes.indexOf(this.currentTheme);
    const nextIndex = (currentIndex + 1) % themes.length;
    this.setTheme(themes[nextIndex]);
  }

  updateThemeToggle() {
    const toggle = document.querySelector('.theme-toggle');
    const icon = toggle.querySelector('.theme-icon');

    // Update icon based on effective theme
    const effectiveTheme = this.getEffectiveTheme();
    icon.className = `theme-icon ${effectiveTheme}-theme`;

    // Update aria-label
    toggle.setAttribute('aria-label', `Current theme: ${this.themes[this.currentTheme].name}`);
  }

  // Theme export/import functionality
  exportTheme() {
    const computedStyle = getComputedStyle(document.documentElement);
    const themeData = {
      name: this.currentTheme,
      timestamp: new Date().toISOString(),
      variables: {}
    };

    // Extract CSS custom properties
    const props = Array.from(document.styleSheets)
      .flatMap(sheet => Array.from(sheet.cssRules))
      .filter(rule => rule.selectorText === ':root')
      .flatMap(rule => Array.from(rule.style))
      .filter(prop => prop.startsWith('--color-') || prop.startsWith('--shadow-'));

    props.forEach(prop => {
      themeData.variables[prop] = computedStyle.getPropertyValue(prop).trim();
    });

    return JSON.stringify(themeData, null, 2);
  }

  importTheme(themeJson) {
    try {
      const themeData = JSON.parse(themeJson);

      // Validate theme data
      if (!themeData.variables || typeof themeData.variables !== 'object') {
        throw new Error('Invalid theme format');
      }

      // Apply custom properties
      Object.entries(themeData.variables).forEach(([prop, value]) => {
        if (prop.startsWith('--color-') || prop.startsWith('--shadow-')) {
          document.documentElement.style.setProperty(prop, value);
        }
      });

      return true;
    } catch (error) {
      console.error('Failed to import theme:', error);
      return false;
    }
  }

  // Get current theme info
  getCurrentTheme() {
    return {
      selected: this.currentTheme,
      effective: this.getEffectiveTheme(),
      system: this.systemTheme
    };
  }
}

// Initialize theme manager
const themeManager = new ThemeManager({
  onThemeChange: (theme) => {
    // Notify other components of theme change
    document.dispatchEvent(new CustomEvent('themechange', {
      detail: { theme }
    }));
  }
});

// Export for global access
window.themeManager = themeManager;
```

## Theme-Aware Component Example
```javascript
class ThemeAwareComponent {
  constructor(element) {
    this.element = element;
    this.init();
  }

  init() {
    // Listen for theme changes
    document.addEventListener('themechange', (e) => {
      this.handleThemeChange(e.detail.theme);
    });

    // Initial theme setup
    this.handleThemeChange(themeManager.getEffectiveTheme());
  }

  handleThemeChange(theme) {
    // Update component based on theme
    this.element.classList.toggle('dark-theme', theme === 'dark');
    this.element.classList.toggle('high-contrast-theme', theme === 'high-contrast');

    // Update any theme-specific logic
    this.updateChartColors(theme);
  }

  updateChartColors(theme) {
    // Example: Update chart.js colors based on theme
    if (this.chart) {
      const colors = this.getThemeColors(theme);
      this.chart.options.scales.x.grid.color = colors.gridColor;
      this.chart.options.scales.y.grid.color = colors.gridColor;
      this.chart.update('none');
    }
  }

  getThemeColors(theme) {
    const style = getComputedStyle(document.documentElement);
    return {
      gridColor: style.getPropertyValue('--color-border').trim(),
      textColor: style.getPropertyValue('--color-text-secondary').trim(),
      backgroundColor: style.getPropertyValue('--color-background').trim()
    };
  }
}
```

## Verification Steps
1. Test theme switching - should change colors smoothly
2. Verify system theme detection - should follow OS theme preference
3. Test theme persistence - should remember selection across sessions
4. Verify accessibility - should support high contrast mode
5. Test keyboard shortcuts - should cycle themes with Ctrl+T
6. Verify mobile integration - should update mobile browser theme color
7. Test theme export/import - should save and restore custom themes
8. Verify component integration - should update all themed components

## Dependencies
- T026: Dashboard Layout and Structure
- T004: Basic Frontend Setup

## Notes
- Implement smooth transitions between themes
- Consider user preferences for reduced motion
- Plan for future custom theme creation tools
- Ensure accessibility compliance for all themes
- Test with various system theme configurations
- Consider implementing seasonal theme variations