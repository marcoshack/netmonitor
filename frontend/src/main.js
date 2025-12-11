import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns';

let currentConfig = null;
let currentRegion = null;
let testResults = {}; // EndpointId -> Array of full TestResult objects
let chartInstances = {}; // id -> Chart instance

// Detail View State
let detailChartInstance = null;
let currentDetailId = null;

// Wails Runtime and Backend variables (injected by Wails)
// We assume window.go.main.App is available

window.addEventListener('DOMContentLoaded', init);

async function init() {
    try {
        // Check for Config Errors
        checkConfigErrors();

        // Load Config
        currentConfig = await window.go.main.App.GetConfig();

        // Setup Regions
        setupRegionSelector();

        // Setup Event Listeners
        window.runtime.EventsOn("test-result", handleTestResult);

        setupSettings();
        setupDetailsModal();
        setupNotificationsModal();

        // Initial Layout
        renderDashboard();

        // Update status
        document.getElementById("status-message").innerText = "Monitoring Active";

    } catch (err) {
        console.error("Initialization failed", err);
        document.getElementById("status-message").innerText = "Error: " + err;
    }
}

async function checkConfigErrors() {
    try {
        const notifications = await window.go.main.App.GetNotifications();
        if (notifications && notifications.length > 0) {
            showNotificationsModal(notifications);
        }
    } catch (err) {
        console.error("Failed to check config errors", err);
    }
}

function setupNotificationsModal() {
    const modal = document.getElementById("notifications-modal");
    document.getElementById("btn-close-notifications").onclick = () => {
        modal.classList.remove("active");
    };
    document.getElementById("btn-ignore-notifications").onclick = () => {
        modal.classList.remove("active");
    };

    document.getElementById("btn-fix-duplicates").onclick = async () => {
        try {
            const err = await window.go.main.App.RemoveDuplicateEndpoints();
            if (err) {
                alert("Error removing duplicates: " + err);
            } else {
                alert("Duplicates removed and configuration saved.");
                modal.classList.remove("active");
                // Reload config
                currentConfig = await window.go.main.App.GetConfig();
                setupRegionSelector();
                renderDashboard();
            }
        } catch (error) {
            console.error(error);
            alert("Failed to fix duplicates: " + error);
        }
    };
}

function showNotificationsModal(notifications) {
    const list = document.getElementById("notifications-list");
    list.innerHTML = "";
    notifications.forEach(n => {
        const li = document.createElement("li");
        li.innerText = n.message;
        list.appendChild(li);
    });
    document.getElementById("notifications-modal").classList.add("active");
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
    // Generate ID consistent with backend: Type:Address
    // Wait, createEndpointCard used to generate ID based on region and name.
    // The backend now sends results with ID = Type:Address.
    // BUT, the frontend needs to match the ID from the result to the card.
    // The previous frontend implementation used: `${currentRegion}-${ep.name}`

    // The backend change I made: `EndpointID = fmt.Sprintf("%s:%s", ep.Type, ep.Address)`

    // So here in frontend I MUST also update how I generate the ID for the card so that `handleTestResult` can find it.

    const id = `${ep.type}:${ep.address}`;

    const div = document.createElement("div");
    div.className = "card";
    // IDs can contain special chars like http:// or : so we should escape or use a safe selector?
    // CSS selectors with colons need escaping. `document.getElementById` is fine with colons.
    div.id = `card-${id}`;

    // Add click event to open details
    div.onclick = (e) => {
        openDetailView(id);
    };
    div.style.cursor = "pointer";

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

    return div;
}

