//go:build darwin

package startup

import (
	"fmt"
	"os"
	"path/filepath"
)

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.marcoshack.netmonitor</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`

func get() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	path := filepath.Join(home, "Library", "LaunchAgents", "com.marcoshack.netmonitor.plist")
	_, err = os.Stat(path)
	return err == nil
}

func set(enabled bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, "Library", "LaunchAgents")
	path := filepath.Join(dir, "com.marcoshack.netmonitor.plist")

	if enabled {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		exe, err := os.Executable()
		if err != nil {
			return err
		}

		content := fmt.Sprintf(plistTemplate, exe)
		return os.WriteFile(path, []byte(content), 0644)
	}

	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
