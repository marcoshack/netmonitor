//go:build windows

package startup

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

func get() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue("NetMonitor")
	if err != nil {
		return false
	}

	exe, err := os.Executable()
	if err != nil {
		return false
	}

	return val == exe
}

func set(enabled bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if enabled {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		return k.SetStringValue("NetMonitor", exe)
	}

	err = k.DeleteValue("NetMonitor")
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}
