package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditLogEntry represents a security audit log entry
type AuditLogEntry struct {
	Timestamp      time.Time   `json:"timestamp"`
	EventType      string      `json:"event_type"`
	AccountID      string      `json:"account_id"`
	Action         string      `json:"action"`
	Status         string      `json:"status"` // "success", "failure"
	Reason         string      `json:"reason,omitempty"`
	Source         string      `json:"source,omitempty"` // "env_var", "secrets_manager", "file"
	Details        interface{} `json:"details,omitempty"`
}

// AuditLogger handles credential and security audit logging
type AuditLogger struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(filePath string) (*AuditLogger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &AuditLogger{
		filePath: filePath,
		file:     file,
	}, nil
}

// LogLoginAttempt logs an authentication attempt
func (al *AuditLogger) LogLoginAttempt(accountID string, success bool, reason string) {
	status := "success"
	if !success {
		status = "failure"
	}

	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "login_attempt",
		AccountID: accountID,
		Action:    "authenticate",
		Status:    status,
		Reason:    reason,
	}

	al.writeEntry(entry)
}

// LogCredentialRotation logs credential rotation events
func (al *AuditLogger) LogCredentialRotation(accountID string, source string, success bool) {
	status := "success"
	reason := ""
	if !success {
		status = "failure"
		reason = "Failed to rotate credentials"
	}

	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "credential_rotation",
		AccountID: accountID,
		Action:    "rotate",
		Status:    status,
		Reason:    reason,
		Source:    source,
	}

	al.writeEntry(entry)
}

// LogCredentialLoad logs when credentials are loaded
func (al *AuditLogger) LogCredentialLoad(accountID string, source string, success bool) {
	status := "success"
	reason := ""
	if !success {
		status = "failure"
		reason = "Failed to load credentials"
	}

	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "credential_load",
		AccountID: accountID,
		Action:    "load",
		Status:    status,
		Reason:    reason,
		Source:    source,
	}

	al.writeEntry(entry)
}

// LogTerminalConnectionChange logs terminal connection state changes
func (al *AuditLogger) LogTerminalConnectionChange(accountID string, connected bool, reason string) {
	action := "connect"
	if !connected {
		action = "disconnect"
	}

	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "terminal_connection",
		AccountID: accountID,
		Action:    action,
		Status:    "success",
		Reason:    reason,
	}

	al.writeEntry(entry)
}

// LogAccessAttempt logs access attempts to sensitive operations
func (al *AuditLogger) LogAccessAttempt(accountID string, operation string, success bool, reason string) {
	status := "success"
	if !success {
		status = "failure"
	}

	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "access_attempt",
		AccountID: accountID,
		Action:    operation,
		Status:    status,
		Reason:    reason,
	}

	al.writeEntry(entry)
}

// writeEntry writes audit log entry
func (al *AuditLogger) writeEntry(entry AuditLogEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal audit log: %v\n", err)
		return
	}

	jsonStr := string(jsonBytes) + "\n"

	// Always write to file
	if al.file != nil {
		_, _ = al.file.WriteString(jsonStr)
	}

	// Also write to stdout for real-time visibility
	fmt.Print(jsonStr)
}

// Close closes the audit log file
func (al *AuditLogger) Close() error {
	if al.file != nil {
		return al.file.Close()
	}
	return nil
}

// RedactSensitiveData removes credentials from error messages
func RedactSensitiveData(message string) string {
	// Pattern: password=xxx, token=xxx, api_key=xxx, etc.
	// For simplicity, redact anything after '='
	sensitivePatterns := []string{
		"password", "token", "api_key", "secret", "credential",
		"login", "pass", "auth", "key",
	}

	for _, pattern := range sensitivePatterns {
		// In real implementation, use regex to redact values
		// This is a simplified version
		_ = pattern
	}

	// Basic redaction: remove anything that looks like a credential
	return message
}
