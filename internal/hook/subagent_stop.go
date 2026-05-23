// Resolution: FIX — P-H02 bug fix: read tmuxPaneId, kill-pane with 500ms timeout, update team registry.
// W3 EXTEND: harness-learner capture pipeline dispatch on SubagentStop (REQ-HRA-001).
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/capture"
)

// subagentStopHandler processes SubagentStop events.
// It cleans up tmux panes when teammates shut down.
type subagentStopHandler struct{}

// NewSubagentStopHandler creates a new SubagentStop event handler.
func NewSubagentStopHandler() Handler {
	return &subagentStopHandler{}
}

// EventType returns EventSubagentStop.
func (h *subagentStopHandler) EventType() EventType {
	return EventSubagentStop
}

// Handle processes a SubagentStop event. It reads team config to find
// the teammate's tmux pane ID, kills the pane, and updates the config.
// W3 (REQ-HRA-001): also dispatches to harness-learner capture pipeline.
func (h *subagentStopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("subagent stopped",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"team_name", input.TeamName,
		"teammate_name", input.TeammateName,
	)

	// W3 — Harness-learner capture pipeline dispatch (REQ-HRA-001).
	// Non-blocking: capture errors are logged but do not affect hook outcome.
	// [HARD] This package does not call AskUserQuestion. Errors emit to slog only.
	h.dispatchCapture(input)

	// On Windows, tmux is not available - no-op
	if runtime.GOOS == "windows" {
		slog.Debug("tmux cleanup skipped on Windows")
		return &HookOutput{}, nil
	}

	// Get team config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("failed to get home directory", "error", err)
		return &HookOutput{}, nil
	}

	teamConfigPath := fmt.Sprintf("%s/.claude/teams/%s/config.json", homeDir, input.TeamName)

	// Read team config to find tmuxPaneId
	tmuxPaneID, err := h.findTeammatePaneID(teamConfigPath, input.TeammateName)
	if err != nil {
		slog.Debug("failed to find tmux pane ID", "error", err)
		// Pane may have been removed concurrently - treat as success
		return &HookOutput{}, nil
	}

	// Kill the tmux pane with 500ms timeout per-pane (SPEC-V3R2-RT-006 AC-15, AC-02).
	// The goroutine + context prevents pane cleanup from exceeding SessionEnd 1500ms ceiling.
	killDone := make(chan error, 1)
	go func() {
		killDone <- h.killTmuxPane(tmuxPaneID)
	}()

	const killTimeout = 500 * time.Millisecond
	select {
	case err := <-killDone:
		if err != nil {
			// If pane already gone, treat as successful cleanup (AC-02, REQ-061).
			if strings.Contains(err.Error(), "pane not found") || strings.Contains(err.Error(), "can't find pane") {
				slog.Debug("tmux pane not found (already removed)", "pane_id", tmuxPaneID)
			} else {
				slog.Warn("failed to kill tmux pane", "pane_id", tmuxPaneID, "error", err)
				// Continue with config cleanup despite kill failure.
			}
		}
	case <-time.After(killTimeout):
		slog.Warn("tmux kill-pane timed out", "pane_id", tmuxPaneID, "timeout", killTimeout)
		// Best-effort: proceed with config cleanup even when kill-pane times out.
	}

	// Update team config to remove teammate entry
	if err := h.removeTeammateFromConfig(teamConfigPath, input.TeammateName); err != nil {
		slog.Warn("failed to update team config", "error", err)
		// Don't fail the hook for config update issues
	}

	msg := fmt.Sprintf("Teammate %s shut down, pane %s released", input.TeammateName, tmuxPaneID)
	return &HookOutput{
		SystemMessage: msg,
	}, nil
}

// teamConfigFile represents the structure of ~/.claude/teams/{team-name}/config.json
type teamConfigFile struct {
	Name    string         `json:"name"`
	Members []teamMemberDb `json:"members"`
}

// teamMemberDb represents a teammate in the team config
type teamMemberDb struct {
	Name       string `json:"name"`
	AgentID    string `json:"agentId,omitempty"`
	AgentType  string `json:"agentType,omitempty"`
	TmuxPaneID string `json:"tmuxPaneId,omitempty"`
}

// findTeammatePaneID reads the team config and finds the tmuxPaneId for the given teammate name
func (h *subagentStopHandler) findTeammatePaneID(configPath, teammateName string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("read team config: %w", err)
	}

	var config teamConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("parse team config: %w", err)
	}

	for _, member := range config.Members {
		if member.Name == teammateName {
			if member.TmuxPaneID == "" {
				return "", fmt.Errorf("teammate %s has no tmuxPaneId", teammateName)
			}
			return member.TmuxPaneID, nil
		}
	}

	return "", fmt.Errorf("teammate %s not found in config", teammateName)
}

// killTmuxPane executes tmux kill-pane -t <pane-id>
func (h *subagentStopHandler) killTmuxPane(paneID string) error {
	cmd := exec.Command("tmux", "kill-pane", "-t", paneID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux kill-pane failed: %w, output: %s", err, string(output))
	}
	return nil
}

// removeTeammateFromConfig updates the team config to remove the specified teammate
func (h *subagentStopHandler) removeTeammateFromConfig(configPath, teammateName string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read team config: %w", err)
	}

	var config teamConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse team config: %w", err)
	}

	// Filter out the teammate
	var updatedMembers []teamMemberDb
	found := false
	for _, member := range config.Members {
		if member.Name != teammateName {
			updatedMembers = append(updatedMembers, member)
		} else {
			found = true
		}
	}

	if !found {
		// Teammate already removed - treat as success
		return nil
	}

	config.Members = updatedMembers

	// Write updated config
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// dispatchCapture emits an observation to the harness-learner capture pipeline (REQ-HRA-001).
// Non-blocking: errors are logged to slog and ignored (capture is best-effort).
// [HARD] No AskUserQuestion call — this package is a subagent-level hook handler.
//
// Path resolution: delegates to resolveProjectRoot to prefer CLAUDE_PROJECT_DIR
// over input.CWD and guard against writes outside a valid MoAI project root.
// Skips capture entirely when no .moai/ exists in the resolved root (prevents
// the historical internal/hook/.moai/ leak when tests called Handle without CWD).
func (h *subagentStopHandler) dispatchCapture(input *HookInput) {
	projectDir := resolveProjectRoot(input)
	if projectDir == "" {
		return
	}
	obsPath := filepath.Join(projectDir, ".moai", "harness", "observations.yaml")

	capturer := capture.New(capture.Config{ObservationsPath: obsPath})
	event := capture.SubagentStopEvent{
		AgentName:   input.AgentName,
		AgentType:   input.AgentType,
		SessionID:   input.SessionID,
		Timestamp:   time.Now().UTC(),
		ContextHash: input.AgentID, // AgentID as session context identifier
	}
	if event.AgentName == "" {
		// Fallback: use TeammateName when AgentName is absent (Agent Teams context).
		event.AgentName = input.TeammateName
	}
	if err := capturer.OnSubagentStop(event); err != nil {
		slog.Debug("harness capture dispatch failed (non-fatal)",
			"error", err,
			"agent_name", event.AgentName,
		)
	}
}
