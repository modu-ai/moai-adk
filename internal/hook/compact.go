package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/hook/memo"
)

// compactHandler processes PreCompact events.
// It captures context information and creates session state snapshots
// for post-compaction recovery (REQ-HOOK-036). Always returns "allow".
type compactHandler struct{}

// NewCompactHandler creates a new PreCompact event handler.
func NewCompactHandler() Handler {
	return &compactHandler{}
}

// EventType returns EventPreCompact.
func (h *compactHandler) EventType() EventType {
	return EventPreCompact
}

// Handle processes a PreCompact event. It captures the current context,
// writes a session memo for post-compaction recovery, and returns preservation
// status in the Data field. Errors are non-blocking.
func (h *compactHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	projectDir := resolveProjectDir(input)

	slog.Info("pre-compact context preservation",
		"session_id", input.SessionID,
		"project_dir", projectDir,
	)

	// Build memo sections from available hook input.
	sections := buildCompactSections(input, projectDir)

	// Write session memo (non-blocking: log and continue on error).
	if projectDir != "" {
		if err := memo.Write(projectDir, sections); err != nil {
			slog.Warn("pre-compact: failed to write session memo",
				"error", err,
				"project_dir", projectDir,
			)
		} else {
			slog.Info("pre-compact: session memo written",
				"path", filepath.Join(projectDir, ".moai/state/session-memo.md"),
			)
		}
	}

	data := map[string]any{
		"session_id":       input.SessionID,
		"status":           "preserved",
		"snapshot_created": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal compact data",
			"error", err.Error(),
		)
		return &HookOutput{}, nil
	}

	return &HookOutput{Data: jsonData}, nil
}

// resolveProjectDir returns the best available project directory from the input.
func resolveProjectDir(input *HookInput) string {
	if input.CWD != "" {
		return input.CWD
	}
	return input.ProjectDir
}

// buildCompactSections assembles the memo sections to be written before compaction.
func buildCompactSections(input *HookInput, projectDir string) []memo.Section {
	var sections []memo.Section

	// P1: session identity and event context.
	p1Content := buildP1Content(input)
	if p1Content != "" {
		sections = append(sections, memo.Section{
			Priority: memo.P1Required,
			Title:    "Session Context",
			Content:  p1Content,
			Budget:   200,
		})
	}

	// P1: persistent mode status (read from .moai/state/persistent-mode.json).
	if projectDir != "" {
		if modeContent := readPersistentMode(projectDir); modeContent != "" {
			sections = append(sections, memo.Section{
				Priority: memo.P1Required,
				Title:    "Execution Mode",
				Content:  modeContent,
				Budget:   200,
			})
		}
	}

	// P2: active worktrees (read from .moai/state/worktrees.json).
	if projectDir != "" {
		if wtContent := readWorktrees(projectDir); wtContent != "" {
			sections = append(sections, memo.Section{
				Priority: memo.P2High,
				Title:    "Active Worktrees",
				Content:  wtContent,
				Budget:   500,
			})
		}
	}

	return sections
}

// buildP1Content assembles the P1 session context text.
func buildP1Content(input *HookInput) string {
	if input.SessionID == "" && input.CWD == "" && input.HookEventName == "" {
		return ""
	}
	content := ""
	if input.SessionID != "" {
		content += "session_id: " + input.SessionID + "\n"
	}
	if input.CWD != "" {
		content += "cwd: " + input.CWD + "\n"
	}
	if input.HookEventName != "" {
		content += "event: " + input.HookEventName + "\n"
	}
	return content
}

// readPersistentMode reads .moai/state/persistent-mode.json and returns
// a human-readable summary, or empty string if the file does not exist.
func readPersistentMode(projectDir string) string {
	path := filepath.Join(projectDir, ".moai", "state", "persistent-mode.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return string(data)
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return string(data)
	}
	return string(out)
}

// readWorktrees reads .moai/state/worktrees.json and returns a human-readable
// summary, or empty string if the file does not exist.
func readWorktrees(projectDir string) string {
	path := filepath.Join(projectDir, ".moai", "state", "worktrees.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var m any
	if err := json.Unmarshal(data, &m); err != nil {
		return string(data)
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return string(data)
	}
	return string(out)
}
