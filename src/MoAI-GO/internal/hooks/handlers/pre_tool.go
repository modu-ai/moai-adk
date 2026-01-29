package handlers

import (
	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
	"github.com/anthropics/moai-adk-go/internal/hooks/security"
)

// PreToolHandler handles pre-tool events (security guard)
type PreToolHandler struct {
	guard *security.SecurityGuard
}

// NewPreToolHandler creates a new pre-tool handler
func NewPreToolHandler() *PreToolHandler {
	return &PreToolHandler{
		guard: security.NewSecurityGuard(),
	}
}

// Handle executes the pre-tool hook
func (h *PreToolHandler) Handle(input *protocol.HookInput) (*protocol.HookResponse, error) {
	// Only process Write, Edit, and Bash tools
	if input.ToolName != "Write" && input.ToolName != "Edit" && input.ToolName != "Bash" {
		return protocol.NewMessageResponse("", true), nil // Suppress output
	}

	// Handle Bash commands
	if input.ToolName == "Bash" {
		return h.handleBash(input)
	}

	// Handle Write and Edit tools
	return h.handleFileOperation(input)
}

// handleBash validates bash commands
func (h *PreToolHandler) handleBash(input *protocol.HookInput) (*protocol.HookResponse, error) {
	command := input.GetToolInputCommand()
	if command == "" {
		return protocol.NewMessageResponse("", true), nil // Suppress output
	}

	decision, reason := h.guard.ValidateCommand(command)
	switch decision {
	case security.DecisionBlock:
		return protocol.NewSecurityResponse(protocol.DecisionBlock, reason), nil
	case security.DecisionWarn:
		return protocol.NewSecurityResponse(protocol.DecisionWarn, reason), nil
	default:
		return protocol.NewMessageResponse("", true), nil // Suppress output for allow
	}
}

// handleFileOperation validates file paths
func (h *PreToolHandler) handleFileOperation(input *protocol.HookInput) (*protocol.HookResponse, error) {
	filePath := input.GetToolInputPath()
	if filePath == "" {
		return protocol.NewMessageResponse("", true), nil // Suppress output
	}

	decision, reason := h.guard.ValidatePath(filePath)
	switch decision {
	case security.DecisionBlock:
		return protocol.NewSecurityResponse(protocol.DecisionBlock, reason), nil
	case security.DecisionWarn:
		return protocol.NewSecurityResponse(protocol.DecisionWarn, reason), nil
	default:
		return protocol.NewMessageResponse("", true), nil // Suppress output for allow
	}
}
