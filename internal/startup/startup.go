package startup

// Get returns true if the application is set to start on boot.
func Get() bool {
	return get()
}

// Set enables or disables start on boot.
func Set(enabled bool) error {
	return set(enabled)
}
