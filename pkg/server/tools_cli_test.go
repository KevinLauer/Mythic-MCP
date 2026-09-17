package server

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceInstalledServiceUnavailable(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		APIToken:  "test-token",
		SSL:       true,
	}
	srv, err := NewServer(cfg)
	require.NoError(t, err)
	defer srv.Close()

	result, data, err := srv.handleReplaceInstalledService(context.Background(), &mcp.CallToolRequest{}, replaceInstalledServiceArgs{
		Name:      "pico",
		SourceDir: "/abs/path/to/agent",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	payload, ok := data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, false, payload["ok"])
	assert.Equal(t, "cli_unavailable", payload["reason"])
	commands, _ := payload["commands"].([]string)
	assert.Contains(t, commands, "./mythic-cli stop pico")
	assert.Contains(t, commands, "./mythic-cli uninstall pico")
	assert.Contains(t, commands, "./mythic-cli install folder /abs/path/to/agent -f")
}

func TestReplaceInstalledServiceRequiresSource(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		APIToken:  "test-token",
		SSL:       true,
	}
	srv, err := NewServer(cfg)
	require.NoError(t, err)
	defer srv.Close()

	_, _, err = srv.handleReplaceInstalledService(context.Background(), &mcp.CallToolRequest{}, replaceInstalledServiceArgs{
		Name: "pico",
	})
	require.Error(t, err)
}
