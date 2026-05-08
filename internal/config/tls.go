package config

import (
	"crypto/tls"
	"fmt"
	"os"
)

// TLSConfig holds TLS configuration for gRPC
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
	Cert     *tls.Certificate
}

// LoadTLSConfig loads TLS configuration from environment
func LoadTLSConfig(env string) (*TLSConfig, error) {
	tlsConfig := &TLSConfig{
		Enabled: env == "production",
	}

	if !tlsConfig.Enabled {
		return tlsConfig, nil
	}

	// In production, load from environment or configuration
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("TLS enabled but TLS_CERT_FILE or TLS_KEY_FILE not set")
	}

	tlsConfig.CertFile = certFile
	tlsConfig.KeyFile = keyFile

	// Load certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	tlsConfig.Cert = &cert
	return tlsConfig, nil
}

// GetServerTLSConfig returns gRPC TLS configuration for server
func (tc *TLSConfig) GetServerTLSConfig() *tls.Config {
	if !tc.Enabled || tc.Cert == nil {
		return nil
	}

	return &tls.Config{
		Certificates: []tls.Certificate{*tc.Cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}

// ValidateTLSConfig validates TLS configuration
func (tc *TLSConfig) ValidateTLSConfig() error {
	if !tc.Enabled {
		return nil
	}

	if tc.Cert == nil {
		return fmt.Errorf("TLS enabled but certificate not loaded")
	}

	if tc.CertFile == "" || tc.KeyFile == "" {
		return fmt.Errorf("TLS enabled but certificate paths not configured")
	}

	// Verify files exist
	if _, err := os.Stat(tc.CertFile); err != nil {
		return fmt.Errorf("certificate file not found: %w", err)
	}

	if _, err := os.Stat(tc.KeyFile); err != nil {
		return fmt.Errorf("key file not found: %w", err)
	}

	return nil
}
