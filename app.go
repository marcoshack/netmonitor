package main

import (
	"context"
	"time"

	"github.com/marcoshack/netmonitor/internal/config"
	"github.com/marcoshack/netmonitor/internal/data"
	"github.com/marcoshack/netmonitor/internal/models"
	"github.com/marcoshack/netmonitor/internal/monitor"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	Config  *models.Configuration
	Monitor *monitor.Monitor
	Storage *data.Storage
	// Paths
	ConfigPath string
	DataDir    string
}

// NewApp creates a new App application struct
func NewApp() *App {
	configPath := "config.json"
	dataDir := "data"

	// Ensure absolute paths in real app, but relative is fine for portable desktop app often.
	// Wails runs from build dir or current dir.

	cfg, _ := config.LoadConfig(configPath)
	// We ignore error here because LoadConfig returns default if fail, or error if completely broken.
	// Ideally we handle it.

	store := data.NewStorage(dataDir)
	mon := monitor.NewMonitor(cfg)

	return &App{
		Config:     cfg,
		Monitor:    mon,
		Storage:    store,
		ConfigPath: configPath,
		DataDir:    dataDir,
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize system tray
	go a.InitSystemTray()

	// Start Monitor
	// Relay results to frontend
	go func() {
		for res := range a.Monitor.ResultsChan {
			// Save to storage
			_ = a.Storage.SaveResult(res)
			// Emit event to frontend
			runtime.EventsEmit(a.ctx, "test-result", res)
		}
	}()

	a.Monitor.Start()
}

// DomReady is called after the front-end is created.
func (a *App) DomReady(ctx context.Context) {
	// Restore Window Position if set
	if a.Config.Settings.WindowX != -1 && a.Config.Settings.WindowY != -1 {
		// Safety check for multi-monitor:
		// If the user disconnected a monitor, coordinates might be off-screen.
		// We use a heuristic: If only 1 screen is detected, we enforce positive coordinates.

		screens, err := runtime.ScreenGetAll(a.ctx)
		if err == nil && len(screens) == 1 {
			// Single monitor setup
			if a.Config.Settings.WindowX < 0 || a.Config.Settings.WindowY < 0 {
				// Window was likely on a secondary monitor (top/left) that is gone.
				runtime.WindowCenter(a.ctx)
				return
			}

			screen := screens[0]
			// Check if out of bounds to the right/bottom
			if a.Config.Settings.WindowX > screen.Size.Width || a.Config.Settings.WindowY > screen.Size.Height {
				runtime.WindowCenter(a.ctx)
				return
			}
		}

		runtime.WindowSetPosition(a.ctx, a.Config.Settings.WindowX, a.Config.Settings.WindowY)
	}
}

// Shutdown is called at termination
func (a *App) Shutdown(ctx context.Context) {
	if a.Monitor != nil {
		a.Monitor.Stop()
	}
}

// Backend Methods exposed to Frontend

func (a *App) GetConfig() models.Configuration {
	return *a.Config
}

func (a *App) SaveConfig(cfg models.Configuration) string {
	a.Config = &cfg         // Update in memory
	a.Monitor.Config = &cfg // Update monitor config reference (simple pointer update)
	// In robust app, better to use setter on monitor to restart ticker if interval changed
	// or protect with mutex. For MVP this is acceptable if careful.
	// Ideally Monitor handles config updates.

	err := config.SaveConfig(a.ConfigPath, a.Config)
	if err != nil {
		return err.Error()
	}

	// Restart monitor to apply new settings (e.g. interval)
	a.Monitor.Stop()
	a.Monitor.Start()

	return ""
}

func (a *App) GetHistory(dateStr string) []models.TestResult {
	// dateStr expected "YYYY-MM-DD"
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// return empty or today
		t = time.Now()
	}
	res, _ := a.Storage.GetResultsForDay(t)
	return res
}

func (a *App) GetHistoryRange(durationStr string) []models.TestResult {
	// durationStr: "24h", "168h" (week), "720h" (month)
	// Or descriptive: "day", "week", "month"

	end := time.Now()
	var start time.Time

	switch durationStr {
	case "week":
		start = end.AddDate(0, 0, -7)
	case "month":
		start = end.AddDate(0, -1, 0)
	case "day":
		fallthrough
	default:
		start = end.Add(-24 * time.Hour)
	}

	res, _ := a.Storage.GetResultsForRange(start, end)
	return res
}

func (a *App) ManualTest(endpoint models.Endpoint) models.TestResult {
	return a.Monitor.TestEndpoint(endpoint)
}

func (a *App) GetRegions() map[string]models.Region {
	return a.Config.Regions
}

func (a *App) WindowResized() {
	if a.ctx == nil {
		return
	}
	width, height := runtime.WindowGetSize(a.ctx)
	x, y := runtime.WindowGetPosition(a.ctx)
	a.Config.Settings.WindowWidth = width
	a.Config.Settings.WindowHeight = height
	a.Config.Settings.WindowX = x
	a.Config.Settings.WindowY = y
	_ = config.SaveConfig(a.ConfigPath, a.Config)
}
