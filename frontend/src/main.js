import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns';

let currentConfig = null;
let currentRegion = null;
let testResults = {}; // EndpointId -> Array of full TestResult objects (using new keys: ts, id, ms, st)
let chartInstances = {}; // id -> Chart instance
let endpointMap = {}; // HashID -> { ...endpoint, regionName }

// Detail View State
let detailChartInstance = null;
let currentDetailId = null;

// Edit Mode State
let isEditMode = false;
let editingEndpointId = null;
let originalEndpoint = null; // {address, type} for detecting changes

// Wails Runtime and Backend variables (injected by Wails)
// We assume window.go.main.App is available

window.addEventListener('DOMContentLoaded', init);

async function init() {
    try {
        // Load Config
        currentConfig = await window.go.main.App.GetConfig();

        // Initialize Endpoints Map (Generate IDs)
        await setupEndpoints();

        // Setup Regions
        setupRegionSelector();

        // Setup Event Listeners
        window.runtime.EventsOn("test-result", handleTestResult);

        setupSettings();
        setupAddMonitor();
        setupDetailsModal();
        setupWindowListeners();
        setupLogsButton();

        // Initial Layout
        renderDashboard();

        // Update status
        document.getElementById("status-message").innerText = "Monitoring Active";

    } catch (err) {
        console.error("Initialization failed", err);
        document.getElementById("status-message").innerText = "Error: " + err;
    }
}

