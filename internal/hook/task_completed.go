package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
)

// specIDPattern matches SPEC identifiers in task subjects (e.g., SPEC-TEAM-001).
var specIDPattern = regexp.MustCompile(`SPEC-[A-Z]+-\d+`)

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
// Returns empty output (exit code 0) to accept completion.
// Returns NewTaskRejectedOutput() (exit code 2) to reject completion.
//
// Validation only applies in team mode (TeamName non-empty).
// If a task subject references a SPEC ID, the corresponding spec.md must exist.
func (h *taskCompletedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("task completed",
		"session_id", input.SessionID,
		"task_id", input.TaskID,
		"task_subject", input.TaskSubject,
		"teammate", input.TeammateName,
		"team", input.TeamName,
	)

	// Only enforce validation in team mode.
	if input.TeamName == "" {
		return &HookOutput{}, nil
	}

	projectDir := input.ProjectDir
	if projectDir == "" {
		projectDir = input.CWD
	}

	// If the task subject references a SPEC ID, verify the SPEC file exists.
	if projectDir != "" && input.TaskSubject != "" {
		if specID := specIDPattern.FindString(input.TaskSubject); specID != "" {
			specPath := filepath.Join(projectDir, ".moai", "specs", specID, "spec.md")
			if _, err := os.Stat(specPath); os.IsNotExist(err) {
				msg := fmt.Sprintf(
					"Task %q references SPEC %s but spec.md not found at %s.\nCreate the SPEC document before marking task complete.",
					input.TaskSubject, specID, specPath,
				)
				fmt.Fprint(os.Stderr, msg)
				slog.Warn("task_completed: rejecting completion - SPEC not found",
					"task_subject", input.TaskSubject,
					"spec_id", specID,
					"spec_path", specPath,
				)
				return NewTaskRejectedOutput(), nil
			}
		}
	}

	return &HookOutput{}, nil
}
