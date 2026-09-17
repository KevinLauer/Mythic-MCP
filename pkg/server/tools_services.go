package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const servicesQuery = `
query Services {
  payloadtype(where: {deleted: {_eq: false}}) {
    id
    name
    wrapper
    container_running
    supported_os
    note
  }
  c2profile(where: {deleted: {_eq: false}}) {
    id
    name
    running
    is_p2p
    container_running
    description
  }
}
`

type getServicesArgs struct{}

func (s *Server) registerServicesTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_get_services",
		Description: "Inventory installed payload types and C2 profiles: container_running vs C2 running (internal listener). " +
			"Use this instead of SSH or custom GraphQL.",
	}, s.handleGetServices)
	s.trackTool("mythic_get_services")
}

func (s *Server) handleGetServices(ctx context.Context, req *mcp.CallToolRequest, args getServicesArgs) (*mcp.CallToolResult, any, error) {
	data, err := s.mythicClient.ExecuteRawGraphQL(ctx, servicesQuery, nil)
	if err != nil {
		return nil, nil, translateError(err)
	}

	text, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Installed services:\n\n%s", string(text))},
		},
	}, data, nil
}
