package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type cliStatusArgs struct{}

type cliLogsArgs struct {
	Name  string `json:"name" jsonschema:"Installed service or core container name"`
	Lines int    `json:"lines,omitempty" jsonschema:"Optional tail length if the CLI supports it"`
}

type cliServiceArgs struct {
	Name string `json:"name" jsonschema:"Installed service name (lowercase)"`
}

type replaceInstalledServiceArgs struct {
	Name      string `json:"name" jsonschema:"Installed service name to replace (lowercase folder/container name)"`
	SourceDir string `json:"source_dir,omitempty" jsonschema:"Absolute path to a local agent/profile repo for mythic-cli install folder"`
	GithubURL string `json:"github_url,omitempty" jsonschema:"GitHub URL for mythic-cli install github"`
	Branch    string `json:"branch,omitempty" jsonschema:"Optional git branch (Mythic-v4.0.0 for public v4 agents)"`
	Timeout   int    `json:"timeout,omitempty" jsonschema:"Seconds to wait for container_running after install (default 180)"`
}

func (s *Server) registerCLITools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_status",
		Description: "Run mythic-cli status when MYTHIC_CLI_PATH and MYTHIC_HOME are set. Otherwise returns commands for a human. Never SSH.",
	}, s.handleCLIStatus)
	s.trackTool("mythic_cli_status")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_logs",
		Description: "Run mythic-cli logs <name> when CLI env is set. Otherwise returns commands for a human.",
	}, s.handleCLILogs)
	s.trackTool("mythic_cli_logs")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_start",
		Description: "Run mythic-cli start <name> when CLI env is set.",
	}, s.handleCLIStart)
	s.trackTool("mythic_cli_start")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_stop",
		Description: "Run mythic-cli stop <name> when CLI env is set.",
	}, s.handleCLIStop)
	s.trackTool("mythic_cli_stop")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_cli_build",
		Description: "Run mythic-cli build <name> when CLI env is set.",
	}, s.handleCLIBuild)
	s.trackTool("mythic_cli_build")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_replace_installed_service",
		Description: "Stop, uninstall, and reinstall a payload type or C2 profile so the old version is gone. " +
			"Requires source_dir or github_url. Uses mythic-cli only when MYTHIC_CLI_PATH and MYTHIC_HOME are set.",
	}, s.handleReplaceInstalledService)
	s.trackTool("mythic_replace_installed_service")
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

func (s *Server) handleCLIStatus(ctx context.Context, req *mcp.CallToolRequest, args cliStatusArgs) (*mcp.CallToolResult, any, error) {
	commands := []string{"./mythic-cli status"}
	if !s.cliConfigured() {
		return cliUnavailableResult(commands)
	}
	return s.cliResult(ctx, commands[0], "status")
}

func (s *Server) handleCLILogs(ctx context.Context, req *mcp.CallToolRequest, args cliLogsArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	commands := []string{fmt.Sprintf("./mythic-cli logs %s", args.Name)}
	if !s.cliConfigured() {
		return cliUnavailableResult(commands)
	}
	return s.cliResult(ctx, commands[0], "logs", args.Name)
}

func (s *Server) handleCLIStart(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	commands := []string{fmt.Sprintf("./mythic-cli start %s", args.Name)}
	if !s.cliConfigured() {
		return cliUnavailableResult(commands)
	}
	return s.cliResult(ctx, commands[0], "start", args.Name)
}

func (s *Server) handleCLIStop(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	commands := []string{fmt.Sprintf("./mythic-cli stop %s", args.Name)}
	if !s.cliConfigured() {
		return cliUnavailableResult(commands)
	}
	return s.cliResult(ctx, commands[0], "stop", args.Name)
}

func (s *Server) handleCLIBuild(ctx context.Context, req *mcp.CallToolRequest, args cliServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	commands := []string{fmt.Sprintf("./mythic-cli build %s", args.Name)}
	if !s.cliConfigured() {
		return cliUnavailableResult(commands)
	}
	return s.cliResult(ctx, commands[0], "build", args.Name)
}

func (s *Server) handleReplaceInstalledService(ctx context.Context, req *mcp.CallToolRequest, args replaceInstalledServiceArgs) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if args.SourceDir == "" && args.GithubURL == "" {
		return nil, nil, fmt.Errorf("source_dir or github_url is required")
	}
	if args.SourceDir != "" && !filepath.IsAbs(args.SourceDir) {
		return nil, nil, fmt.Errorf("source_dir must be an absolute path")
	}

	commands := replaceCommands(args)
	if !s.cliConfigured() {
		return cliUnavailableResult(commands)
	}

	steps := [][]string{
		{"stop", args.Name},
		{"uninstall", args.Name},
	}
	if args.SourceDir != "" {
		steps = append(steps, []string{"install", "folder", args.SourceDir, "-f"})
	} else {
		install := []string{"install", "github", args.GithubURL}
		if args.Branch != "" {
			install = append(install, "-b", args.Branch)
		}
		install = append(install, "-f")
		steps = append(steps, install)
	}

	outputs := make([]map[string]any, 0, len(steps)+1)
	for _, step := range steps {
		stdout, stderr, err := s.runMythicCLI(ctx, step...)
		entry := map[string]any{
			"args":   step,
			"stdout": stdout,
			"stderr": stderr,
		}
		if err != nil {
			entry["error"] = err.Error()
			outputs = append(outputs, entry)
			payload := map[string]any{"ok": false, "steps": outputs}
			text, _ := json.MarshalIndent(payload, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
			}, payload, nil
		}
		outputs = append(outputs, entry)
	}

	timeout := time.Duration(args.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 180 * time.Second
	}
	deadline := time.Now().Add(timeout)
	running := false
	for time.Now().Before(deadline) {
		running = s.serviceContainerRunning(ctx, args.Name)
		if running {
			break
		}
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	payload := map[string]any{
		"ok":                 running,
		"name":               args.Name,
		"container_running":  running,
		"steps":              outputs,
		"polled_services":    true,
	}
	if !running {
		payload["error"] = "install finished but container_running is still false"
	}
	text, _ := json.MarshalIndent(payload, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}, payload, nil
}

func replaceCommands(args replaceInstalledServiceArgs) []string {
	commands := []string{
		fmt.Sprintf("./mythic-cli stop %s", args.Name),
		fmt.Sprintf("./mythic-cli uninstall %s", args.Name),
	}
	if args.SourceDir != "" {
		commands = append(commands, fmt.Sprintf("./mythic-cli install folder %s -f", args.SourceDir))
		return commands
	}
	cmd := fmt.Sprintf("./mythic-cli install github %s", args.GithubURL)
	if args.Branch != "" {
		cmd += " -b " + args.Branch
	}
	cmd += " -f"
	return append(commands, cmd)
}

func (s *Server) serviceContainerRunning(ctx context.Context, name string) bool {
	data, err := s.mythicClient.ExecuteRawGraphQL(ctx, servicesQuery, nil)
	if err != nil {
		return false
	}
	want := strings.ToLower(name)
	for _, key := range []string{"payloadtype", "c2profile"} {
		rows, _ := data[key].([]interface{})
		for _, row := range rows {
			m, _ := row.(map[string]interface{})
			n, _ := m["name"].(string)
			if strings.ToLower(n) != want {
				continue
			}
			running, _ := m["container_running"].(bool)
			return running
		}
	}
	return false
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
