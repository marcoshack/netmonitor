package storage

import (
	"encoding/json"
	"time"
)

// DetailedTestResult extends TestResult with additional timing breakdown
type DetailedTestResult struct {
	TestResult                    // Embedded basic result
	ExecutionTime    time.Duration `json:"-"`                  // Total test execution time
	DNSLookupTime    time.Duration `json:"-"`                  // DNS resolution time (HTTP)
	ConnectionTime   time.Duration `json:"-"`                  // TCP connection establishment (TCP/HTTP)
	TLSHandshakeTime time.Duration `json:"-"`                  // TLS handshake time (HTTPS)
	FirstByteTime    time.Duration `json:"-"`                  // Time to first byte (HTTP)
	TransferTime     time.Duration `json:"-"`                  // Data transfer time (HTTP)
	IntermediateSteps []string     `json:"intermediateSteps"` // Step-by-step execution log
}

// MarshalJSON implements custom JSON marshaling for DetailedTestResult
func (dtr *DetailedTestResult) MarshalJSON() ([]byte, error) {
	type Alias DetailedTestResult
	return json.Marshal(&struct {
		*Alias
		LatencyInMs        float64 `json:"latencyInMs"`
		ExecutionTimeMs    float64 `json:"executionTimeMs"`
		DNSLookupTimeMs    float64 `json:"dnsLookupTimeMs"`
		ConnectionTimeMs   float64 `json:"connectionTimeMs"`
		TLSHandshakeTimeMs float64 `json:"tlsHandshakeTimeMs"`
		FirstByteTimeMs    float64 `json:"firstByteTimeMs"`
		TransferTimeMs     float64 `json:"transferTimeMs"`
	}{
		Alias:              (*Alias)(dtr),
		LatencyInMs:        float64(dtr.Latency.Nanoseconds()) / 1_000_000.0,
		ExecutionTimeMs:    float64(dtr.ExecutionTime.Nanoseconds()) / 1_000_000.0,
		DNSLookupTimeMs:    float64(dtr.DNSLookupTime.Nanoseconds()) / 1_000_000.0,
		ConnectionTimeMs:   float64(dtr.ConnectionTime.Nanoseconds()) / 1_000_000.0,
		TLSHandshakeTimeMs: float64(dtr.TLSHandshakeTime.Nanoseconds()) / 1_000_000.0,
		FirstByteTimeMs:    float64(dtr.FirstByteTime.Nanoseconds()) / 1_000_000.0,
		TransferTimeMs:     float64(dtr.TransferTime.Nanoseconds()) / 1_000_000.0,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for DetailedTestResult
func (dtr *DetailedTestResult) UnmarshalJSON(data []byte) error {
	type Alias DetailedTestResult
	aux := &struct {
		*Alias
		LatencyInMs        float64 `json:"latencyInMs"`
		ExecutionTimeMs    float64 `json:"executionTimeMs"`
		DNSLookupTimeMs    float64 `json:"dnsLookupTimeMs"`
		ConnectionTimeMs   float64 `json:"connectionTimeMs"`
		TLSHandshakeTimeMs float64 `json:"tlsHandshakeTimeMs"`
		FirstByteTimeMs    float64 `json:"firstByteTimeMs"`
		TransferTimeMs     float64 `json:"transferTimeMs"`
	}{
		Alias: (*Alias)(dtr),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	dtr.Latency = time.Duration(aux.LatencyInMs * 1_000_000.0)
	dtr.ExecutionTime = time.Duration(aux.ExecutionTimeMs * 1_000_000.0)
	dtr.DNSLookupTime = time.Duration(aux.DNSLookupTimeMs * 1_000_000.0)
	dtr.ConnectionTime = time.Duration(aux.ConnectionTimeMs * 1_000_000.0)
	dtr.TLSHandshakeTime = time.Duration(aux.TLSHandshakeTimeMs * 1_000_000.0)
	dtr.FirstByteTime = time.Duration(aux.FirstByteTimeMs * 1_000_000.0)
	dtr.TransferTime = time.Duration(aux.TransferTimeMs * 1_000_000.0)
	return nil
}
