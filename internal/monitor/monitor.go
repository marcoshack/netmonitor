package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/marcoshack/netmonitor/internal/models"
)

type Monitor struct {
	Config      *models.Configuration
	StopChan    chan struct{}
	ResultsChan chan models.TestResult
	IsRunning   bool
	mu          sync.Mutex
}

func NewMonitor(cfg *models.Configuration) *Monitor {
	return &Monitor{
		Config:      cfg,
		StopChan:    make(chan struct{}),
		ResultsChan: make(chan models.TestResult, 100),
	}
}

func (m *Monitor) Start() {
	m.mu.Lock()
	if m.IsRunning {
		m.mu.Unlock()
		return
	}
	m.IsRunning = true
	m.StopChan = make(chan struct{}) // Recreate in case it was closed
	m.mu.Unlock()

	go m.runLoop()
}

func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.IsRunning {
		return
	}
	close(m.StopChan)
	m.IsRunning = false
}

func (m *Monitor) runLoop() {
	ticker := time.NewTicker(time.Duration(m.Config.Settings.TestIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	m.RunAllTests()

	for {
		select {
		case <-m.StopChan:
			return
		case <-ticker.C:
			m.RunAllTests()
		}
	}
}

func (m *Monitor) RunAllTests() {
	var wg sync.WaitGroup

	for regionName, region := range m.Config.Regions {
		for _, endpoint := range region.Endpoints {
			wg.Add(1)
			go func(rName string, ep models.Endpoint) {
				defer wg.Done()
				result := m.TestEndpoint(ep)
				result.EndpointID = fmt.Sprintf("%s-%s", rName, ep.Name) // Simple ID generation
				m.ResultsChan <- result
			}(regionName, endpoint)
		}
	}

	wg.Wait()
}

func (m *Monitor) TestEndpoint(ep models.Endpoint) models.TestResult {
	start := time.Now()
	var err error
	var status string

	timeout := time.Duration(ep.Timeout) * time.Millisecond

	switch ep.Type {
	case models.TypeHTTP:
		err = checkHTTP(ep.Address, timeout)
	case models.TypeTCP:
		err = checkTCP(ep.Address, timeout)
	case models.TypeUDP:
		err = checkUDP(ep.Address, timeout)
	case models.TypeICMP:
		err = checkICMP(ep.Address, timeout)
	default:
		err = fmt.Errorf("unknown endpoint type: %s", ep.Type)
	}

	duration := time.Since(start).Milliseconds()
	if err != nil {
		status = "failure"
		// If it failed immediately, ensure duration reflects that or is recorded
	} else {
		status = "success"
	}

	return models.TestResult{
		Timestamp: time.Now(),
		LatencyMs: duration,
		Status:    status,
		Error:     errStr(err),
	}
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func checkHTTP(url string, timeout time.Duration) error {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http status %d", resp.StatusCode)
	}
	return nil
}

func checkTCP(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func checkUDP(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Attempt to write a byte to verify socket
	_, err = conn.Write([]byte{0})
	return err
}

func checkICMP(address string, timeout time.Duration) error {
	var cmd *exec.Cmd

	// Normalize address logic if needed (remove http:// if user put it by mistake for ICMP, etc)
	// For now assume address is IP or Hostname

	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", fmt.Sprintf("%d", timeout.Milliseconds()), address)
	} else {
		// Linux/Mac
		// -c 1 count, -W 1 wait in seconds (some ping versions take float, some int. safely use 1s or calculate)
		// MacOS ping -W is in milliseconds? No, seconds usually. Linux iputils-ping -W is seconds.
		// To be safe with timeout, we just use Context with timeout for the command execution,
		// but ping flags are better for actual latency constraints.
		// However, accurately mapping milliseconds to Ping flags across distros is hard.
		// Let's rely on Context timeout for defining "failure" and parse output if needed?
		// Or just use basic ping.
		cmd = exec.Command("ping", "-c", "1", address)
	}

	// Use context to enforce our timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// We need to run command with context
	// Go 1.20+ has exec.CommandContext
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", fmt.Sprintf("%d", timeout.Milliseconds()), address)
	} else {
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", address)
	}

	// Run
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
