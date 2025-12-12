package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

type InputRecord struct {
	Timestamp  string      `json:"timestamp"`
	EndpointID string      `json:"endpoint_id"`
	LatencyMS  interface{} `json:"latency_ms"` // Could be float or int
	Status     string      `json:"status"`
}

type OutputRecord struct {
	Ts int64       `json:"ts"`
	Id string      `json:"id"`
	Ms interface{} `json:"ms"`
	St int         `json:"st"`
}

func main() {
	inputFile := "D:\\Workspaces\\netmonitorv2\\build\\bin\\data\\2025-12-10.json"
	outputFile := "D:\\Workspaces\\netmonitorv2\\build\\bin\\data\\2025-12-10_optimized.json"

	// Check file existence
	info, err := os.Stat(inputFile)
	if err != nil {
		fmt.Printf("Error stating file: %v\n", err)
		os.Exit(1)
	}
	originalSize := info.Size()

	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var input []InputRecord
	if err := json.Unmarshal(content, &input); err != nil {
		fmt.Printf("Error unmarshaling input: %v\n", err)
		os.Exit(1)
	}

	output := make([]OutputRecord, len(input))
	endpointIdByName := make(map[string]bool)

	for i, r := range input {
		// Parse timestamp
		// Format: 2025-12-10T00:00:00.70894-05:00
		t, err := time.Parse(time.RFC3339, r.Timestamp)
		if err != nil {
			fmt.Printf("Error parsing time sat index %d (%s): %v\n", i, r.Timestamp, err)
			continue
		}

		shortId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(r.EndpointID)).String()[:7]
		var ok bool

		if _, ok = endpointIdByName[shortId]; !ok {
			endpointIdByName[shortId] = true
		}

		var status int
		switch r.Status {
		case "success":
			status = 0
		case "failure":
			status = 1
		}

		output[i] = OutputRecord{
			Ts: t.Unix(), // Seconds only
			Id: shortId,
			Ms: r.LatencyMS,
			St: status,
		}
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if _, err := file.WriteString("[\n"); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	for i, r := range output {
		line, err := json.Marshal(r)
		if err != nil {
			fmt.Printf("Error marshaling record %d: %v\n", i, err)
			continue
		}
		if _, err := file.WriteString("  "); err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			os.Exit(1)
		}
		if _, err := file.Write(line); err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			os.Exit(1)
		}
		if i < len(output)-1 {
			if _, err := file.WriteString(","); err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				os.Exit(1)
			}
		}
		if _, err := file.WriteString("\n"); err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			os.Exit(1)
		}
	}
	if _, err := file.WriteString("]"); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	newInfo, _ := os.Stat(outputFile)
	newSize := newInfo.Size()

	fmt.Printf("Original Size: %d bytes\n", originalSize)
	fmt.Printf("New Size: %d bytes\n", newSize)
	fmt.Printf("Reduction: %.2f%%\n", float64(originalSize-newSize)/float64(originalSize)*100)
}
