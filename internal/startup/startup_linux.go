//go:build linux

package startup

import (
	"fmt"
	"os"
	"path/filepath"
)

func get() bool {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return false
	}
	path := filepath.Join(configDir, "autostart", "netmonitor.desktop")
	_, err = os.Stat(path)
	return err == nil
}

func set(enabled bool) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	autostartDir := filepath.Join(configDir, "autostart")
	path := filepath.Join(autostartDir, "netmonitor.desktop")

	if enabled {
		if err := os.MkdirAll(autostartDir, 0755); err != nil {
			return err
		}

		exe, err := os.Executable()
		if err != nil {
			return err
		}

		content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=NetMonitor
Exec=%s
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
Comment=Network Monitoring Tool
`, exe)

		return os.WriteFile(path, []byte(content), 0644)
	}

	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
