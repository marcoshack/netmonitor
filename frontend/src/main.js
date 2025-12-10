import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns'; // If using time scale, often needed, but let's check if chart.js/auto includes it or if we need an adapter.
// Actually handling dates usually requires an adapter in Chart.js v3+.
// The user installed chart.js but maybe not an adapter.
// Let's check package.json again. 
// "chart.js": "^4.5.1". Chart.js v4 requires a date adapter for time text.
// If I use 'time' scale, I need 'chartjs-adapter-date-fns' or 'luxon' etc.
// But I might not have installed it.
// I ran `npm install chart.js`. I did NOT install an adapter.
// This might be another reason why graphs fail if I used `type: 'time'`.
// Let's check if I used `type: 'time'`. Yes I did.
// So I need to install a date adapter too.
// For now, let's fix the import. I will also install the adapter.

let currentConfig = null;
let currentRegion = null;
let testResults = {}; // Map endpointID -> Array of results
const maxHistory = 30; // Points on graph

// Wails Runtime and Backend variables (injected by Wails)
// We assume window.go.main.App is available

window.addEventListener('DOMContentLoaded', init);

async function init() {
    try {
        // Load Config
        currentConfig = await window.go.main.App.GetConfig();

        // Setup Regions
        setupRegionSelector();

        // Setup Event Listeners
        window.runtime.EventsOn("test-result", handleTestResult);

        setupSettings();

        // Initial Layout
        renderDashboard();

        // Update status
        document.getElementById("status-message").innerText = "Monitoring Active";

    } catch (err) {
        console.error("Initialization failed", err);
        document.getElementById("status-message").innerText = "Error: " + err;
    }
}

function setupRegionSelector() {
    const selector = document.getElementById("region-select");
    selector.innerHTML = "";

    // Get region names
    const regions = Object.keys(currentConfig.regions);
    if (regions.length === 0) return;

    regions.forEach(r => {
        const opt = document.createElement("option");
        opt.value = r;
        opt.innerText = r;
        selector.appendChild(opt);
    });

    // Default to first or "Default"
    if (regions.includes("Default")) {
        currentRegion = "Default";
        selector.value = "Default";
    } else {
        currentRegion = regions[0];
    }

    selector.addEventListener("change", (e) => {
        currentRegion = e.target.value;
        renderDashboard();
    });

    // Time Range Selector
    const timeSelector = document.getElementById("time-range-select");
    timeSelector.addEventListener("change", (e) => {
        const range = e.target.value;
        fetchHistory(range);
    });

    // Initial fetch
    fetchHistory(timeSelector.value);
}

function renderDashboard() {
    const grid = document.getElementById("endpoints-grid");
    grid.innerHTML = ""; // Clear existing

    if (!currentRegion || !currentConfig.regions[currentRegion]) return;

    const endpoints = currentConfig.regions[currentRegion].endpoints;

    endpoints.forEach(ep => {
        const card = createEndpointCard(ep);
        grid.appendChild(card);
    });
}

function createEndpointCard(ep) {
    const id = `${currentRegion}-${ep.name}`; // Consistent ID generation

    const div = document.createElement("div");
    div.className = "card";
    div.id = `card-${id}`;

    div.innerHTML = `
        <div class="card-header">
            <div class="flex items-center gap-sm">
                <div id="status-dot-${id}" class="status-dot"></div>
                <div class="text-lg">${ep.name}</div>
            </div>
            <div class="text-sm text-muted">${ep.type}</div>
        </div>
        
        <div class="flex items-center justify-between" style="margin-bottom: 0.5rem">
            <div class="text-sm text-muted">Latency</div>
            <div class="latency-value"><span id="latency-${id}">--</span> <span class="text-sm text-muted font-normal">ms</span></div>
        </div>

        <div class="chart-container">
            <canvas id="canvas-${id}"></canvas>
        </div>
    `;

    // Initialize chart for this card
    // We defer chart init slightly to ensure DOM is ready? No, div is not attached yet? 
    // It will be attached.
    // We'll store chart instances in a map effectively via closure or global
    // But canvas needs to be in DOM or at least created.

    return div;
}

// Chart.js helper
const chartInstances = {}; // id -> Chart instance
let historyData = {}; // endpointId -> []{timestamp, latency}

async function fetchHistory(range) {
    historyData = {}; // Clear
    try {
        const results = await window.go.main.App.GetHistoryRange(range);
        // Process results into map
        results.forEach(r => {
            const id = r.endpoint_id;
            if (!historyData[id]) historyData[id] = [];
            historyData[id].push({
                x: new Date(r.timestamp),
                y: r.latency_ms
            });
        });

        // Refresh all charts
        Object.keys(chartInstances).forEach(id => {
            updateChartHistory(id);
        });

        // Also ensure charts are initialized if they don't exist yet (e.g. on first load)
        // Iterate over current DOM cards to init them
        const cards = document.querySelectorAll('[id^="card-"]');
        cards.forEach(card => {
            const id = card.id.replace("card-", "");
            if (!chartInstances[id]) {
                initChart(id);
            } else {
                updateChartHistory(id);
            }
        });

    } catch (err) {
        console.error("Failed to fetch history", err);
    }
}

