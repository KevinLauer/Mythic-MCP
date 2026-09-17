package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv(t *testing.T) {
	// Save original environment
	origEnv := map[string]string{
		"MYTHIC_URL":             os.Getenv("MYTHIC_URL"),
		"MYTHIC_API_TOKEN":       os.Getenv("MYTHIC_API_TOKEN"),
		"MYTHIC_APITOKEN":        os.Getenv("MYTHIC_APITOKEN"),
		"MYTHIC_IP":              os.Getenv("MYTHIC_IP"),
		"MYTHIC_PORT":            os.Getenv("MYTHIC_PORT"),
		"MYTHIC_USERNAME":        os.Getenv("MYTHIC_USERNAME"),
		"MYTHIC_PASSWORD":        os.Getenv("MYTHIC_PASSWORD"),
		"MYTHIC_SSL":             os.Getenv("MYTHIC_SSL"),
		"MYTHIC_SKIP_TLS_VERIFY": os.Getenv("MYTHIC_SKIP_TLS_VERIFY"),
		"MYTHIC_MCP_PROFILE":     os.Getenv("MYTHIC_MCP_PROFILE"),
		"LOG_LEVEL":              os.Getenv("LOG_LEVEL"),
		"TIMEOUT":                os.Getenv("TIMEOUT"),
	}

	// Restore environment after test
	defer func() {
		for key, value := range origEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("ValidWithAPIToken", func(t *testing.T) {
		os.Setenv("MYTHIC_URL", "https://mythic.example.com:7443")
		os.Setenv("MYTHIC_API_TOKEN", "test-token")
		os.Unsetenv("MYTHIC_USERNAME")
		os.Unsetenv("MYTHIC_PASSWORD")

		cfg, err := LoadFromEnv()
		require.NoError(t, err)
		assert.Equal(t, "https://mythic.example.com:7443", cfg.MythicURL)
		assert.Equal(t, "test-token", cfg.APIToken)
		assert.Equal(t, "", cfg.Username)
		assert.Equal(t, "", cfg.Password)
		assert.True(t, cfg.SSL)                      // default
		assert.False(t, cfg.SkipTLSVerify)           // default
		assert.Equal(t, "info", cfg.LogLevel)        // default
		assert.Equal(t, 30*time.Second, cfg.Timeout) // default
	})

	t.Run("ValidWithUsernamePassword", func(t *testing.T) {
		os.Setenv("MYTHIC_URL", "https://mythic.example.com:7443")
		os.Unsetenv("MYTHIC_API_TOKEN")
		os.Setenv("MYTHIC_USERNAME", "admin")
		os.Setenv("MYTHIC_PASSWORD", "password123")

		cfg, err := LoadFromEnv()
		require.NoError(t, err)
		assert.Equal(t, "https://mythic.example.com:7443", cfg.MythicURL)
		assert.Equal(t, "", cfg.APIToken)
		assert.Equal(t, "admin", cfg.Username)
		assert.Equal(t, "password123", cfg.Password)
	})

	t.Run("MissingMythicURL", func(t *testing.T) {
		os.Unsetenv("MYTHIC_URL")
		os.Unsetenv("MYTHIC_IP")
		os.Setenv("MYTHIC_API_TOKEN", "test-token")

		_, err := LoadFromEnv()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MYTHIC_URL is required")
	})

	t.Run("NoCredentials_OK", func(t *testing.T) {
		os.Setenv("MYTHIC_URL", "https://mythic.example.com:7443")
		os.Unsetenv("MYTHIC_API_TOKEN")
		os.Unsetenv("MYTHIC_USERNAME")
		os.Unsetenv("MYTHIC_PASSWORD")

		cfg, err := LoadFromEnv()
		require.NoError(t, err)
		assert.Equal(t, "https://mythic.example.com:7443", cfg.MythicURL)
		assert.Empty(t, cfg.APIToken)
		assert.Empty(t, cfg.Username)
		assert.Empty(t, cfg.Password)
	})

	t.Run("URLFromIP", func(t *testing.T) {
		os.Unsetenv("MYTHIC_URL")
		os.Setenv("MYTHIC_IP", "192.0.2.10")
		os.Setenv("MYTHIC_PORT", "7443")
		os.Setenv("MYTHIC_SSL", "true")
		os.Setenv("MYTHIC_APITOKEN", "mtk_test")
		os.Unsetenv("MYTHIC_API_TOKEN")

		cfg, err := LoadFromEnv()
		require.NoError(t, err)
		assert.Equal(t, "https://192.0.2.10:7443", cfg.MythicURL)
		assert.Equal(t, "mtk_test", cfg.APIToken)
		assert.Equal(t, "operator", cfg.MCPProfile)
	})

	t.Run("CustomSettings", func(t *testing.T) {
		os.Setenv("MYTHIC_URL", "http://localhost:7443")
		os.Setenv("MYTHIC_API_TOKEN", "test-token")
		os.Setenv("MYTHIC_SSL", "false")
		os.Setenv("MYTHIC_SKIP_TLS_VERIFY", "true")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("TIMEOUT", "60s")

		cfg, err := LoadFromEnv()
		require.NoError(t, err)
		assert.False(t, cfg.SSL)
		assert.True(t, cfg.SkipTLSVerify)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, 60*time.Second, cfg.Timeout)
	})
}

func TestValidate(t *testing.T) {
	t.Run("ValidWithAPIToken", func(t *testing.T) {
		cfg := &Config{
			MythicURL: "https://mythic.example.com:7443",
			APIToken:  "test-token",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("ValidWithUsernamePassword", func(t *testing.T) {
		cfg := &Config{
			MythicURL: "https://mythic.example.com:7443",
			Username:  "admin",
			Password:  "password",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("MissingURL", func(t *testing.T) {
		cfg := &Config{
			APIToken: "test-token",
		}
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("URLOnly_OK", func(t *testing.T) {
		cfg := &Config{
			MythicURL: "https://mythic.example.com:7443",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("PartialCreds_OK", func(t *testing.T) {
		cfg := &Config{
			MythicURL: "https://mythic.example.com:7443",
			Password:  "password",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
