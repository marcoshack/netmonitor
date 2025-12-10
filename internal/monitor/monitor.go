package monitor

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/marcoshack/netmonitor/internal/models"
	probing "github.com/prometheus-community/pro-bing"
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
	ticker := time.NewTicker(time.Duration(m.Config.Settings.TestIntervalSeconds) * time.Second)
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
	pinger, err := probing.NewPinger(address)
	if err != nil {
		return err
	}

	pinger.Count = 1
	pinger.Timeout = timeout

	// On Windows, this triggers the use of the IcmpSendEcho API which works for unprivileged users.
	// On Linux, it attempts raw sockets (requires root) unless configured otherwise.
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	err = pinger.Run()
	if err != nil {
		return err
	}

	if pinger.Statistics().PacketsRecv == 0 {
		return fmt.Errorf("packet loss")
	}

	return nil
}
