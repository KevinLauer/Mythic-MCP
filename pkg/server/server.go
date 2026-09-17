package server

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/nbaertsch/Mythic-MCP/pkg/filestore"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
)

// Server is the main MCP server wrapping Mythic SDK
type Server struct {
	config       *config.Config
	mcpServer    *mcp.Server
	mythicClient *mythic.Client
	fileStore    *filestore.FileStore
	toolNames    []string
}

// NewServer creates a new MCP server with Mythic integration
func NewServer(cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Create Mythic SDK client configuration
	mythicConfig := &mythic.Config{
		ServerURL:     cfg.MythicURL,
		APIToken:      cfg.APIToken,
		Username:      cfg.Username,
		Password:      cfg.Password,
		SSL:           cfg.SSL,
		SkipTLSVerify: cfg.SkipTLSVerify,
		Timeout:       cfg.Timeout,
	}

	// Create Mythic client
	mythicClient, err := mythic.NewClient(mythicConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mythic client: %w", err)
	}

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "mythic-mcp",
		Version: "1.0.0",
	}, nil)

	server := &Server{
		config:       cfg,
		mcpServer:    mcpServer,
		mythicClient: mythicClient,
	}

	// Initialize file store for file vending
	if cfg.FileVendingEnabled {
		fsCfg := &filestore.Config{
			Enabled:         true,
			StoragePath:     cfg.FileStoragePath,
			TokenExpiry:     cfg.FileTokenExpiry,
			MaxFileSizeMB:   cfg.FileMaxSizeMB,
			CleanupInterval: cfg.FileCleanupInterval,
			BaseURL:         cfg.FileVendingBaseURL,
		}
		fs, err := filestore.New(fsCfg)
		if err != nil {
			log.Printf("Warning: failed to initialize file store: %v (file vending disabled)", err)
		} else {
			server.fileStore = fs
		}
	}

	// Register MCP tools
	server.registerTools()

	return server, nil
}

// Authenticate logs in to the Mythic instance. Call this before serving.
func (s *Server) Authenticate(ctx context.Context) error {
	log.Println("Authenticating with Mythic...")
	if err := s.mythicClient.Login(ctx); err != nil {
		return fmt.Errorf("failed to authenticate with Mythic: %w", err)
	}
	log.Println("Successfully authenticated with Mythic")
	return nil
}

// MCPServer returns the underlying *mcp.Server so callers can wire it into
// any transport (StreamableHTTPHandler, SSEHandler, StdioTransport, etc.).
func (s *Server) MCPServer() *mcp.Server {
	return s.mcpServer
}

// Run starts the MCP server on a single-session transport (e.g. stdio).
// The server starts unauthenticated; the user must call mythic_login to
// establish a session.
// For HTTP serving, use MCPServer() with a StreamableHTTPHandler instead.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	log.Println("Starting MCP server (unauthenticated — call mythic_login to authenticate)...")
	return s.mcpServer.Run(ctx, transport)
}

// Close cleans up server resources
func (s *Server) Close() error {
	if s.fileStore != nil {
		s.fileStore.Close()
	}
	if s.mythicClient != nil {
		return s.mythicClient.Close()
	}
	return nil
}

// FileStore returns the file store instance (may be nil if disabled).
func (s *Server) FileStore() *filestore.FileStore {
	return s.fileStore
}

func (s *Server) ToolNames() []string {
	out := make([]string, len(s.toolNames))
	copy(out, s.toolNames)
	return out
}

func (s *Server) trackTool(name string) {
	s.toolNames = append(s.toolNames, name)
}

// registerTools registers MCP tools. Default profile is operator (slim).
func (s *Server) registerTools() {
	if s.config != nil && s.config.OperatorProfile() {
		s.registerOperatorProfile()
		return
	}
	s.registerFullProfile()
}

func (s *Server) registerFullProfile() {
	s.registerAuthTools()
	s.registerOperationsTools()
	s.registerFilesTools()
	s.registerOperatorsTools()
	s.registerTagsTools()
	s.registerCredentialsTools()
	s.registerArtifactsTools()
	s.registerCallbacksTools()
	s.registerTasksTools()
	s.registerPayloadsTools()
	s.registerPayloadDiscoveryTools()
	s.registerC2ProfilesTools()
	s.registerCommandsTools()
	s.registerAttackTools()
	s.registerProcessesTools()
	s.registerHostsTools()
	s.registerScreenshotsTools()
	s.registerKeylogsTools()
	s.registerDocumentationTools()
	s.registerBulkTools()
	s.registerServicesTools()
	s.registerCLITools()
}
