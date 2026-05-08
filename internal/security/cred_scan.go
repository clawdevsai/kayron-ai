package security

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CredentialScanner scans source files for hardcoded credentials
type CredentialScanner struct {
	patterns []*regexp.Regexp
}

// NewCredentialScanner creates a new credential scanner
func NewCredentialScanner() *CredentialScanner {
	return &CredentialScanner{
		patterns: initScanPatterns(),
	}
}

// initScanPatterns initializes regex patterns for credential detection
func initScanPatterns() []*regexp.Regexp {
	patterns := []string{
		`(?i)mt5_login\s*=\s*["']?[^"'\s]+["']?`,
		`(?i)mt5_password\s*=\s*["']?[^"'\s]+["']?`,
		`(?i)password\s*=\s*["']?[^"'\s]+["']?`,
		`(?i)api_key\s*=\s*["']?[a-zA-Z0-9_\-]{20,}["']?`,
		`(?i)secret\s*=\s*["']?[a-zA-Z0-9_\-]{20,}["']?`,
		`(?i)token\s*=\s*["']?[a-zA-Z0-9_\-]{20,}["']?`,
		`(?i)auth\s*=\s*["']?[^"'\s]+["']?`,
		`(?i)credentials\s*=\s*["']?[^"'\s]+["']?`,
		`aws_access_key_id\s*=\s*[A-Z0-9]{20}`,
		`aws_secret_access_key\s*=\s*[a-zA-Z0-9/+=]{40}`,
	}

	compiled := make([]*regexp.Regexp, 0)
	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, re)
		}
	}

	return compiled
}

// CredentialMatch represents a found credential
type CredentialMatch struct {
	File     string
	Line     int
	Match    string
	Pattern  string
}

// ScanFile scans a single file for credentials
func (cs *CredentialScanner) ScanFile(filePath string) ([]CredentialMatch, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var matches []CredentialMatch
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		for _, pattern := range cs.patterns {
			if pattern.MatchString(line) {
				match := CredentialMatch{
					File:    filePath,
					Line:    lineNum + 1,
					Match:   line,
					Pattern: pattern.String(),
				}
				matches = append(matches, match)
			}
		}
	}

	return matches, nil
}

// ScanDirectory recursively scans directory for credentials
func (cs *CredentialScanner) ScanDirectory(dirPath string) ([]CredentialMatch, error) {
	var allMatches []CredentialMatch

	// Extensions to scan
	scanExtensions := map[string]bool{
		".go": true, ".env": true, ".yaml": true, ".yml": true,
		".json": true, ".conf": true, ".config": true, ".txt": true,
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-source files
		if info.IsDir() {
			// Skip common dirs
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !scanExtensions[ext] {
			return nil
		}

		matches, err := cs.ScanFile(path)
		if err != nil {
			return nil
		}

		allMatches = append(allMatches, matches...)
		return nil
	})

	return allMatches, err
}

// IsCredentialExposed checks if any credentials are hardcoded
func (cs *CredentialScanner) IsCredentialExposed(matches []CredentialMatch) bool {
	return len(matches) > 0
}

// ReportMatches prints credential scan results
func (cs *CredentialScanner) ReportMatches(matches []CredentialMatch) {
	if len(matches) == 0 {
		fmt.Println("No hardcoded credentials detected.")
		return
	}

	fmt.Printf("SECURITY ALERT: Found %d potential hardcoded credentials:\n\n", len(matches))
	for _, match := range matches {
		fmt.Printf("File: %s (Line %d)\n", match.File, match.Line)
		fmt.Printf("Match: %s\n", match.Match)
		fmt.Printf("Pattern: %s\n\n", match.Pattern)
	}
}
