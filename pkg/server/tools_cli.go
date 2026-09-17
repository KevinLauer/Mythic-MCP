package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type cliStatusArgs struct{}

type cliVersionArgs struct{}

type cliLogsArgs struct {
	Name  string `json:"name" jsonschema:"Installed service or core container name"`
	Lines int    `json:"lines,omitempty" jsonschema:"Optional tail length passed as mythic-cli logs -l"`
}

type cliServiceArgs struct {
	Name string `json:"name" jsonschema:"Installed service or core container name (lowercase)"`
}

func (s *Server) registerCLITools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_status",
		Description: "Run mythic-cli status when MYTHIC_CLI_PATH and MYTHIC_HOME are set. Otherwise returns commands for a human. Never SSH.",
	}, s.handleCLIStatus)
	s.trackTool("mythic_cli_status")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_health",
		Description: "Run mythic-cli health <name>. Requires a container name. CLI env must be set.",
	}, s.handleCLIHealth)
	s.trackTool("mythic_cli_health")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_logs",
		Description: "Run mythic-cli logs <name> when CLI env is set. Optional lines. Never follow (would hang).",
	}, s.handleCLILogs)
	s.trackTool("mythic_cli_logs")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_start",
		Description: "Run mythic-cli start <name> when CLI env is set. Name is required so this does not start every container.",
	}, s.handleCLIStart)
	s.trackTool("mythic_cli_start")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_stop",
		Description: "Run mythic-cli stop <name> when CLI env is set. Name is required so this does not stop all of Mythic.",
	}, s.handleCLIStop)
	s.trackTool("mythic_cli_stop")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_restart",
		Description: "Run mythic-cli restart <name>. Name is required. Restart with no name would bounce every Mythic container.",
	}, s.handleCLIRestart)
	s.trackTool("mythic_cli_restart")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_build",
		Description: "Run mythic-cli build <name> when CLI env is set.",
	}, s.handleCLIBuild)
	s.trackTool("mythic_cli_build")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_version",
		Description: "Run mythic-cli version (CLI, server VERSION file, React UI).",
	}, s.handleCLIVersion)
	s.trackTool("mythic_cli_version")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_cli_service",
		Description: "Install, uninstall, or update a payload type or C2 profile via mythic-cli. " +
			"action=uninstall stops and uninstalls. action=install only installs. " +
			"action=update is uninstall then install. Requires source_dir or github_url for install and update. " +
			"Uses mythic-cli only when MYTHIC_CLI_PATH and MYTHIC_HOME are set.",
	}, s.handleCLIService)
	s.trackTool("mythic_cli_service")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_cli_config",
		Description: "mythic-cli config. subcommand=get|set|help|service|show. " +
			"get/help need keys. set needs key+value. show dumps the full .env including passwords; prefer get.",
	}, s.handleCLIConfig)
	s.trackTool("mythic_cli_config")
}

func (s *Server) cliConfigured() bool {
	return s.config != nil && s.config.CLIPath != "" && s.config.MythicHome != ""
}

func cliUnavailableResult(commands []string) (*mcp.CallToolResult, any, error) {
	payload := map[string]any{
		"ok":       false,
		"reason":   "cli_unavailable",
		"commands": commands,
	}
	text, _ := json.MarshalIndent(payload, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}, payload, nil
}

func (s *Server) runMythicCLI(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, s.config.CLIPath, args...)
	cmd.Dir = s.config.MythicHome
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (s *Server) runCLIOrUnavailable(ctx context.Context, display string, args ...string) (*mcp.CallToolResult, any, error) {
	if !s.cliConfigured() {
		return cliUnavailableResult([]string{display})
	}
	return s.cliResult(ctx, display, args...)
}

func (s *Server) handleCLIStatus(ctx context.Context, req *mcp.CallToolRequest, args cliStatusArgs) (*mcp.CallToolResult, any, error) {
	return s.runCLIOrUnavailable(ctx, "./mythic-cli status", "status")
}

func (s *Server) handleCLIVersion(ctx context.Context, req *mcp.CallToolRequest, args cliVersionArgs) (*mcp.CallToolResult, any, error) {
	return s.runCLIOrUnavailable(ctx, "./mythic-cli version", "version")
}

func (s *Server) handleCLILogs(ctx context.Context, req *mcp.CallToolRequest, args cliLogsArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	cliArgs := []string{"logs", args.Name}
	if args.Lines > 0 {
		cliArgs = append(cliArgs, "-l", strconv.Itoa(args.Lines))
	}
	return s.runCLIOrUnavailable(ctx, "./mythic-cli "+strings.Join(cliArgs, " "), cliArgs...)
}

func (s *Server) handleCLIStart(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	return s.runCLIOrUnavailable(ctx, "./mythic-cli start "+args.Name, "start", args.Name)
}

func (s *Server) handleCLIStop(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	return s.runCLIOrUnavailable(ctx, "./mythic-cli stop "+args.Name, "stop", args.Name)
}

func (s *Server) handleCLIRestart(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required; restart without a name bounces all of Mythic")
	}
	return s.runCLIOrUnavailable(ctx, "./mythic-cli restart "+args.Name, "restart", args.Name)
}

func (s *Server) handleCLIHealth(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	return s.runCLIOrUnavailable(ctx, "./mythic-cli health "+args.Name, "health", args.Name)
}

func (s *Server) handleCLIBuild(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	return s.runCLIOrUnavailable(ctx, "./mythic-cli build "+args.Name, "build", args.Name)
}

func (s *Server) cliResult(ctx context.Context, display string, args ...string) (*mcp.CallToolResult, any, error) {
	stdout, stderr, err := s.runMythicCLI(ctx, args...)
	payload := map[string]any{
		"ok":      err == nil,
		"command": display,
		"stdout":  stdout,
		"stderr":  stderr,
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	text, _ := json.MarshalIndent(payload, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}, payload, nil
}
