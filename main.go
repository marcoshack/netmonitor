package main

import (
	"context"
	"embed"

	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/marcoshack/netmonitor/internal/logging"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize zerolog
	logger := logging.Initialize()

	// Create context with logger
	ctx := logger.WithContext(context.Background())

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "NetMonitor",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(appCtx context.Context) {
			// Pass the logger context to startup
			app.startup(ctx)
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to start application")
	}
}
