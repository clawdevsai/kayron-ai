package config

import (
	"fmt"
	"os"
	"time"
)

// SecretsConfig holds secrets management configuration
type SecretsConfig struct {
	Source           string // "env", "aws_secrets_manager", "vault"
	MT5Login         string
	MT5Password      string
	MT5ServerAddress string
	LastRotation     time.Time
	RotationInterval time.Duration
}

// LoadSecrets loads credentials from configured source
func LoadSecrets(source string) (*SecretsConfig, error) {
	if source == "" {
		source = "env"
	}

	sc := &SecretsConfig{
		Source:           source,
		RotationInterval: 24 * time.Hour, // Default 24-hour rotation
	}

	switch source {
	case "env":
		return loadFromEnv(sc)
	case "aws_secrets_manager":
		return loadFromAWSSecretsManager(sc)
	default:
		return loadFromEnv(sc)
	}
}

// loadFromEnv loads secrets from environment variables
func loadFromEnv(sc *SecretsConfig) (*SecretsConfig, error) {
	login := os.Getenv("MT5_LOGIN")
	password := os.Getenv("MT5_PASSWORD")
	server := os.Getenv("MT5_SERVER")

	if login == "" || password == "" || server == "" {
		return nil, fmt.Errorf("missing required MT5 environment variables")
	}

	sc.MT5Login = login
	sc.MT5Password = password
	sc.MT5ServerAddress = server
	sc.LastRotation = time.Now()

	return sc, nil
}

// loadFromAWSSecretsManager loads secrets from AWS Secrets Manager
func loadFromAWSSecretsManager(sc *SecretsConfig) (*SecretsConfig, error) {
	// Placeholder implementation
	// In production, use AWS SDK v2
	// secretArn := os.Getenv("AWS_SECRET_ARN")
	// region := os.Getenv("AWS_REGION")

	// For now, fall back to environment
	return loadFromEnv(sc)
}

// RequiresRotation checks if credentials need rotation
func (sc *SecretsConfig) RequiresRotation() bool {
	return time.Since(sc.LastRotation) > sc.RotationInterval
}

// RotateCredentials rotates credentials from source
func (sc *SecretsConfig) RotateCredentials() error {
	switch sc.Source {
	case "env":
		// Reload from environment
		updated, err := loadFromEnv(&SecretsConfig{})
		if err != nil {
			return fmt.Errorf("failed to rotate credentials from env: %w", err)
		}
		sc.MT5Login = updated.MT5Login
		sc.MT5Password = updated.MT5Password
		sc.MT5ServerAddress = updated.MT5ServerAddress
		sc.LastRotation = time.Now()
		return nil
	case "aws_secrets_manager":
		// In production, fetch from Secrets Manager
		return fmt.Errorf("AWS Secrets Manager rotation not yet implemented")
	default:
		return fmt.Errorf("unknown secrets source: %s", sc.Source)
	}
}

// Validate validates the secrets configuration
func (sc *SecretsConfig) Validate() error {
	if sc.MT5Login == "" {
		return fmt.Errorf("MT5_LOGIN not configured")
	}
	if sc.MT5Password == "" {
		return fmt.Errorf("MT5_PASSWORD not configured")
	}
	if sc.MT5ServerAddress == "" {
		return fmt.Errorf("MT5_SERVER not configured")
	}
	return nil
}

// GetMaskedLogin returns login with masked characters for logging
func (sc *SecretsConfig) GetMaskedLogin() string {
	if len(sc.MT5Login) <= 4 {
		return "****"
	}
	return sc.MT5Login[:2] + "****" + sc.MT5Login[len(sc.MT5Login)-2:]
}

// CheckIfExpired checks if credentials are expired (placeholder)
func (sc *SecretsConfig) CheckIfExpired() bool {
	// Placeholder - in production, check with credential provider
	return false
}

// NeverLogPassword ensures password is never logged
func (sc *SecretsConfig) NeverLogPassword() string {
	return "***REDACTED***"
}
