package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	MT5Login    string
	MT5Password string
	MT5Server   string
	MT5Timeout  time.Duration
	GRPCPort    int
	HTTPPort    int
	DebugMode   bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		MT5Login:    getEnv("MT5_LOGIN", ""),
		MT5Password: getEnv("MT5_PASSWORD", ""),
		MT5Server:   getEnv("MT5_SERVER", "localhost"),
		MT5Timeout:  getDuration("MT5_TIMEOUT", 30*time.Second),
		GRPCPort:    getInt("GRPC_PORT", 50051),
		HTTPPort:    getInt("HTTP_PORT", 8080),
		DebugMode:   getBool("DEBUG", false),
	}
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getInt retrieves an environment variable as integer
func getInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getBool retrieves an environment variable as boolean
func getBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// getDuration retrieves an environment variable as duration
func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
