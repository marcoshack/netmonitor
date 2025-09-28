import './style.css';
import './app.css';
import '../css/themes.css';
import '../css/main.css';

import { NetMonitorApp } from '../js/main.js';

// Initialize the application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    console.log('NetMonitor frontend initializing...');
    window.netMonitorApp = new NetMonitorApp();
    window.netMonitorApp.init();
});

// Export for debugging
window.NetMonitorApp = NetMonitorApp;