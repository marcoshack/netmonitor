// Theme Management Module

class ThemeManager {
    constructor() {
        this.currentTheme = 'auto';
        this.systemTheme = 'light';
        this.storageKey = 'netmonitor-theme';
        this.mediaQuery = null;
    }

    async init() {
        this.detectSystemTheme();
        this.loadSavedTheme();
        this.setupSystemThemeDetection();
        this.applyTheme();
        console.log('Theme manager initialized');
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
            this.mediaQuery.addEventListener('change', (e) => {
                this.systemTheme = e.matches ? 'dark' : 'light';
                if (this.currentTheme === 'auto') {
                    this.applyTheme();
                }
            });
        }

        // Detect high contrast mode
        const highContrastQuery = window.matchMedia('(prefers-contrast: high)');
        highContrastQuery.addEventListener('change', () => {
            this.applyTheme();
        });
    }

    loadSavedTheme() {
        try {
            const saved = localStorage.getItem(this.storageKey);
            this.currentTheme = saved || 'auto';
        } catch (error) {
            console.warn('Failed to load saved theme:', error);
            this.currentTheme = 'auto';
        }
    }

    saveTheme(theme) {
        try {
            localStorage.setItem(this.storageKey, theme);
        } catch (error) {
            console.warn('Failed to save theme:', error);
        }
    }

    setTheme(theme) {
        const validThemes = ['light', 'dark', 'auto', 'high-contrast'];
        if (!validThemes.includes(theme)) {
            console.warn(`Invalid theme: ${theme}`);
            return;
        }

        this.currentTheme = theme;
        this.saveTheme(theme);
        this.applyTheme();
        this.updateThemeToggle();
        
        // Notify backend of theme change
        this.notifyBackend(theme);
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

        console.log(`Theme applied: ${effectiveTheme}`);
    }

    getEffectiveTheme() {
        if (this.currentTheme === 'auto') {
            // Check for high contrast preference
            if (window.matchMedia && window.matchMedia('(prefers-contrast: high)').matches) {
                return 'high-contrast';
            }
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

    updateThemeToggle() {
        const themeToggle = document.getElementById('themeToggle');
        if (!themeToggle) return;

        const effectiveTheme = this.getEffectiveTheme();
        
        // Update icon based on effective theme
        const icons = {
            light: '☀️',
            dark: '🌙',
            'high-contrast': '🔳',
            auto: '🔄'
        };

        themeToggle.textContent = icons[effectiveTheme] || '🔄';
        themeToggle.setAttribute('aria-label', `Current theme: ${this.currentTheme}`);
        themeToggle.setAttribute('title', `Current theme: ${this.currentTheme} (${effectiveTheme})`);
    }

    toggleTheme() {
        const themes = ['light', 'dark', 'auto'];
        const currentIndex = themes.indexOf(this.currentTheme);
        const nextIndex = (currentIndex + 1) % themes.length;
        this.setTheme(themes[nextIndex]);
    }

    async notifyBackend(theme) {
        try {
            if (window.go && window.go.main && window.go.main.App) {
                await window.go.main.App.SetTheme(theme);
            }
        } catch (error) {
            console.warn('Failed to notify backend of theme change:', error);
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

    // Add theme change listener
    onThemeChange(callback) {
        this._themeChangeCallback = callback;
    }

    // Notify theme change listeners
    notifyThemeChange() {
        if (this._themeChangeCallback) {
            this._themeChangeCallback(this.getCurrentTheme());
        }

        // Dispatch custom event
        document.dispatchEvent(new CustomEvent('themechange', {
            detail: this.getCurrentTheme()
        }));
    }
}

export { ThemeManager };