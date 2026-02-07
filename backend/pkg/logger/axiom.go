package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	axiomToken   = os.Getenv("AXIOM_TOKEN")
	axiomDataset = os.Getenv("AXIOM_DATASET")
	axiomEnabled = axiomToken != "" && axiomDataset != ""
	httpClient   = &http.Client{Timeout: 5 * time.Second}

	// Buffer for batching logs
	logBuffer   []map[string]interface{}
	bufferMutex sync.Mutex
	bufferSize  = 100 // Send logs in batches of 100
)

// LogData represents a structured log entry
type LogData map[string]interface{}

// Log sends a structured log entry to console and Axiom
func Log(level string, message string, data map[string]interface{}) {
	entry := LogData{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}

	// Merge additional data
	for k, v := range data {
		entry[k] = v
	}

	// Always log to console
	logJSON, _ := json.Marshal(entry)
	log.Println(string(logJSON))

	// Send to Axiom if enabled
	if axiomEnabled {
		bufferMutex.Lock()
		logBuffer = append(logBuffer, entry)
		shouldFlush := len(logBuffer) >= bufferSize
		bufferMutex.Unlock()

		if shouldFlush {
			go flushToAxiom()
		}
	}
}

// flushToAxiom sends buffered logs to Axiom
func flushToAxiom() {
	bufferMutex.Lock()
	if len(logBuffer) == 0 {
		bufferMutex.Unlock()
		return
	}

	// Copy buffer and clear it
	logs := make([]map[string]interface{}, len(logBuffer))
	copy(logs, logBuffer)
	logBuffer = logBuffer[:0]
	bufferMutex.Unlock()

	// Send to Axiom
	payload, err := json.Marshal(logs)
	if err != nil {
		log.Printf("Failed to marshal logs for Axiom: %v", err)
		return
	}

	url := fmt.Sprintf("https://api.axiom.co/v1/datasets/%s/ingest", axiomDataset)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to create Axiom request: %v", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+axiomToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to send logs to Axiom: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("Axiom returned non-OK status: %d", resp.StatusCode)
	}
}

// FlushLogs forces flush of buffered logs (call before shutdown)
func FlushLogs() {
	flushToAxiom()
}

// Convenience functions
func Info(message string, data map[string]interface{}) {
	Log("info", message, data)
}

func Warn(message string, data map[string]interface{}) {
	Log("warn", message, data)
}

func Error(message string, data map[string]interface{}) {
	Log("error", message, data)
}

func Security(message string, data map[string]interface{}) {
	Log("security", message, data)
}

// Start periodic flush (call once at startup)
func StartPeriodicFlush() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			flushToAxiom()
		}
	}()
}
