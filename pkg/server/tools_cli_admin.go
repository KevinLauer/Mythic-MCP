package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type cliConfigArgs struct {
	Subcommand string   `json:"subcommand" jsonschema:"get, set, help, service, or show. show dumps the full .env including passwords; prefer get with explicit keys."`
	Keys       []string `json:"keys,omitempty" jsonschema:"Config key names for get or help"`
	Key        string   `json:"key,omitempty" jsonschema:"Config key for set"`
	Value      string   `json:"value,omitempty" jsonschema:"Config value for set"`
}

func (s *Server) handleCLIConfig(ctx context.Context, req *mcp.CallToolRequest, args cliConfigArgs) (*mcp.CallToolResult, any, error) {
	cliArgs, err := configCLIArgs(args)
	if err != nil {
		return nil, nil, err
	}
	return s.runCLIOrUnavailable(ctx, "./mythic-cli "+strings.Join(cliArgs, " "), cliArgs...)
}

func configCLIArgs(args cliConfigArgs) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(args.Subcommand)) {
	case "show":
		return []string{"config"}, nil
	case "service":
		return []string{"config", "service"}, nil
	case "get":
		if len(args.Keys) == 0 {
			return nil, fmt.Errorf("keys are required for config get")
		}
		return append([]string{"config", "get"}, args.Keys...), nil
	case "help":
		if len(args.Keys) == 0 {
			return nil, fmt.Errorf("keys are required for config help")
		}
		return append([]string{"config", "help"}, args.Keys...), nil
	case "set":
		if args.Key == "" || args.Value == "" {
			return nil, fmt.Errorf("key and value are required for config set")
		}
		return []string{"config", "set", args.Key, args.Value}, nil
	default:
		return nil, fmt.Errorf("subcommand must be get, set, help, service, or show")
	}
}
