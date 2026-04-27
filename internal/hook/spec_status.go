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

// Handle processes a PostToolUse event. It checks if the tool was a git commit,
// extracts SPEC-IDs from the commit message, and updates their status to "implemented".
// Errors are non-blocking - the handler always returns success.
func (h *specStatusHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	projectDir := resolveProjectDir(input)

	// Parse hook input data
	var data struct {
		ToolName string `json:"tool_name"`
		Command  string `json:"command"`
	}
	if input.Data != nil {
		if err := json.Unmarshal(input.Data, &data); err != nil {
			slog.Debug("spec-status: failed to unmarshal hook data", "error", err)
			return &HookOutput{}, nil
		}
	}

	// Only process git commit commands
	if !isGitCommitCommand(data.Command) {
		return &HookOutput{}, nil
	}

	// Extract commit message (simulate "git log -1 --format=%s")
	commitMsg := extractCommitMessage(data.Command)
	if commitMsg == "" {
		return &HookOutput{}, nil
	}

	// Extract SPEC-IDs from commit message
	specIDs := extractSPECIDs(commitMsg)
	if len(specIDs) == 0 {
		return &HookOutput{}, nil
	}

	// Update each SPEC status
	updatedCount := 0
	for _, specID := range specIDs {
		specDir := projectDir + "/.moai/specs/" + specID
		if err := spec.UpdateStatus(specDir, "implemented"); err != nil {
			// Log but don't fail - SPEC might not exist or already be implemented
			slog.Debug("spec-status: failed to update", "spec", specID, "error", err)
			continue
		}
		slog.Info("spec-status: updated", "spec", specID, "status", "implemented")
		updatedCount++
	}

	// Build response data
	responseData := map[string]any{
		"updated_count": updatedCount,
		"spec_ids":      specIDs,
	}
	rawData, _ := json.Marshal(responseData)

	return &HookOutput{Data: rawData}, nil
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

// extractSPECIDs extracts all SPEC-XXX patterns from a commit message.
// Pattern: SPEC-[A-Z0-9]+-[0-9]+ (e.g., SPEC-STATUS-AUTO-001, SPEC-V3R2-CON-001)
func extractSPECIDs(commitMsg string) []string {
	// Pattern matches SPEC- followed by alphanumeric chars and dashes, ending with -digits
	// Examples: SPEC-TEST-001, SPEC-STATUS-AUTO-001, SPEC-V3R2-CON-001
	re := regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]+`)
	matches := re.FindAllString(commitMsg, -1)

	// Deduplicate while preserving order
	seen := make(map[string]bool)
	var result []string
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			result = append(result, match)
		}
	}

	return result
}
