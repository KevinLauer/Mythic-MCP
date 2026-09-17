package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds the MCP server configuration
type Config struct {
	// Mythic connection settings
	MythicURL     string
	APIToken      string
	Username      string
	Password      string
	SSL           bool
	SkipTLSVerify bool

	// Server settings
	LogLevel   string
	Timeout    time.Duration
	MCPProfile string

	// Optional mythic-cli on the Mythic host (never SSH)
	CLIPath    string
	MythicHome string

	// File vending settings
	FileVendingEnabled  bool
	FileVendingBaseURL  string
	FileStoragePath     string
	FileTokenExpiry     time.Duration
	FileMaxSizeMB       int
	FileCleanupInterval time.Duration
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	ssl := getEnvBool("MYTHIC_SSL", true)
	cfg := &Config{
		MythicURL:     mythicURLFromEnv(ssl),
		APIToken:      firstEnv("MYTHIC_API_TOKEN", "MYTHIC_APITOKEN"),
		Username:      os.Getenv("MYTHIC_USERNAME"),
		Password:      os.Getenv("MYTHIC_PASSWORD"),
		SSL:           ssl,
		SkipTLSVerify: getEnvBool("MYTHIC_SKIP_TLS_VERIFY", false),
		LogLevel:      getEnvString("LOG_LEVEL", "info"),
		Timeout:       getEnvDuration("TIMEOUT", 30*time.Second),
		MCPProfile:    getEnvString("MYTHIC_MCP_PROFILE", "operator"),
		CLIPath:       os.Getenv("MYTHIC_CLI_PATH"),
		MythicHome:    os.Getenv("MYTHIC_HOME"),

		FileVendingEnabled:  getEnvBool("FILE_VENDING_ENABLED", true),
		FileVendingBaseURL:  getEnvString("FILE_VENDING_BASE_URL", ""),
		FileStoragePath:     getEnvString("FILE_STORAGE_PATH", "/tmp/mythic-files"),
		FileTokenExpiry:     getEnvDuration("FILE_TOKEN_EXPIRY", 5*time.Minute),
		FileMaxSizeMB:       getEnvInt("FILE_MAX_SIZE_MB", 100),
		FileCleanupInterval: getEnvDuration("FILE_CLEANUP_INTERVAL", 60*time.Second),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) OperatorProfile() bool {
	profile := strings.ToLower(strings.TrimSpace(c.MCPProfile))
	return profile == "" || profile == "operator"
}

func mythicURLFromEnv(ssl bool) string {
	if u := os.Getenv("MYTHIC_URL"); u != "" {
		return u
	}
	ip := os.Getenv("MYTHIC_IP")
	if ip == "" {
		return ""
	}
	port := getEnvString("MYTHIC_PORT", "7443")
	scheme := "https"
	if !ssl {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s:%s", scheme, ip, port)
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

// Validate checks that required configuration is present.
// Only MYTHIC_URL is required at startup; credentials are optional because
// the user can authenticate later via the mythic_login MCP tool.
func (c *Config) Validate() error {
	if c.MythicURL == "" {
		return fmt.Errorf("MYTHIC_URL is required")
	}

	return nil
}

// getEnvString returns environment variable value or default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns environment variable as bool or default
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// getEnvDuration returns environment variable as duration or default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

// getEnvInt returns environment variable as int or default
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
		return defaultValue
	}
	return n
}
