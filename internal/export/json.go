package export

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/marcoshack/netmonitor/internal/storage"
)

// ExportData represents the structure of exported JSON data
type ExportData struct {
	ExportInfo ExportInfo              `json:"exportInfo"`
	RawData    []*storage.TestResult   `json:"rawData,omitempty"`
	Metadata   map[string]interface{}  `json:"metadata,omitempty"`
}

// ExportInfo contains metadata about the export
type ExportInfo struct {
	ExportID    string `json:"exportId"`
	ExportDate  string `json:"exportDate"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Format      string `json:"format"`
	RecordCount int    `json:"recordCount"`
}

// exportJSON exports data to JSON format
func (m *Manager) exportJSON(ctx context.Context, job *ExportJob, filePath string) error {
	// Prepare export data structure
	exportData := &ExportData{
		ExportInfo: ExportInfo{
			ExportID:   job.ID,
			ExportDate: job.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			StartDate:  job.Request.StartDate.Format("2006-01-02"),
			EndDate:    job.Request.EndDate.Format("2006-01-02"),
			Format:     "json",
		},
		Metadata: make(map[string]interface{}),
	}

	// Fetch raw data if requested
	if job.Request.IncludeRaw {
		rawData, err := m.fetchRawData(ctx, job)
		if err != nil {
			return fmt.Errorf("failed to fetch raw data: %w", err)
		}
		exportData.RawData = rawData
		exportData.ExportInfo.RecordCount = len(rawData)
	}

	// Add metadata
	exportData.Metadata["endpoints"] = job.Request.Endpoints
	exportData.Metadata["regions"] = job.Request.Regions
	exportData.Metadata["includeRaw"] = job.Request.IncludeRaw
	exportData.Metadata["includeAgg"] = job.Request.IncludeAgg

	// Create output file or zip
	var writer io.Writer
	var file *os.File
	var zipWriter *zip.Writer
	var err error

	if job.Request.Compressed {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create zip file: %w", err)
		}
		defer file.Close()

		zipWriter = zip.NewWriter(file)
		defer zipWriter.Close()

		// Create JSON file inside ZIP
		jsonFilename := fmt.Sprintf("export-%s.json", job.ID[:8])
		zipFile, err := zipWriter.Create(jsonFilename)
		if err != nil {
			return fmt.Errorf("failed to create JSON in zip: %w", err)
		}
		writer = zipFile
	} else {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create JSON file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// Write JSON with indentation for readability
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(exportData); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Update final progress
	m.jobsMutex.Lock()
	job.Progress = 1.0
	m.jobsMutex.Unlock()

	return nil
}

// fetchRawData fetches all raw test results for the export
func (m *Manager) fetchRawData(ctx context.Context, job *ExportJob) ([]*storage.TestResult, error) {
	var allResults []*storage.TestResult

	// Calculate total days for progress tracking
	totalDays := int(job.Request.EndDate.Sub(job.Request.StartDate).Hours()/24) + 1
	processedDays := 0

	// Iterate through date range
	currentDate := job.Request.StartDate
	for !currentDate.After(job.Request.EndDate) {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get results for this day
		results, err := m.storage.GetResults(currentDate)
		if err != nil {
			log.Ctx(ctx).Warn().
				Err(err).
				Time("date", currentDate).
				Msg("Failed to get results for date")
		} else {
			// Filter results
			for _, result := range results {
				if m.shouldIncludeResult(result, &job.Request) {
					allResults = append(allResults, result)
				}
			}
		}

		// Update progress
		processedDays++
		progress := float64(processedDays) / float64(totalDays)
		m.jobsMutex.Lock()
		job.Progress = progress * 0.9 // Reserve last 10% for writing
		m.jobsMutex.Unlock()

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return allResults, nil
}
