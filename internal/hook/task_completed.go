// Resolution: KEEP — SPEC acceptance criteria validation; Continue:false on fail.
package hook

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/workflow"
)

// taskCompletedHandler processes TaskCompleted events.
// In team mode, it validates task deliverables before accepting completion.
type taskCompletedHandler struct{}

// NewTaskCompletedHandler creates a new TaskCompleted event handler.
func NewTaskCompletedHandler() Handler {
	return &taskCompletedHandler{}
}

// EventType returns EventTaskCompleted.
func (h *taskCompletedHandler) EventType() EventType {
	return EventTaskCompleted
}

// Handle processes a TaskCompleted event.
// Returns empty output to accept completion.
// Returns NewTaskRejectedOutput(reason) — top-level decision:"block" + reason
// JSON with exit 0, per official TaskCompleted protocol — to reject completion.
//
// The task is identified by the official task_name stdin field, falling back
// to the legacy MoAI task_subject field. Validation applies when either the
// legacy team-mode marker (team_name) or an official task_name is present.
// If the task name references a SPEC ID, the corresponding spec.md must exist.
func (h *taskCompletedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Task identification: official task_name first, legacy task_subject fallback.
	taskName := input.TaskName
	if taskName == "" {
		taskName = input.TaskSubject
	}

	slog.Info("task completed",
		"session_id", input.SessionID,
		"task_id", input.TaskID,
		"task_name", taskName,
		"teammate", input.TeammateName,
		"team", input.TeamName,
	)

	// Only enforce validation in team context: legacy team_name marker OR
	// official task_name (the official runtime never sends team_name).
	if input.TeamName == "" && input.TaskName == "" {
		return &HookOutput{}, nil
	}

	projectDir := input.ProjectDir
	if projectDir == "" {
		projectDir = input.CWD
	}

	// If the task name references a SPEC ID, verify the SPEC file exists.
	if projectDir != "" && taskName != "" {
		if specID := workflow.SpecIDPattern.FindString(taskName); specID != "" {
			specPath := filepath.Join(projectDir, ".moai", "specs", specID, "spec.md")
			if _, err := os.Stat(specPath); os.IsNotExist(err) {
				msg := fmt.Sprintf(
					"Task %q references SPEC %s but spec.md not found at %s.\nCreate the SPEC document before marking task complete.",
					taskName, specID, specPath,
				)
				fmt.Fprint(os.Stderr, msg)
				slog.Warn("task_completed: rejecting completion - SPEC not found",
					"task_name", taskName,
					"spec_id", specID,
					"spec_path", specPath,
				)
				return NewTaskRejectedOutput(msg), nil
			}

			// Check for unchecked acceptance criteria in spec.md.
			if unchecked := parseUncheckedCriteria(specPath); len(unchecked) > 0 {
				var sb strings.Builder
				fmt.Fprintf(&sb, "Task %q has %d unchecked acceptance criteria in SPEC %s:\n",
					taskName, len(unchecked), specID)
				for _, criterion := range unchecked {
					fmt.Fprintf(&sb, "  %s\n", criterion)
				}
				sb.WriteString("Mark these criteria as complete ([x]) in spec.md before marking the task complete.")
				msg := sb.String()
				fmt.Fprint(os.Stderr, msg)
				slog.Warn("task_completed: rejecting completion - unchecked acceptance criteria",
					"task_name", taskName,
					"spec_id", specID,
					"unchecked_count", len(unchecked),
				)
				return NewTaskRejectedOutput(msg), nil
			}
		}
	}

	return &HookOutput{}, nil
}

// parseUncheckedCriteria reads a spec.md file and returns unchecked acceptance criteria.
// Returns nil if the acceptance criteria section is not found or the file cannot be read.
func parseUncheckedCriteria(specPath string) []string {
	f, err := os.Open(specPath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var (
		inSection bool
		unchecked []string
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Detect section headers.
		if strings.HasPrefix(line, "## ") {
			if inSection {
				// Reached the next section; stop collecting.
				break
			}
			if strings.EqualFold(strings.TrimSpace(line), "## Acceptance Criteria") {
				inSection = true
			}
			continue
		}

		if inSection && strings.HasPrefix(strings.TrimSpace(line), "- [ ] ") {
			unchecked = append(unchecked, strings.TrimSpace(line))
		}
	}

	return unchecked
}
