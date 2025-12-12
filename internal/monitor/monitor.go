package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/marcoshack/netmonitor/internal/models"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/rs/zerolog/log"
)

type Monitor struct {
	Ctx         context.Context
	Config      *models.Configuration
	StopChan    chan struct{}
	ResultsChan chan models.TestResult
	IsRunning   bool
	mu          sync.Mutex
}

func NewMonitor(ctx context.Context, cfg *models.Configuration) *Monitor {
	return &Monitor{
		Ctx:         ctx,
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

	log.Ctx(m.Ctx).Info().Msg("Monitor started")
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
	log.Ctx(m.Ctx).Info().Msg("Monitor stopped")
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
				// ID is already generated in TestEndpoint based on address/protocol
				// If we needed region in hash, we'd pass it. User said Address + Protocol.
				m.ResultsChan <- result
			}(regionName, endpoint)
		}
	}

	wg.Wait()
}

const (
	ResultSuccess = 0
	ResultTimeout = 1
	ResultError   = 2
)

func (m *Monitor) TestEndpoint(ep models.Endpoint) models.TestResult {
	var err error
	var status int
	var durationMs int64

	timeout := time.Duration(ep.Timeout) * time.Millisecond
	var d time.Duration

	switch ep.Type {
	case models.TypeHTTP:
		d, err = checkHTTP(ep.Address, timeout)
	case models.TypeTCP:
		d, err = checkTCP(ep.Address, timeout)
	case models.TypeUDP:
		d, err = checkUDP(ep.Address, timeout)
	case models.TypeICMP:
		d, err = checkICMP(ep.Address, timeout)
	default:
		err = fmt.Errorf("unknown endpoint type: %s", ep.Type)
	}

	durationMs = d.Milliseconds()
	if err != nil {
		if d >= timeout {
			status = ResultTimeout
		} else {
			status = ResultError
		}
	} else {
		status = ResultSuccess
	}

	// Generate ID: last 7 chars of UUID SHA1(NameSpaceURL, Address + Protocol)
	// Note: Providing Protocol (Type) and Address ensures uniqueness for same address different protocols.
	idData := ep.Address + string(ep.Type)
	shortId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(idData)).String()[:7]

	log.Ctx(m.Ctx).Debug().
		Str("id", shortId).
		Str("address", ep.Address).
		Str("type", string(ep.Type)).
		Int64("latency_ms", durationMs).
		Int("status", status).
		Msg("Endpoint tested")

	return models.TestResult{
		Ts: time.Now().Unix(),
		Id: shortId,
		Ms: durationMs,
		St: status,
	}
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func checkHTTP(url string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return time.Since(start), err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return time.Since(start), fmt.Errorf("http status %d", resp.StatusCode)
	}
	return time.Since(start), nil
}

func checkTCP(address string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return time.Since(start), err
	}
	conn.Close()
	return time.Since(start), nil
}

func checkUDP(address string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		return time.Since(start), err
	}
	defer conn.Close()

	// Attempt to write a byte to verify socket
	_, err = conn.Write([]byte{0})
	return time.Since(start), err
}

func checkICMP(address string, timeout time.Duration) (time.Duration, error) {
	pinger, err := probing.NewPinger(address)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return 0, fmt.Errorf("packet loss")
	}

	return stats.AvgRtt, nil
}
