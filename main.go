package main

import (
	"context"
	"embed"
	"flag"
	"os"
	"path/filepath"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/marcoshack/netmonitor/internal/logger"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Parse CLI flags
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Get User Config Directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		println("Error getting user config directory:", err.Error())
		configDir = "." // Fallback to current directory
	}
	appDir := filepath.Join(configDir, "NetMonitor")
	_ = os.MkdirAll(appDir, 0755)

	// Initialize Logger
	logDir := filepath.Join(appDir, "logs")
	l, closeLogger, err := logger.New(logDir, *debug)
	if err != nil {
		println("Error initializing logger:", err.Error())
		// Proceed without logger or exit? Proceeding might be safer for UX, but logging is broken.
		// For now, print error. The app might still run.
	}
	defer closeLogger()

	// Create context with logger
	ctx := l.WithContext(context.Background())

	// Create an instance of the app structure
	app := NewApp(ctx, appDir)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "netmonitor",
		Width:  app.Config.Settings.WindowWidth,
		Height: app.Config.Settings.WindowHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		OnDomReady:       app.DomReady,
		OnShutdown:       app.Shutdown,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			// Save state before hiding
			app.WindowResized()
			// Prevent window close and hide to tray instead
			app.HideWindow()
			return true
		},
		Bind: []interface{}{
			app,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "e345b678-9012-3499-7890-123456789012",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				app.ShowWindow()
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}

	// Clean up systray on exit
	systray.Quit()
}
