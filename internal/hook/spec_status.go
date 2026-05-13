package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// specStatusHandler processes PostToolUse events to auto-update SPEC status on git commits.
// It implements REQ-3 of SPEC-STATUS-AUTO-001.
type specStatusHandler struct{}

// NewSpecStatusHandler creates a new spec status hook handler.
func NewSpecStatusHandler() Handler {
	return &specStatusHandler{}
}

// EventType returns EventPostToolUse.
func (h *specStatusHandler) EventType() EventType {
	return EventPostToolUse
}

// Handle processes a PostToolUse event. It checks if the tool was a git commit or gh pr merge,
// extracts SPEC-IDs from the commit/PR title, and updates their status based on transition rules.
// Errors are non-blocking - the handler always returns success.
func (h *specStatusHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	projectDir := resolveProjectDir(input)

	// Parse hook input data from ToolInput field
	var data struct {
		ToolName string `json:"tool_name"`
		Command  string `json:"command"`
		Title    string `json:"title"` // PR title for gh pr merge
	}
	if len(input.ToolInput) > 0 {
		if err := json.Unmarshal(input.ToolInput, &data); err != nil {
			slog.Debug("spec-status: failed to unmarshal hook data", "error", err)
			return &HookOutput{}, nil
		}
	}

	var targetStatus string
	var specIDs []string

	// Handle gh pr merge commands (Wave 2: AC-02.b)
	if h.isGhPrMergeCommand(data.Command) {
		// Classify PR title to determine target status
		category, status, err := spec.ClassifyPRTitle(data.Title)
		if err != nil {
			slog.Debug("spec-status: failed to classify PR title", "title", data.Title, "error", err)
			return &HookOutput{}, nil
		}

		// Handle special categories
		switch category {
		case "no-op", "skip-meta":
			// Reverts and auto-sync don't change status
			return &HookOutput{}, nil
		case "unknown":
			// Unknown prefix - extract SPEC-IDs but keep current status
			specIDs = spec.ExtractSPECIDs(data.Title)
			if len(specIDs) == 0 {
				return &HookOutput{}, nil
			}
			// Don't update status for unknown prefixes
			slog.Debug("spec-status: unknown prefix, skipping status update", "title", data.Title)
			return h.buildOutput(0, specIDs), nil
		}

		targetStatus = status
		specIDs = spec.ExtractSPECIDs(data.Title)

	// Handle git commit commands (existing behavior: Wave 1)
	} else if isGitCommitCommand(data.Command) {
		commitMsg := extractCommitMessage(data.Command)
		if commitMsg == "" {
			return &HookOutput{}, nil
		}

		targetStatus = "implemented" // Git commits always go to implemented
		specIDs = spec.ExtractSPECIDs(commitMsg)
	} else {
		// Not a command we process
		return &HookOutput{}, nil
	}

	if len(specIDs) == 0 {
		return &HookOutput{}, nil
	}

	// Update each SPEC status with idempotency check
	updatedCount := 0
	for _, specID := range specIDs {
		specDir := projectDir + "/.moai/specs/" + specID

		// Idempotency: Check current status before updating
		currentStatus, err := spec.ParseStatus(specDir)
		if err != nil {
			slog.Debug("spec-status: failed to parse current status", "spec", specID, "error", err)
			continue
		}

		// Skip if already at target status (idempotent)
		if currentStatus == targetStatus {
			slog.Debug("spec-status: already at target status", "spec", specID, "status", targetStatus)
			continue
		}

		// Update to target status
		if err := spec.UpdateStatus(specDir, targetStatus); err != nil {
			slog.Debug("spec-status: failed to update", "spec", specID, "error", err)
			continue
		}

		slog.Info("spec-status: updated", "spec", specID, "status", targetStatus)
		updatedCount++
	}

	return h.buildOutput(updatedCount, specIDs), nil
}

// buildOutput creates a HookOutput with update results
func (h *specStatusHandler) buildOutput(updatedCount int, specIDs []string) *HookOutput {
	responseData := map[string]any{
		"updated_count": updatedCount,
		"spec_ids":      specIDs,
	}
	rawData, _ := json.Marshal(responseData)
	return &HookOutput{Data: rawData}
}

// isGhPrMergeCommand checks if a command is a gh pr merge command.
// It handles flags like --squash, --merge, --delete-branch, etc.
func (h *specStatusHandler) isGhPrMergeCommand(command string) bool {
	// Pattern: "gh pr merge" followed by optional PR number and any flags
	re := regexp.MustCompile(`gh\s+pr\s+merge`)
	return re.MatchString(command)
}

// isGitCommitCommand checks if a command is a git commit.
func isGitCommitCommand(command string) bool {
	// Simple heuristic: command contains "git" and "commit"
	return strings.Contains(command, "git") && strings.Contains(command, "commit")
}

// extractCommitMessage extracts the commit message from a git commit command.
// It looks for -m flags or --message flags.
func extractCommitMessage(command string) string {
	// Look for -m "message" or --message "message"
	re := regexp.MustCompile(`(?:-m|--message)\s+['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: return empty string
	return ""
}

