package export

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/marcoshack/netmonitor/internal/storage"
)

// exportCSV exports data to CSV format
func (m *Manager) exportCSV(ctx context.Context, job *ExportJob, filePath string) error {
	// Determine columns
	columns := job.Request.Columns
	if len(columns) == 0 {
		columns = DefaultCSVColumns()
	}

	// Create output file or zip
	var writer io.Writer
	var file *os.File
	var zipWriter *zip.Writer
	var zipFile io.Writer
	var err error

	if job.Request.Compressed {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create zip file: %w", err)
		}
		defer file.Close()

		zipWriter = zip.NewWriter(file)
		defer zipWriter.Close()

		// Create CSV file inside ZIP
		csvFilename := fmt.Sprintf("export-%s.csv", job.ID[:8])
		zipFile, err = zipWriter.Create(csvFilename)
		if err != nil {
			return fmt.Errorf("failed to create CSV in zip: %w", err)
		}
		writer = zipFile
	} else {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create CSV file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	if err := csvWriter.Write(columns); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Fetch and write data
	if job.Request.IncludeRaw {
		if err := m.writeRawDataCSV(ctx, job, csvWriter, columns); err != nil {
			return err
		}
	}

	// Flush CSV writer before closing zip
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("CSV writer error: %w", err)
	}

	return nil
}

// writeRawDataCSV writes raw test result data to CSV
func (m *Manager) writeRawDataCSV(ctx context.Context, job *ExportJob, csvWriter *csv.Writer, columns []string) error {
	// Calculate total days for progress tracking
	totalDays := int(job.Request.EndDate.Sub(job.Request.StartDate).Hours()/24) + 1
	processedDays := 0

	// Iterate through date range
	currentDate := job.Request.StartDate
	for !currentDate.After(job.Request.EndDate) {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
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
			// Filter and write results
			for _, result := range results {
				if m.shouldIncludeResult(result, &job.Request) {
					row := m.resultToCSVRow(result, columns)
					if err := csvWriter.Write(row); err != nil {
						return fmt.Errorf("failed to write CSV row: %w", err)
					}
				}
			}
		}

		// Update progress
		processedDays++
		progress := float64(processedDays) / float64(totalDays)
		m.jobsMutex.Lock()
		job.Progress = progress * 0.9 // Reserve last 10% for finalization
		m.jobsMutex.Unlock()

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return nil
}

// shouldIncludeResult checks if a result should be included based on filters
func (m *Manager) shouldIncludeResult(result *storage.TestResult, request *ExportRequest) bool {
	// Check timestamp range
	if result.Timestamp.Before(request.StartDate) || result.Timestamp.After(request.EndDate) {
		return false
	}

	// Check endpoint filter
	if len(request.Endpoints) > 0 {
		found := false
		for _, endpoint := range request.Endpoints {
			if result.EndpointID == endpoint {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Note: Region filtering would require endpoint-to-region mapping
	// This could be added by passing config.Manager to the export package
	// For now, we'll skip region filtering at the result level

	return true
}

// resultToCSVRow converts a test result to a CSV row
func (m *Manager) resultToCSVRow(result *storage.TestResult, columns []string) []string {
	row := make([]string, len(columns))

	for i, col := range columns {
		switch col {
		case ColumnTimestamp:
			row[i] = result.Timestamp.Format(time.RFC3339)
		case ColumnEndpointID:
			row[i] = result.EndpointID
		case ColumnRegion:
			row[i] = "" // Would need config mapping
		case ColumnProtocol:
			row[i] = result.Protocol
		case ColumnStatus:
			row[i] = result.Status
		case ColumnLatency:
			row[i] = fmt.Sprintf("%.2f", float64(result.Latency.Nanoseconds())/1_000_000.0)
		case ColumnError:
			row[i] = result.Error
		default:
			row[i] = ""
		}
	}

	return row
}
