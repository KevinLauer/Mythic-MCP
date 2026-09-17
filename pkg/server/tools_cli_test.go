package server

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCLITestServer(t *testing.T) *Server {
	t.Helper()
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		APIToken:  "test-token",
		SSL:       true,
	}
	srv, err := NewServer(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })
	return srv
}

func TestCLIServiceUpdateUnavailable(t *testing.T) {
	srv := newCLITestServer(t)

	result, data, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Action:    "update",
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
	assert.Equal(t, []string{
		"./mythic-cli stop pico",
		"./mythic-cli uninstall pico",
		"./mythic-cli install folder /abs/path/to/agent -f",
	}, commands)
}

func TestCLIServiceInstallUnavailable(t *testing.T) {
	srv := newCLITestServer(t)

	_, data, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Action:    "install",
		GithubURL: "https://github.com/MythicAgents/poseidon",
		Branch:    "Mythic-v4.0.0",
		Name:      "poseidon",
	})
	require.NoError(t, err)
	payload := data.(map[string]any)
	commands, _ := payload["commands"].([]string)
	assert.Equal(t, []string{
		"./mythic-cli install github https://github.com/MythicAgents/poseidon -b Mythic-v4.0.0 -f",
	}, commands)
	assert.NotContains(t, commands, "./mythic-cli stop poseidon")
	assert.NotContains(t, commands, "./mythic-cli uninstall poseidon")
}

func TestCLIServiceUninstallUnavailable(t *testing.T) {
	srv := newCLITestServer(t)

	_, data, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Action: "uninstall",
		Name:   "pico",
	})
	require.NoError(t, err)
	payload := data.(map[string]any)
	commands, _ := payload["commands"].([]string)
	assert.Equal(t, []string{
		"./mythic-cli stop pico",
		"./mythic-cli uninstall pico",
	}, commands)
}

func TestCLIServiceRequiresAction(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Name:      "pico",
		SourceDir: "/abs/path/to/agent",
	})
	require.Error(t, err)
}

func TestCLIServiceInstallRequiresSource(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Action: "install",
		Name:   "pico",
	})
	require.Error(t, err)
}

func TestCLIServiceUninstallRequiresName(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Action: "uninstall",
	})
	require.Error(t, err)
}

func TestCLIServiceUpdateRequiresName(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Action:    "update",
		SourceDir: "/abs/path/to/agent",
	})
	require.Error(t, err)
}

func TestCLIServiceRejectsRelativeSource(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIService(context.Background(), &mcp.CallToolRequest{}, cliServiceLifecycleArgs{
		Action:    "install",
		SourceDir: "relative/path",
	})
	require.Error(t, err)
}

func TestCLIHealthUnavailable(t *testing.T) {
	srv := newCLITestServer(t)
	_, data, err := srv.handleCLIHealth(context.Background(), &mcp.CallToolRequest{}, cliServiceArgs{Name: "mythic_nginx"})
	require.NoError(t, err)
	payload := data.(map[string]any)
	assert.Equal(t, "cli_unavailable", payload["reason"])
	assert.Equal(t, []string{"./mythic-cli health mythic_nginx"}, payload["commands"])
}

func TestCLIHealthRequiresName(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIHealth(context.Background(), &mcp.CallToolRequest{}, cliServiceArgs{})
	require.Error(t, err)
}

func TestCLIRestartRequiresName(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIRestart(context.Background(), &mcp.CallToolRequest{}, cliServiceArgs{})
	require.Error(t, err)
}

func TestCLIRestartUnavailable(t *testing.T) {
	srv := newCLITestServer(t)
	_, data, err := srv.handleCLIRestart(context.Background(), &mcp.CallToolRequest{}, cliServiceArgs{Name: "poseidon"})
	require.NoError(t, err)
	payload := data.(map[string]any)
	assert.Equal(t, []string{"./mythic-cli restart poseidon"}, payload["commands"])
}

func TestCLIVersionUnavailable(t *testing.T) {
	srv := newCLITestServer(t)
	_, data, err := srv.handleCLIVersion(context.Background(), &mcp.CallToolRequest{}, cliVersionArgs{})
	require.NoError(t, err)
	payload := data.(map[string]any)
	assert.Equal(t, []string{"./mythic-cli version"}, payload["commands"])
}

func TestCLIConfigGetRequiresKeys(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIConfig(context.Background(), &mcp.CallToolRequest{}, cliConfigArgs{Subcommand: "get"})
	require.Error(t, err)
}

func TestCLIConfigGetUnavailable(t *testing.T) {
	srv := newCLITestServer(t)
	_, data, err := srv.handleCLIConfig(context.Background(), &mcp.CallToolRequest{}, cliConfigArgs{
		Subcommand: "get",
		Keys:       []string{"MYTHIC_SERVER_PORT"},
	})
	require.NoError(t, err)
	payload := data.(map[string]any)
	assert.Equal(t, []string{"./mythic-cli config get MYTHIC_SERVER_PORT"}, payload["commands"])
}

func TestCLIConfigSetRequiresKeyValue(t *testing.T) {
	srv := newCLITestServer(t)
	_, _, err := srv.handleCLIConfig(context.Background(), &mcp.CallToolRequest{}, cliConfigArgs{
		Subcommand: "set",
		Key:        "DATE_FORMAT",
	})
	require.Error(t, err)
}
