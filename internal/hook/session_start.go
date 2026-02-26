package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// sessionStartHandler processes SessionStart events.
// It initializes the session, loads project configuration, and validates
// the execution environment (REQ-HOOK-030).
type sessionStartHandler struct {
	cfg ConfigProvider
}

// NewSessionStartHandler creates a new SessionStart event handler.
func NewSessionStartHandler(cfg ConfigProvider) Handler {
	return &sessionStartHandler{cfg: cfg}
}

// EventType returns EventSessionStart.
func (h *sessionStartHandler) EventType() EventType {
	return EventSessionStart
}

// Handle processes a SessionStart event. It logs the session ID, loads
// project configuration, and returns project information in the Data field.
// If memory injection is enabled, MEMORY.md content is set as SystemMessage.
// Errors are non-blocking: the handler logs warnings and returns allow.
func (h *sessionStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session started",
		"session_id", input.SessionID,
		"cwd", input.CWD,
		"project_dir", input.ProjectDir,
	)

	data := map[string]any{
		"session_id": input.SessionID,
		"status":     "initialized",
	}

	// Load project information from config if available
	cfg := h.getConfig()
	if cfg != nil {
		if cfg.Project.Name != "" {
			data["project_name"] = cfg.Project.Name
		}
		if string(cfg.Project.Type) != "" {
			data["project_type"] = string(cfg.Project.Type)
		}
		if cfg.Project.Language != "" {
			data["project_language"] = cfg.Project.Language
		}
	} else {
		slog.Warn("configuration not available, proceeding with defaults",
			"session_id", input.SessionID,
		)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal session data",
			"error", err.Error(),
		)
		return &HookOutput{}, nil
	}

	output := &HookOutput{Data: jsonData}

	// Load project memory for injection via SystemMessage
	projectDir := input.CWD
	if projectDir == "" {
		projectDir = input.ProjectDir
	}
	memoryContent := h.loadMemory(projectDir)
	if memoryContent != "" {
		output.SystemMessage = memoryContent
	}

	return output, nil
}

// loadMemory reads .moai/memory/MEMORY.md from the project directory.
// Returns empty string if memory is disabled, file does not exist, or on error.
func (h *sessionStartHandler) loadMemory(projectDir string) string {
	cfg := h.getConfig()
	if cfg == nil || !cfg.Memory.Enabled || !cfg.Memory.AutoInject {
		return ""
	}

	memoryDir := cfg.Memory.MemoryDir
	if memoryDir == "" {
		memoryDir = ".moai/memory"
	}

	memoryFile := filepath.Join(projectDir, memoryDir, "MEMORY.md")
	data, err := os.ReadFile(memoryFile)
	if err != nil {
		// File not found is normal (memory not yet created)
		if !os.IsNotExist(err) {
			slog.Warn("session_start: failed to read memory file",
				"path", memoryFile,
				"error", err,
			)
		}
		return ""
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return ""
	}

	// Truncate to max tokens (rough estimate: 4 chars per token)
	maxChars := cfg.Memory.MaxTokens * 4
	if maxChars > 0 && len(content) > maxChars {
		content = content[:maxChars] + "\n[truncated]"
	}

	return content
}

// getConfig safely retrieves the configuration, returning nil if unavailable.
func (h *sessionStartHandler) getConfig() *config.Config {
	if h.cfg == nil {
		return nil
	}
	return h.cfg.Get()
}
