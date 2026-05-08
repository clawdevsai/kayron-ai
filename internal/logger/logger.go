package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Logger provides structured JSON logging
type Logger struct {
	component string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string      `json:"timestamp"`
	Level      string      `json:"level"`
	Component  string      `json:"component"`
	Message    string      `json:"message"`
	LatencyMs  int64       `json:"latency_ms,omitempty"`
	Error      string      `json:"error,omitempty"`
	ExtraField interface{} `json:"extra,omitempty"`
}

// New creates a new logger for a component
func New(component string) *Logger {
	return &Logger{
		component: component,
	}
}

// log writes a structured log entry
func (l *Logger) log(level string, message string, err error, latencyMs int64, extra interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Component: l.component,
		Message:   message,
		LatencyMs: latencyMs,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if extra != nil {
		entry.ExtraField = extra
	}

	data, _ := json.Marshal(entry)
	log.Println(string(data))
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.log("INFO", message, nil, 0, nil)
}

// InfoWithLatency logs an info message with latency
func (l *Logger) InfoWithLatency(message string, latencyMs int64) {
	l.log("INFO", message, nil, latencyMs, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log("WARN", message, nil, 0, nil)
}

// WarnWithError logs a warning with error details
func (l *Logger) WarnWithError(message string, err error) {
	l.log("WARN", message, err, 0, nil)
}

// Error logs an error message
func (l *Logger) Error(message string, err error) {
	l.log("ERROR", message, err, 0, nil)
}

// ErrorWithLatency logs an error with latency
func (l *Logger) ErrorWithLatency(message string, err error, latencyMs int64) {
	l.log("ERROR", message, err, latencyMs, nil)
}

// Debug logs a debug message (only in debug mode)
func (l *Logger) Debug(message string) {
	if os.Getenv("DEBUG") == "true" {
		l.log("DEBUG", message, nil, 0, nil)
	}
}

// WithExtra attaches extra data to a log entry
func (l *Logger) WithExtra(level string, message string, err error, extra interface{}) {
	l.log(level, message, err, 0, extra)
}
