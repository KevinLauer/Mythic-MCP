package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerOperatorProfile() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_login",
		Description: "Authenticate with Mythic using username and password when no API token is configured",
	}, s.handleLogin)
	s.trackTool("mythic_login")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_current_user",
		Description: "Get the current authenticated user and current_operation_id",
	}, s.handleGetCurrentUser)
	s.trackTool("mythic_get_current_user")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_current_operation",
		Description: "Get the currently selected Mythic operation id",
	}, s.handleGetCurrentOperation)
	s.trackTool("mythic_get_current_operation")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_set_current_operation",
		Description: "Select the current Mythic operation. Required when current_operation_id is 0",
	}, s.handleSetCurrentOperation)
	s.trackTool("mythic_set_current_operation")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_event_log",
		Description: "Read the operation event feed (callbacks, builds, warnings, UUID correlation errors)",
	}, s.handleGetEventLog)
	s.trackTool("mythic_get_event_log")

	s.registerServicesTools()
	s.registerCLITools()

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_active_callbacks",
		Description: "List active callbacks. Use display_id for tasking",
	}, s.handleGetActiveCallbacks)
	s.trackTool("mythic_get_active_callbacks")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_all_callbacks",
		Description: "List all callbacks including inactive",
	}, s.handleGetAllCallbacks)
	s.trackTool("mythic_get_all_callbacks")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_callback",
		Description: "Get one callback by display_id",
	}, s.handleGetCallback)
	s.trackTool("mythic_get_callback")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_loaded_commands",
		Description: "List commands loaded on a callback (display_id)",
	}, s.handleGetLoadedCommands)
	s.trackTool("mythic_get_loaded_commands")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_issue_task",
		Description: "Issue a command to a callback by display_id, then use wait/output tools",
	}, s.handleIssueTask)
	s.trackTool("mythic_issue_task")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_wait_for_task",
		Description: "Wait until a task display_id completes or errors",
	}, s.handleWaitForTask)
	s.trackTool("mythic_wait_for_task")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_task_output",
		Description: "Get decoded-ready responses for a task display_id",
	}, s.handleGetTaskOutput)
	s.trackTool("mythic_get_task_output")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_callback_tasks",
		Description: "List tasks for a callback display_id",
	}, s.handleGetCallbackTasks)
	s.trackTool("mythic_get_callback_tasks")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payload_types",
		Description: "List installed payload types",
	}, s.handleGetPayloadTypes)
	s.trackTool("mythic_get_payload_types")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payload_type_build_parameters",
		Description: "Build-parameter schema for a payload type id",
	}, s.handleGetPayloadTypeBuildParameters)
	s.trackTool("mythic_get_payload_type_build_parameters")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_c2_profile_parameters",
		Description: "C2 parameter schema for a profile id. Query before create_payload",
	}, s.handleGetC2ProfileParameters)
	s.trackTool("mythic_get_c2_profile_parameters")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_command_with_parameters",
		Description: "Parameter schema for one command on a payload type",
	}, s.handleGetCommandWithParameters)
	s.trackTool("mythic_get_command_with_parameters")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_payload",
		Description: "Build a payload. Use discovered parameter names only",
	}, s.handleCreatePayload)
	s.trackTool("mythic_create_payload")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_wait_for_payload",
		Description: "Wait for payload build_phase to leave building",
	}, s.handleWaitForPayload)
	s.trackTool("mythic_wait_for_payload")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_download_payload",
		Description: "One-time download URL for a built payload uuid",
	}, s.handleDownloadPayload)
	s.trackTool("mythic_download_payload")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payloads",
		Description: "List built payloads",
	}, s.handleGetPayloads)
	s.trackTool("mythic_get_payloads")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_c2_profiles",
		Description: "List C2 profiles. running is the internal listener; container_running is Docker",
	}, s.handleGetC2Profiles)
	s.trackTool("mythic_get_c2_profiles")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_start_c2_profile",
		Description: "Start a C2 internal listener by profile id",
	}, s.handleStartC2Profile)
	s.trackTool("mythic_start_c2_profile")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_stop_c2_profile",
		Description: "Stop a C2 internal listener by profile id",
	}, s.handleStopC2Profile)
	s.trackTool("mythic_stop_c2_profile")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_c2_profile_output",
		Description: "C2 container stdout/stderr",
	}, s.handleGetC2ProfileOutput)
	s.trackTool("mythic_get_c2_profile_output")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_upload_file",
		Description: "Upload a file for later tasking (base64)",
	}, s.handleUploadFile)
	s.trackTool("mythic_upload_file")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_download_file",
		Description: "One-time download URL for a Mythic file uuid",
	}, s.handleDownloadFile)
	s.trackTool("mythic_download_file")
}
