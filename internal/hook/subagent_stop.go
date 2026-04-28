package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
func (h *subagentStopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("subagent stopped",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"team_name", input.TeamName,
		"teammate_name", input.TeammateName,
	)

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

	// Kill the tmux pane
	if err := h.killTmuxPane(tmuxPaneID); err != nil {
		// If pane not found, treat as success and proceed with config cleanup
		if strings.Contains(err.Error(), "pane not found") || strings.Contains(err.Error(), "can't find pane") {
			slog.Debug("tmux pane not found (may already be removed)", "pane_id", tmuxPaneID)
		} else {
			slog.Warn("failed to kill tmux pane", "pane_id", tmuxPaneID, "error", err)
			// Continue with config cleanup despite kill failure
		}
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
	Name      string `json:"name"`
	AgentID   string `json:"agentId,omitempty"`
	AgentType string `json:"agentType,omitempty"`
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
