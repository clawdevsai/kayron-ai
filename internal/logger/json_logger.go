package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// JSONLogEntry represents a structured log entry for tool invocation
type JSONLogEntry struct {
	Timestamp   time.Time       `json:"timestamp"`
	Level       string          `json:"level"`
	ToolName    string          `json:"tool_name,omitempty"`
	Input       interface{}     `json:"input,omitempty"`
	Output      interface{}     `json:"output,omitempty"`
	Error       string          `json:"error,omitempty"`
	LatencyMS   int64           `json:"latency_ms,omitempty"`
	AccountID   string          `json:"account_id,omitempty"`
	Message     string          `json:"message,omitempty"`
	RequestID   string          `json:"request_id,omitempty"`
}

// JSONLogger handles structured JSON logging
type JSONLogger struct {
	mu        sync.Mutex
	stdout    bool
	filePath  string
	file      *os.File
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger(filePath string) (*JSONLogger, error) {
	var file *os.File
	var err error

	if filePath != "" {
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
	}

	return &JSONLogger{
		stdout:   true,
		filePath: filePath,
		file:     file,
	}, nil
}

// LogToolInvocation logs a tool invocation with input/output
func (jl *JSONLogger) LogToolInvocation(toolName string, accountID string, input interface{}, output interface{}, latencyMS int64, err error) {
	entry := JSONLogEntry{
		Timestamp:  time.Now().UTC(),
		Level:      "INFO",
		ToolName:   toolName,
		Input:      input,
		Output:     output,
		LatencyMS:  latencyMS,
		AccountID:  accountID,
	}

	if err != nil {
		entry.Error = err.Error()
		entry.Level = "ERROR"
	}

	jl.writeEntry(entry)
}

// LogError logs an error
func (jl *JSONLogger) LogError(message string, err error, accountID string) {
	entry := JSONLogEntry{
		Timestamp:  time.Now().UTC(),
		Level:      "ERROR",
		Message:    message,
		Error:      err.Error(),
		AccountID:  accountID,
	}

	jl.writeEntry(entry)
}

// LogInfo logs an info message
func (jl *JSONLogger) LogInfo(message string, accountID string, data map[string]interface{}) {
	entry := JSONLogEntry{
		Timestamp:  time.Now().UTC(),
		Level:      "INFO",
		Message:    message,
		AccountID:  accountID,
		Output:     data,
	}

	jl.writeEntry(entry)
}

// LogDebug logs debug-level message
func (jl *JSONLogger) LogDebug(message string, data map[string]interface{}) {
	entry := JSONLogEntry{
		Timestamp: time.Now().UTC(),
		Level:     "DEBUG",
		Message:   message,
		Output:    data,
	}

	jl.writeEntry(entry)
}

// LogWarning logs a warning
func (jl *JSONLogger) LogWarning(message string, accountID string) {
	entry := JSONLogEntry{
		Timestamp:  time.Now().UTC(),
		Level:      "WARN",
		Message:    message,
		AccountID:  accountID,
	}

	jl.writeEntry(entry)
}

// writeEntry writes a log entry to stdout and file
func (jl *JSONLogger) writeEntry(entry JSONLogEntry) {
	jl.mu.Lock()
	defer jl.mu.Unlock()

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	jsonStr := string(jsonBytes) + "\n"

	// Write to stdout
	if jl.stdout {
		fmt.Print(jsonStr)
	}

	// Write to file if configured
	if jl.file != nil {
		_, _ = jl.file.WriteString(jsonStr)
	}
}

// Close closes the log file
func (jl *JSONLogger) Close() error {
	if jl.file != nil {
		return jl.file.Close()
	}
	return nil
}