async function setupEndpoints() {
    endpointMap = {};
    if (!currentConfig || !currentConfig.regions) return;

    for (const [regionName, region] of Object.entries(currentConfig.regions)) {
        for (const ep of region.endpoints) {
            const id = await window.go.main.App.GenerateEndpointID(ep.address, ep.type);
            ep._id = id; // Store ID on config object for ordering
            endpointMap[id] = { ...ep, regionName: regionName, id: id };
        }
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
    // Destroy existing chart instances before wiping DOM
    Object.keys(chartInstances).forEach(id => {
        if (chartInstances[id]) {
            chartInstances[id].destroy();
        }
    });
    chartInstances = {};

    const grid = document.getElementById("endpoints-grid");
    grid.innerHTML = ""; // Clear existing

    if (!currentRegion) return;

    // Enable Drop on Grid
    grid.ondragover = (e) => {
        e.preventDefault();
        const afterElement = getDragAfterElement(grid, e.clientY);
        const draggable = document.querySelector('.dragging');
        if (afterElement == null) {
            grid.appendChild(draggable);
        } else {
            grid.insertBefore(draggable, afterElement);
        }
    };

    grid.ondrop = (e) => {
        e.preventDefault();
        handleDrop();
    };

    // Render in Config Order
    const regionData = currentConfig.regions[currentRegion];
    if (regionData && regionData.endpoints) {
        regionData.endpoints.forEach(ep => {
            // Use the _id we attached in setupEndpoints to find the full data in map
            if (ep._id && endpointMap[ep._id]) {
                const card = createEndpointCard(endpointMap[ep._id]);
                grid.appendChild(card);
            }
        });
    }
}

function createEndpointCard(ep) {
    const id = ep.id; // Correct Hash ID

    const div = document.createElement("div");
    div.className = "card";
    div.id = `card-${id}`;

    // Add click event to open details
    div.onclick = (e) => {
        // Prevent click if clicking on chart specifically if we wanted, 
        // but user asked "click in any endpoint", so clicking anywhere on card is good.
        openDetailView(id);
    };
    div.style.cursor = "pointer";

    // Drag and Drop
    div.draggable = true;
    div.ondragstart = (e) => {
        div.classList.add('dragging');
        // e.dataTransfer.setData('text/plain', id); // We use class for selection
    };
    div.ondragend = (e) => {
        div.classList.remove('dragging');
    };

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

function getDragAfterElement(container, y) {
    const draggableElements = [...container.querySelectorAll('.card:not(.dragging)')];

    return draggableElements.reduce((closest, child) => {
        const box = child.getBoundingClientRect();
        const offset = y - box.top - box.height / 2;
        if (offset < 0 && offset > closest.offset) {
            return { offset: offset, element: child };
        } else {
            return closest;
        }
    }, { offset: Number.NEGATIVE_INFINITY }).element;
}

async function handleDrop() {
    // Get new order from DOM
    const grid = document.getElementById("endpoints-grid");
    const cards = [...grid.children];
    const newOrderIDs = cards.map(c => c.id.replace("card-", ""));

    // Update Local Config Memory
    const region = currentConfig.regions[currentRegion];
    if (region && region.endpoints) {
        // Sort endpoints array based on newOrderIDs
        // Map ID to endpoint object first
        const epMap = {};
        region.endpoints.forEach(ep => {
            if (ep._id) epMap[ep._id] = ep;
        });

        const newEndpoints = [];
        newOrderIDs.forEach(id => {
            if (epMap[id]) newEndpoints.push(epMap[id]);
        });

        // Append leftovers?
        region.endpoints.forEach(ep => {
            if (ep._id && !newEndpoints.includes(ep)) {
                newEndpoints.push(ep);
            }
        });

        region.endpoints = newEndpoints;
        currentConfig.regions[currentRegion] = region;
    }

    // Call Backend
    try {
        const err = await window.go.main.App.ReorderEndpoints(currentRegion, newOrderIDs);
        if (err) {
            console.error("Failed to reorder:", err);
            // Revert changes or show error?
            // For now just log. Local state is already updated so UI looks right.
            // If backend failed, next reload will revert.
        } else {
            console.log("Reorder saved");
        }
    } catch (e) {
        console.error(e);
    }
}

async function fetchHistory(range) {
    // Clear existing results or keep them? 
    // Usually fetching range means "reload all data for this range".
    testResults = {};

    try {
        const results = await window.go.main.App.GetHistoryRange(range);

        // Process results into map
        results.forEach(r => {
            const id = r.id; // Correct Hash ID
            if (!testResults[id]) testResults[id] = [];

            // Standardize timestamp
            r.timestamp = new Date(r.ts);
            r.latency_ms = r.ms; // Map for compatibility with charts
            r.statusStr = r.st === 0 ? "success" : "failure";

            testResults[id].push(r);
        });

        // Refresh all charts & details if open
        Object.keys(chartInstances).forEach(id => {
            updateChartHistory(id);
        });

        // Initialize any missing charts (only for visible cards)
        const visibleCards = document.querySelectorAll('.card');
        visibleCards.forEach(card => {
            const id = card.id.replace("card-", "");
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
    // Normalize data
    result.timestamp = new Date(result.ts);
    result.latency_ms = result.ms;
    result.statusStr = result.st === 0 ? "success" : "failure";

    const id = result.id; // Hash ID

    // Add to store
    if (!testResults[id]) testResults[id] = [];
    testResults[id].push(result);

    // Trim store if too big (optional, maybe run every N updates)
    if (testResults[id].length > 2000) {
        testResults[id].shift();
    }

    // Is it visible on current dashboard?
    // Check if the endpoint belongs to current region using endpointMap
    const ep = endpointMap[id];
    if (ep && ep.regionName === currentRegion) {
        // Update Card UI
        const latSpan = document.getElementById(`latency-${id}`);
        const dot = document.getElementById(`status-dot-${id}`);
        // removed updated timestamp per request previously? or just kept it. Kept it. 
        // But element id was `last-updated` which is unique? 
        // Ah, `last-updated` seems global or duplicated?
        // In createEndpointCard I don't see `last-updated`. 
        // Wait, look at previous code: `updated = document.getElementById("last-updated");`
        // If there are multiple cards, `last-updated` ID would be duplicate if inside card?
        // In createEndpointCard, I don't see `last-updated` ID being created.
        // It might be a header element (global status).
        // Let's keep it if it exists.

        if (latSpan) latSpan.innerText = result.latency_ms;
        if (dot) {
            dot.className = "status-dot " + (result.statusStr === "success" ? "success" : "failure");
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
    document.getElementById("btn-edit-details").onclick = () => {
        if (currentDetailId && endpointMap[currentDetailId]) {
            openEditMonitor(endpointMap[currentDetailId]);
        }
    };
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
    // 1. Find Endpoint Config form map
    const endpoint = endpointMap[id];
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

        if (last.statusStr === "success") {
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
        if (r.statusStr === "success") {
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

function setupWindowListeners() {
    let resizeTimeout;
    window.addEventListener('resize', () => {
        clearTimeout(resizeTimeout);
        resizeTimeout = setTimeout(() => {
            if (window.go && window.go.main && window.go.main.App) {
                window.go.main.App.WindowResized();
            }
        }, 500);
    });
}

function setupLogsButton() {
    const btn = document.getElementById("btn-logs");
    if (btn) {
        btn.addEventListener("click", () => {
            if (window.go && window.go.main && window.go.main.App) {
                window.go.main.App.OpenLogDirectory();
            }
        });
    }
}

function setupAddMonitor() {
    const modal = document.getElementById("add-monitor-modal");

    // Open
    const btn = document.getElementById("btn-add-monitor");
    if (btn) {
        btn.addEventListener("click", () => {
            isEditMode = false;
            editingEndpointId = null;
            originalEndpoint = null;
            document.querySelector("#add-monitor-modal h2").innerText = "Add Monitor";
            document.querySelector("#add-monitor-form button[type='submit']").innerText = "Add Monitor";

            // Reset form and enable fields
            const form = document.getElementById("add-monitor-form");
            form.reset();
            document.getElementById("add-address").disabled = false;
            document.getElementById("add-type").disabled = false;
            document.getElementById("add-address").title = "";
            document.getElementById("add-type").title = "";

            // Hide warning
            document.getElementById("edit-warning").style.display = "none";

            modal.classList.add("active");
        });
    }

    // Close
    document.getElementById("btn-close-add-monitor").addEventListener("click", () => {
        modal.classList.remove("active");
    });

    // Close on background click
    modal.addEventListener("click", (e) => {
        if (e.target === modal) modal.classList.remove("active");
    });

    // Submit
    document.getElementById("add-monitor-form").addEventListener("submit", async (e) => {
        e.preventDefault();

        const name = document.getElementById("add-name").value;
        const type = document.getElementById("add-type").value;
        const address = document.getElementById("add-address").value;
        const timeout = parseInt(document.getElementById("add-timeout").value);

        const newEndpoint = {
            name: name,
            type: type, // Ensure exact casing matches struct if needed, but select values are uppercase
            address: address,
            timeout: timeout
        };

        try {
            let err;
            if (isEditMode) {
                // Pass old identity to find the correct record to update
                err = await window.go.main.App.UpdateEndpoint(
                    originalEndpoint.address,
                    originalEndpoint.type,
                    newEndpoint
                );
            } else {
                err = await window.go.main.App.AddEndpoint(newEndpoint);
            }

            if (err) {
                alert("Error: " + err);
            } else {
                modal.classList.remove("active");
                document.getElementById("add-monitor-form").reset();

                // Refresh App State
                currentConfig = await window.go.main.App.GetConfig();
                await setupEndpoints();
                // Ensure we are on the region where we added it? 
                // Currently only Default region exists in UI logic mostly, or it adds to Default.
                // If currentRegion is not Default, user might not see it. 
                // But AddEndpoint adds to "Default".
                if (currentRegion !== "Default" && Object.keys(currentConfig.regions).includes("Default")) {
                    // Optionally switch to Default or let user switch. 
                    // Let's just render.
                }
                renderDashboard();

                // Re-fetch and re-init charts
                const range = document.getElementById("time-range-select").value;
                await fetchHistory(range);

                document.getElementById("status-message").innerText = isEditMode ? "Monitor Updated" : "Monitor Added";

                // If editing, also update details view if it's the one open
                if (isEditMode && currentDetailId === editingEndpointId) {
                    // Start: Fix for ID change
                    const newId = await window.go.main.App.GenerateEndpointID(newEndpoint.address, newEndpoint.type);
                    if (newId !== currentDetailId) {
                        // ID changed (address/protocol changed)
                        currentDetailId = newId;
                        // Re-open/Refresh details with new ID
                        // We need to ensure graph exists or init it
                        updateDetailView(newId);

                        // Destroy old chart instance if key changed? 
                        // Actually renderDashboard destroyed all charts and created new cards.
                        // But detailChartInstance is global single instance for the modal?
                        // Let's check initDetailChart logic. It uses "detail-canvas". 
                        // It doesn't seem to depend on ID for the instance itself, just data.
                        renderDetailChart(newId);
                    } else {
                        updateDetailView(currentDetailId);
                    }
                    // End: Fix for ID change
                }
            }
        } catch (error) {
            console.error(error);
            alert("System error: " + error);
        }
    });
}
function openEditMonitor(endpoint) {
    const modal = document.getElementById("add-monitor-modal");
    isEditMode = true;
    editingEndpointId = endpoint.id;
    originalEndpoint = {
        address: endpoint.address,
        type: endpoint.type
    };

    document.querySelector("#add-monitor-modal h2").innerText = "Edit Monitor";
    document.querySelector("#add-monitor-form button[type='submit']").innerText = "Save Changes";

    // Populate
    document.getElementById("add-name").value = endpoint.name;
    document.getElementById("add-type").value = endpoint.type;
    document.getElementById("add-address").value = endpoint.address;
    document.getElementById("add-timeout").value = endpoint.timeout;

    // Enable identity fields (now allowed)
    const addrInput = document.getElementById("add-address");
    const typeInput = document.getElementById("add-type");

    addrInput.disabled = false;
    typeInput.disabled = false;
    addrInput.title = "";
    typeInput.title = "";

    // Setup warning listeners
    const checkChanges = () => {
        const newAddr = addrInput.value;
        const newType = typeInput.value;
        const changed = (newAddr !== originalEndpoint.address) || (newType !== originalEndpoint.type);
        document.getElementById("edit-warning").style.display = changed ? "block" : "none";
    };

    addrInput.oninput = checkChanges;
    typeInput.onchange = checkChanges;

    // Initial check (should be hidden)
    document.getElementById("edit-warning").style.display = "none";

    modal.classList.add("active");
}

