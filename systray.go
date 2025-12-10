package main

import (
	_ "embed"
	"log"
	"os"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/trayicon.ico
var iconData []byte

// onReady is called when the system tray is ready
func (a *App) onReady() {
	// Set icon with error checking
	if len(iconData) == 0 {
		log.Println("Warning: icon data is empty")
	} else {
		log.Printf("Setting tray icon, size: %d bytes\n", len(iconData))
		systray.SetIcon(iconData)
	}

	systray.SetTitle("NetMonitor")
	systray.SetTooltip("NetMonitor - Network Monitoring Tool")

	// Add menu items
	mShow := systray.AddMenuItem("Show App", "Show the application window")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Exit", "Quit the application")

	// Handle menu actions in a goroutine
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				a.ShowWindow()
			case <-mQuit.ClickedCh:
				log.Println("Exit menu clicked, quitting...")
				// Quit systray first - this will trigger onExit
				systray.Quit()
				return
			}
		}
	}()
}

// onExit is called when the system tray is exiting
func (a *App) onExit() {
	log.Println("System tray exiting, quitting Wails app...")
	// Quit the Wails application
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
	// Force exit if Wails doesn't quit properly
	os.Exit(0)
}

// InitSystemTray initializes the system tray
func (a *App) InitSystemTray() {
	systray.Run(a.onReady, a.onExit)
}

// ShowWindow shows the application window
func (a *App) ShowWindow() {
	if a.ctx != nil {
		// Show and unminimize the window
		runtime.WindowShow(a.ctx)
		runtime.WindowUnminimise(a.ctx)
		// Bring window to front
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
	}
}

// HideWindow hides the application window
func (a *App) HideWindow() {
	if a.ctx != nil {
		runtime.WindowHide(a.ctx)
	}
}

// QuitApp properly shuts down the application
func (a *App) QuitApp() {
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}