function updateChartHistory(id) {
    const chart = chartInstances[id];
    if (!chart) return;

    // reset data
    const data = historyData[id] || [];
    // sort by time just in case, though backend should be ok? backend was daily files.
    data.sort((a, b) => a.x - b.x);

    chart.data.datasets[0].data = data;
    chart.update();
}

function initChart(id) {
    if (chartInstances[id]) return chartInstances[id];

    const canvas = document.getElementById(`canvas-${id}`);
    if (!canvas) return null;

    const ctx = canvas.getContext('2d');

    // Gradient
    const gradient = ctx.createLinearGradient(0, 0, 0, 80);
    gradient.addColorStop(0, 'rgba(59, 130, 246, 0.4)');
    gradient.addColorStop(1, 'rgba(59, 130, 246, 0.0)');

    chartInstances[id] = new Chart(ctx, {
        type: 'line',
        data: {
            datasets: [{
                label: 'Latency (ms)',
                data: [], // populated async
                borderColor: '#3b82f6',
                backgroundColor: gradient,
                borderWidth: 2,
                pointRadius: 0,
                pointHoverRadius: 4,
                fill: true,
                tension: 0.3
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    mode: 'index',
                    intersect: false,
                    backgroundColor: 'rgba(15, 23, 42, 0.9)',
                    titleColor: '#94a3b8',
                    bodyColor: '#f8fafc',
                    borderColor: 'rgba(255,255,255,0.1)',
                    borderWidth: 1,
                    displayColors: false,
                    callbacks: {
                        title: (items) => {
                            if (!items.length) return '';
                            const d = new Date(items[0].raw.x);
                            return d.toLocaleTimeString();
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        unit: 'minute',
                        displayFormats: {
                            minute: 'HH:mm'
                        }
                    },
                    grid: { display: false },
                    ticks: { color: '#64748b' }
                },
                y: {
                    beginAtZero: true,
                    grid: { color: 'rgba(255,255,255,0.05)' },
                    ticks: { color: '#64748b' }
                }
            },
            interaction: {
                mode: 'nearest',
                axis: 'x',
                intersect: false
            }
        }
    });

    // Initial load
    updateChartHistory(id);

    return chartInstances[id];
}

function handleTestResult(result) {
    // Check if result belongs to current view
    if (!result.endpoint_id.startsWith(currentRegion + "-")) return;

    const id = result.endpoint_id;

    // Update DOM
    const latSpan = document.getElementById(`latency-${id}`);
    const dot = document.getElementById(`status-dot-${id}`);
    const updated = document.getElementById(`last-updated`);

    if (latSpan) latSpan.innerText = result.latency_ms;
    if (dot) {
        dot.className = "status-dot " + (result.status === "success" ? "success" : "failure");
    }
    if (updated) {
        updated.innerText = new Date().toLocaleTimeString();
    }

    // Update Chart Live
    const chart = chartInstances[id] || initChart(id);
    if (chart) {
        chart.data.datasets[0].data.push({
            x: new Date(result.timestamp),
            y: result.latency_ms
        });

        // Pruning? depends on range?
        // If range is "24h", we keep lots.
        // Let's rely on Chart.js performance or simple limit
        if (chart.data.datasets[0].data.length > 2000) {
            chart.data.datasets[0].data.shift();
        }

        chart.update('none'); // efficient update
    }
}

function setupSettings() {
    const modal = document.getElementById("settings-modal");

    // Open
    document.getElementById("btn-settings").addEventListener("click", () => {
        // Populate fields
        if (currentConfig && currentConfig.settings) {
            document.getElementById("setting-interval").value = currentConfig.settings.test_interval_seconds;
            document.getElementById("setting-retention").value = currentConfig.settings.data_retention_days;
            document.getElementById("setting-notifications").checked = currentConfig.settings.notifications_enabled;
        }
        modal.classList.add("active");
    });

    // Close
    document.getElementById("btn-close-settings").addEventListener("click", () => {
        modal.classList.remove("active");
    });

    // Save
    document.getElementById("settings-form").addEventListener("submit", async (e) => {
        e.preventDefault();

        const newSettings = {
            test_interval_seconds: parseInt(document.getElementById("setting-interval").value),
            data_retention_days: parseInt(document.getElementById("setting-retention").value),
            notifications_enabled: document.getElementById("setting-notifications").checked
        };
        
        // Update local object deeply
        currentConfig.settings = newSettings;

        // Send to backend
        try {
            const err = await window.go.main.App.SaveConfig(currentConfig);
            if (err) {
                alert("Error saving config: " + err);
            } else {
                modal.classList.remove("active");
                document.getElementById("status-message").innerText = "Settings Saved";
            }
        } catch (error) {
            console.error(error);
            alert("Failed to save settings: " + error);
        }
    });

    // Close on click outside
    modal.addEventListener("click", (e) => {
        if (e.target === modal) modal.classList.remove("active");
    });
}
