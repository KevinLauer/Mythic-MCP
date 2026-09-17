package server

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type cliServiceLifecycleArgs struct {
	Action    string `json:"action" jsonschema:"uninstall, install, or update. update is uninstall then install. Do not default to update."`
	Name      string `json:"name,omitempty" jsonschema:"Installed service name (lowercase). Required for uninstall and update. Used to poll after install."`
	SourceDir string `json:"source_dir,omitempty" jsonschema:"Absolute path for mythic-cli install folder"`
	GithubURL string `json:"github_url,omitempty" jsonschema:"GitHub URL for mythic-cli install github"`
	Branch    string `json:"branch,omitempty" jsonschema:"Optional git branch (Mythic-v4.0.0 for public v4 agents)"`
	Timeout   int    `json:"timeout,omitempty" jsonschema:"Seconds to wait for container_running after install or update (default 180)"`
}

func (s *Server) handleCLIService(ctx context.Context, req *mcp.CallToolRequest, args cliServiceLifecycleArgs) (*mcp.CallToolResult, any, error) {
	action := strings.ToLower(strings.TrimSpace(args.Action))
	if err := validateCLIServiceArgs(action, args); err != nil {
		return nil, nil, err
	}

	commands := cliServiceCommands(action, args)
	if !s.cliConfigured() {
		return cliUnavailableResult(commands)
	}

	steps := cliServiceSteps(action, args)
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
			if len(step) > 0 && step[0] == "stop" {
				continue
			}
			payload := map[string]any{"ok": false, "action": action, "steps": outputs}
			text, _ := json.MarshalIndent(payload, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
			}, payload, nil
		}
		outputs = append(outputs, entry)
	}

	payload := map[string]any{
		"ok":     true,
		"action": action,
		"name":   args.Name,
		"steps":  outputs,
	}

	if action == "install" || action == "update" {
		if args.Name != "" {
			running, err := s.waitContainerRunning(ctx, args.Name, args.Timeout)
			if err != nil {
				return nil, nil, err
			}
			payload["container_running"] = running
			payload["polled_services"] = true
			if !running {
				payload["ok"] = false
				payload["error"] = "install finished but container_running is still false"
			}
		}
	}

	text, _ := json.MarshalIndent(payload, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}, payload, nil
}

func validateCLIServiceArgs(action string, args cliServiceLifecycleArgs) error {
	switch action {
	case "uninstall":
		if args.Name == "" {
			return fmt.Errorf("name is required for uninstall")
		}
	case "install":
		if err := requireInstallSource(args); err != nil {
			return err
		}
	case "update":
		if args.Name == "" {
			return fmt.Errorf("name is required for update")
		}
		if err := requireInstallSource(args); err != nil {
			return err
		}
	default:
		return fmt.Errorf("action must be uninstall, install, or update")
	}
	return nil
}

func requireInstallSource(args cliServiceLifecycleArgs) error {
	if args.SourceDir == "" && args.GithubURL == "" {
		return fmt.Errorf("source_dir or github_url is required")
	}
	if args.SourceDir != "" && !filepath.IsAbs(args.SourceDir) {
		return fmt.Errorf("source_dir must be an absolute path")
	}
	return nil
}

func installCLIArgs(args cliServiceLifecycleArgs) []string {
	if args.SourceDir != "" {
		return []string{"install", "folder", args.SourceDir, "-f"}
	}
	install := []string{"install", "github", args.GithubURL}
	if args.Branch != "" {
		install = append(install, "-b", args.Branch)
	}
	return append(install, "-f")
}

func cliServiceSteps(action string, args cliServiceLifecycleArgs) [][]string {
	switch action {
	case "uninstall":
		return [][]string{
			{"stop", args.Name},
			{"uninstall", args.Name},
		}
	case "install":
		return [][]string{installCLIArgs(args)}
	default:
		return [][]string{
			{"stop", args.Name},
			{"uninstall", args.Name},
			installCLIArgs(args),
		}
	}
}

func cliServiceCommands(action string, args cliServiceLifecycleArgs) []string {
	var commands []string
	for _, step := range cliServiceSteps(action, args) {
		commands = append(commands, "./mythic-cli "+strings.Join(step, " "))
	}
	return commands
}

func (s *Server) waitContainerRunning(ctx context.Context, name string, timeoutSec int) (bool, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 180 * time.Second
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if s.serviceContainerRunning(ctx, name) {
			return true, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return false, nil
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
