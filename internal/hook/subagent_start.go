package hook

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// subagentStartHandler processes SubagentStart events.
// It logs subagent startup for session tracking and optionally injects project context.
type subagentStartHandler struct {
	cfg        ConfigProvider
	projectDir string
}

// NewSubagentStartHandler creates a new SubagentStart event handler without config.
// This constructor preserves backward compatibility; no additionalContext is injected.
func NewSubagentStartHandler() Handler {
	return &subagentStartHandler{}
}

// NewSubagentStartHandlerWithConfig creates a new SubagentStart event handler with
// config access for project context injection.
// projectDir is resolved from CLAUDE_PROJECT_DIR env var or os.Getwd() as fallback.
func NewSubagentStartHandlerWithConfig(cfg ConfigProvider) Handler {
	dir := os.Getenv("CLAUDE_PROJECT_DIR")
	if dir == "" {
		if cwd, err := os.Getwd(); err == nil {
			dir = cwd
		}
	}
	return &subagentStartHandler{cfg: cfg, projectDir: dir}
}

// EventType returns EventSubagentStart.
func (h *subagentStartHandler) EventType() EventType {
	return EventSubagentStart
}

// Handle processes a SubagentStart event.
// If config is available, it injects project context via additionalContext.
// Errors are non-blocking; an empty output is returned on failure.
func (h *subagentStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("subagent started",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_transcript_path", input.AgentTranscriptPath,
	)

	contextStr := h.buildContext(input)
	if contextStr == "" {
		return &HookOutput{}, nil
	}

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     string(EventSubagentStart),
			AdditionalContext: contextStr,
		},
	}, nil
}

// buildContext constructs a concise project context string (under 200 chars).
// Returns empty string if no useful context is available.
func (h *subagentStartHandler) buildContext(input *HookInput) string {
	if h.cfg == nil {
		return ""
	}

	cfg := h.cfg.Get()
	if cfg == nil {
		return ""
	}

	var parts []string

	if cfg.Project.Name != "" {
		parts = append(parts, "project:"+cfg.Project.Name)
	}
	if string(cfg.Project.Type) != "" {
		parts = append(parts, "type:"+string(cfg.Project.Type))
	}
	if cfg.Project.Language != "" {
		parts = append(parts, "lang:"+cfg.Project.Language)
	}

	// Resolve project directory from input CWD, handler field, or env
	dir := input.CWD
	if dir == "" {
		dir = h.projectDir
	}

	if spec := h.detectActiveSpec(dir); spec != "" {
		parts = append(parts, "spec:"+spec)
	}

	if len(parts) == 0 {
		return ""
	}

	result := strings.Join(parts, " | ")
	// Truncate to stay under 200 chars
	if len(result) > 199 {
		result = result[:199]
	}
	return result
}

// detectActiveSpec returns the SPEC ID of the most recently modified spec.md under dir.
// Returns empty string if no SPEC is found.
func (h *subagentStartHandler) detectActiveSpec(dir string) string {
	if dir == "" {
		return ""
	}

	pattern := filepath.Join(dir, specFilePattern)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}

	var latestMatch string
	var latestModTime int64
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		mt := info.ModTime().UnixNano()
		if mt > latestModTime {
			latestModTime = mt
			latestMatch = match
		}
	}
	if latestMatch == "" {
		return ""
	}

	// Extract SPEC ID from directory name (e.g., "SPEC-FOO-001")
	specDirName := filepath.Base(filepath.Dir(latestMatch))
	return specDirName
}