async function fetchHistory(range) {
    testResults = {};

    try {
        const results = await window.go.main.App.GetHistoryRange(range);

        // Process results into map
        results.forEach(r => {
            const id = r.endpoint_id;
            if (!testResults[id]) testResults[id] = [];

            // Standardize timestamp
            r.timestamp = new Date(r.timestamp);
            testResults[id].push(r);
        });

        // Refresh all charts & details if open
        Object.keys(chartInstances).forEach(id => {
            updateChartHistory(id);
        });

        // Initialize any missing charts
        // We need to iterate over currently rendered cards
        const cards = document.querySelectorAll('.card');
        cards.forEach(card => {
            const id = card.id.replace("card-", ""); // This might be tricky if ID contains 'card-' itself, but unlikely.
            if (!chartInstances[id]) {
                initChart(id);
            } else {
                updateChartHistory(id);
            }
        });

        // If detail view is open, update it
        if (currentDetailId) {
            updateDetailView(currentDetailId);
        }

    } catch (err) {
        console.error("Failed to fetch history", err);
    }
}

function updateChartHistory(id) {
    const chart = chartInstances[id];
    if (!chart) return;

    const results = testResults[id] || [];

    // Sort
    results.sort((a, b) => a.timestamp - b.timestamp);

    // Map to {x, y}
    const data = results.map(r => ({
        x: r.timestamp,
        y: r.latency_ms
    }));

    chart.data.datasets[0].data = data;
    chart.update();
}

function initChart(id) {
    if (chartInstances[id]) return chartInstances[id];

    // IDs with special chars need to be handled carefully in getElementById? No, strings are fine.
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
                    displayColors: false,
                }
            },
            scales: {
                x: {
                    type: 'time',
                    time: { unit: 'minute', displayFormats: { minute: 'HH:mm' } },
                    grid: { display: false },
                    ticks: { display: false } // Hide ticks on small cards for cleaner look
                },
                y: {
                    beginAtZero: true,
                    grid: { display: false },
                    ticks: { display: false } // Hide ticks on small cards
                }
            },
            interaction: {
                mode: 'nearest',
                axis: 'x',
                intersect: false
            },
            events: [] // Disable hover events for performance on small cards? Or keep them.
            // Let's keep default events but maybe reduce hover overhead if needed.
        }
    });

    // Initial load
    updateChartHistory(id);

    return chartInstances[id];
}

function handleTestResult(result) {
    // result comes with timestamp as string probably if from JSON?
    // Wails JSON encoding might keep it as string.
    result.timestamp = new Date(result.timestamp);

    const id = result.endpoint_id;

    // Add to store
    if (!testResults[id]) testResults[id] = [];
    testResults[id].push(result);

    // Trim store if too big (optional, maybe run every N updates)
    if (testResults[id].length > 2000) {
        testResults[id].shift();
    }

    // Is it visible on current dashboard?
    // Previously checked `id.startsWith(currentRegion + "-")`.
    // Now ID is `Type:Address`.
    // We need to check if this endpoint belongs to the current region.
    // We can do this by finding the endpoint in config and comparing IDs.

    // However, finding it every time might be slow.
    // But since we only have one region active, we can check if the ID corresponds to any card currently in DOM.
    if (document.getElementById(`card-${id}`)) {
        // Update Card UI
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

        // Update Chart
        const chart = chartInstances[id] || initChart(id);
        if (chart) {
            chart.data.datasets[0].data.push({
                x: result.timestamp,
                y: result.latency_ms
            });
            // Simple trim for chart
            if (chart.data.datasets[0].data.length > 2000) {
                chart.data.datasets[0].data.shift();
            }
            chart.update('none');
        }
    }

    // Is it the currently detailed endpoint?
    if (currentDetailId && currentDetailId === id) {
        updateDetailView(id);
    }
}

// --- Details View Functions ---

function setupDetailsModal() {
    const modal = document.getElementById("details-modal");
    document.getElementById("btn-close-details").onclick = closeDetailView;
    modal.onclick = (e) => {
        if (e.target === modal) closeDetailView();
    };
}

function openDetailView(id) {
    currentDetailId = id;
    const modal = document.getElementById("details-modal");
    modal.classList.add("active");

    updateDetailView(id);
    initDetailChart(); // Create if not exists
    renderDetailChart(id);
}

function closeDetailView() {
    document.getElementById("details-modal").classList.remove("active");
    currentDetailId = null;
}

