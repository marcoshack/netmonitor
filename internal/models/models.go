package models

// EndpointType defines the protocol used for testing
type EndpointType string

const (
	TypeHTTP EndpointType = "HTTP"
	TypeTCP  EndpointType = "TCP"
	TypeUDP  EndpointType = "UDP"
	TypeICMP EndpointType = "ICMP"
)

// Endpoint represents a single network target to monitor
type Endpoint struct {
	Name    string       `json:"name"`
	Type    EndpointType `json:"type"`
	Address string       `json:"address"`
	Timeout int          `json:"timeout"` // Timeout in milliseconds
}

// Thresholds defines when to trigger alerts for a region
type Thresholds struct {
	LatencyMs           int     `json:"latency_ms"`
	AvailabilityPercent float64 `json:"availability_percent"`
}

// Region groups endpoints geographically or logically
type Region struct {
	Endpoints  []Endpoint `json:"endpoints"`
	Thresholds Thresholds `json:"thresholds"`
}

// TestResult captures the outcome of a single endpoint test
type TestResult struct {
	Ts  int64  `json:"ts"`
	Id  string `json:"id"`
	Ms  int64  `json:"ms"`
	St  int    `json:"st"` // 0=success, 1=timeout, 2=error
	Err error  `json:"err"`
}

// AppSettings defines global application settings
type AppSettings struct {
	TestIntervalSeconds  int  `json:"test_interval_seconds"`
	DataRetentionDays    int  `json:"data_retention_days"`
	NotificationsEnabled bool `json:"notifications_enabled"`
	WindowWidth          int  `json:"window_width,omitempty"`
	WindowHeight         int  `json:"window_height,omitempty"`
	WindowX              int  `json:"window_x,omitempty"`
	WindowY              int  `json:"window_y,omitempty"`
}

// Configuration represents the entire application config structure
type Configuration struct {
	Regions  map[string]Region `json:"regions"`
	Settings AppSettings       `json:"settings"`
}
