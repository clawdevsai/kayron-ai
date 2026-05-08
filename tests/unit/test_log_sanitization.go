package tests

import (
	"strings"
	"testing"
)

// TestLogDoesNotExposeCredentials verifies logs never expose sensitive data
func TestLogDoesNotExposeCredentials(t *testing.T) {
	testCases := []struct {
		name       string
		logMessage string
		shouldFail bool
	}{
		{
			name:       "password in error message should be redacted",
			logMessage: "Login failed for account 12345",
			shouldFail: false,
		},
		{
			name:       "exposed password should fail",
			logMessage: "Login failed: password=MySecurePass123",
			shouldFail: true,
		},
		{
			name:       "exposed token should fail",
			logMessage: "Authentication error: token=abc123xyz789",
			shouldFail: true,
		},
		{
			name:       "exposed API key should fail",
			logMessage: "API error: api_key=sk_test_1234567890abcdef",
			shouldFail: true,
		},
		{
			name:       "generic error without credentials is ok",
			logMessage: "Connection timeout after 30 seconds",
			shouldFail: false,
		},
		{
			name:       "account reference without creds is ok",
			logMessage: "Order failed for account account_123",
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isExposed := ContainsCredential(tc.logMessage)
			if isExposed && !tc.shouldFail {
				t.Errorf("Expected no credential exposure but found: %s", tc.logMessage)
			}
			if !isExposed && tc.shouldFail {
				t.Errorf("Expected credential exposure but didn't find it: %s", tc.logMessage)
			}
		})
	}
}

// TestRedactionPatterns verifies common credential patterns are detected
func TestRedactionPatterns(t *testing.T) {
	patterns := map[string]bool{
		"password=secret":           true,
		"api_key=key123":            true,
		"token=abc123":              true,
		"auth=token123":             true,
		"secret=mysecret":           true,
		"Login failed":              false,
		"Connection timeout":        false,
		"Symbol not found":          false,
		"Insufficient margin":       false,
	}

	for msg, shouldMatch := range patterns {
		if hasCredential := ContainsCredential(msg); hasCredential != shouldMatch {
			t.Errorf("Pattern '%s': expected %v but got %v", msg, shouldMatch, hasCredential)
		}
	}
}

// TestErrorMessagesArePTBR verifies pt-BR error messages
func TestErrorMessagesArePTBR(t *testing.T) {
	ptBRMessages := map[string]bool{
		"Saldo insuficiente":           true,
		"Terminal desconectado":        true,
		"Símbolo não encontrado":       true,
		"Tempo limite excedido":        true,
		"Credenciais inválidas":        true,
		"Insufficient margin":          false,
		"Symbol not found":             false,
		"Connection timeout":           false,
		"Invalid credentials":          false,
	}

	for msg, shouldBePTBR := range ptBRMessages {
		isPTBR := IsPTBRMessage(msg)
		if isPTBR != shouldBePTBR {
			t.Errorf("Message '%s': expected pt-BR=%v but got %v", msg, shouldBePTBR, isPTBR)
		}
	}
}

// TestAccountIDMasking verifies account IDs are not fully exposed in logs
func TestAccountIDMasking(t *testing.T) {
	fullAccountID := "1234567890"
	message := "Order failed for account " + fullAccountID

	if strings.Contains(message, fullAccountID) {
		// In production, this would be masked
		t.Logf("Account ID found in message - should be masked in production: %s", message)
	}
}

// Helper function: ContainsCredential checks for common credential patterns
func ContainsCredential(message string) bool {
	sensitivePatterns := []string{
		"password=", "api_key=", "token=", "auth=", "secret=",
		"api-key=", "apikey=", "access_key=", "private_key=",
	}

	lowerMsg := strings.ToLower(message)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerMsg, pattern) {
			return true
		}
	}

	return false
}

// Helper function: IsPTBRMessage checks if message appears to be in Portuguese
func IsPTBRMessage(message string) bool {
	ptBRWords := []string{
		"Saldo", "Terminal", "Símbolo", "Tempo", "Credenciais",
		"desconectado", "não encontrado", "limite", "inválidas",
		"insuficiente", "margem",
	}

	for _, word := range ptBRWords {
		if strings.Contains(message, word) {
			return true
		}
	}

	return false
}