function updateDetailView(id) {
    // 1. Find Endpoint Config
    // ID is Type:Address

    // Search in current region endpoints
    const endpoint = currentConfig.regions[currentRegion].endpoints.find(e => `${e.type}:${e.address}` === id);

    if (!endpoint) return;

    // 2. Populate Info
    document.getElementById("detail-title").innerText = endpoint.name;
    document.getElementById("detail-address").innerText = endpoint.address;
    document.getElementById("detail-protocol").innerText = endpoint.type;

    // 3. Get Latest Data
    const results = testResults[id] || [];
    if (results.length > 0) {
        const last = results[results.length - 1];
        document.getElementById("detail-latency").innerText = last.latency_ms;

        const statusText = document.getElementById("detail-status-text");
        const dot = document.getElementById("detail-status-dot");

        if (last.status === "success") {
            statusText.innerText = "Operational";
            statusText.className = "text-success font-bold";
            dot.className = "status-dot success";
        } else {
            statusText.innerText = "Failure: " + (last.error || "Unknown");
            statusText.className = "text-error font-bold";
            dot.className = "status-dot failure";
        }
    } else {
        document.getElementById("detail-latency").innerText = "--";
        document.getElementById("detail-status-text").innerText = "No Data";
    }

    // 4. Populate Table (Last 10)
    const tbody = document.getElementById("detail-history-body");
    tbody.innerHTML = "";

    // Copy and reverse for table
    const last10 = [...results].sort((a, b) => b.timestamp - a.timestamp).slice(0, 10);

    last10.forEach(r => {
        const tr = document.createElement("tr");

        // Time
        const tdTime = document.createElement("td");
        tdTime.innerText = r.timestamp.toLocaleTimeString();
        tr.appendChild(tdTime);

        // Status
        const tdStatus = document.createElement("td");
        tdStatus.className = "text-center";
        const statusSpan = document.createElement("span");
        if (r.status === "success") {
            statusSpan.className = "text-success";
            statusSpan.innerText = "✔"; // Checkmark
        } else {
            statusSpan.className = "text-error";
            statusSpan.innerText = "✖"; // X
        }
        tdStatus.appendChild(statusSpan);
        tr.appendChild(tdStatus);

        // Latency
        const tdLat = document.createElement("td");
        tdLat.className = "text-right font-mono";
        tdLat.innerText = r.latency_ms + " ms";
        tr.appendChild(tdLat);

        tbody.appendChild(tr);
    });

    // 5. Update Chart (if already initialized)
    if (detailChartInstance) {
        renderDetailChart(id);
    }
}

function initDetailChart() {
    if (detailChartInstance) return;

    const ctx = document.getElementById("detail-canvas").getContext("2d");

    // Gradient
    const gradient = ctx.createLinearGradient(0, 0, 0, 300);
    gradient.addColorStop(0, 'rgba(59, 130, 246, 0.4)');
    gradient.addColorStop(1, 'rgba(59, 130, 246, 0.0)');

    detailChartInstance = new Chart(ctx, {
        type: 'line',
        data: {
            datasets: [{
                label: 'Latency',
                data: [],
                borderColor: '#3b82f6',
                backgroundColor: gradient,
                borderWidth: 2,
                pointRadius: 1, // Visible points
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
                    callbacks: {
                        title: (items) => {
                            if (!items.length) return '';
                            return new Date(items[0].raw.x).toLocaleString();
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'time',
                    time: { unit: 'minute', displayFormats: { minute: 'HH:mm:ss' } },
                    grid: { color: 'rgba(255,255,255,0.05)' },
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
}

function renderDetailChart(id) {
    if (!detailChartInstance) return;

    const results = testResults[id] || [];
    // Sort
    const sorted = [...results].sort((a, b) => a.timestamp - b.timestamp);

    const data = sorted.map(r => ({
        x: r.timestamp,
        y: r.latency_ms
    }));

    detailChartInstance.data.datasets[0].data = data;
    detailChartInstance.update();
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
