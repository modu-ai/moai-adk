package hook

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/workflow"
)

// specFilePattern is the glob pattern for finding spec.md files in SPEC directories.
const specFilePattern = ".moai/specs/*/spec.md"

// userPromptSubmitHandler handles UserPromptSubmit events.
// It automatically generates a session title from the CWD and returns it to Claude Code.
type userPromptSubmitHandler struct {
	cfg ConfigProvider
}

// NewUserPromptSubmitHandler creates a new UserPromptSubmit event handler.
func NewUserPromptSubmitHandler(cfg ConfigProvider) Handler {
	return &userPromptSubmitHandler{cfg: cfg}
}

// EventType returns EventUserPromptSubmit.
func (h *userPromptSubmitHandler) EventType() EventType {
	return EventUserPromptSubmit
}

// workflowKeywords are prompt keywords that indicate an active MoAI workflow context.
var workflowKeywords = []string{"loop", "run", "plan"}

// detectWorkflowContext checks whether the prompt contains any workflow keywords
// and returns a non-empty additionalContext string if a match is found.
func detectWorkflowContext(prompt string) string {
	lower := strings.ToLower(prompt)
	for _, kw := range workflowKeywords {
		if strings.Contains(lower, kw) {
			return "workflow keyword '" + kw + "' detected — MoAI workflow context may be active"
		}
	}
	return ""
}

// Handle processes a UserPromptSubmit event.
// It detects active SPECs from CWD to build a session title and returns it via hookSpecificOutput.
// Errors are silently handled; an empty title is returned without blocking the prompt.
func (h *userPromptSubmitHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Log prompt for audit purposes (truncated to 100 chars)
	prompt := input.Prompt
	preview := prompt
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	slog.Info("user prompt submitted",
		"session_id", input.SessionID,
		"prompt_preview", preview,
	)

	// Build session title (errors are silently ignored, falls back to empty title)
	title := h.buildSessionTitle(input.CWD)

	// Detect workflow context
	additionalCtx := detectWorkflowContext(prompt)

	// Return empty output if no context to report
	if title == "" && additionalCtx == "" {
		return &HookOutput{}, nil
	}

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     string(EventUserPromptSubmit),
			SessionTitle:      title,
			AdditionalContext: additionalCtx,
		},
	}, nil
}

// buildSessionTitle generates a session title based on CWD.
// With SPEC: "SPEC-ID: title"
// Without SPEC: "projectName / branchName"
// On error: returns empty string
func (h *userPromptSubmitHandler) buildSessionTitle(cwd string) string {
	if cwd == "" {
		return ""
	}

	// Attempt to detect active SPEC
	if title := detectActiveSpec(cwd); title != "" {
		return title
	}

	// Fall back to projectName/branchName combination
	return buildProjectBranchTitle(cwd)
}

// detectActiveSpec searches cwd/.moai/specs/*/spec.md and returns the title
// of the most recently modified SPEC.
// Returns empty string if no SPEC is found or read fails.
func detectActiveSpec(cwd string) string {
	pattern := filepath.Join(cwd, specFilePattern)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}

	// Select the most recently modified spec.md
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

	// Extract SPEC ID from directory name
	specDirName := filepath.Base(filepath.Dir(latestMatch))
	specID := workflow.SpecIDPattern.FindString(specDirName)
	if specID == "" {
		return ""
	}

	// Read the first heading (line starting with #) from spec.md
	heading := readFirstHeading(latestMatch)
	if heading == "" {
		return specID
	}

	// Strip SPEC-ID prefix from heading if already present (e.g., "SPEC-SRS-003: Dashboard...")
	if strings.HasPrefix(heading, specID+": ") {
		heading = strings.TrimPrefix(heading, specID+": ")
	} else if strings.HasPrefix(heading, specID) {
		heading = strings.TrimPrefix(heading, specID)
		heading = strings.TrimLeft(heading, ": ")
	}

	if heading == "" {
		return specID
	}

	return fmt.Sprintf("%s: %s", specID, heading)
}

// readFirstHeading returns the first # heading text from a markdown file.
func readFirstHeading(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// buildProjectBranchTitle generates a title in "projectName / branchName" format.
// Returns "projectName / unknown" if git branch lookup fails.
func buildProjectBranchTitle(cwd string) string {
	projectName := filepath.Base(cwd)
	if projectName == "" || projectName == "." {
		projectName = "unknown"
	}

	branch := getGitBranch(cwd)
	if branch == "" {
		branch = "unknown"
	}

	return fmt.Sprintf("%s / %s", projectName, branch)
}

// getGitBranch retrieves the current branch name using git rev-parse --abbrev-ref HEAD.
// Returns empty string on failure.
func getGitBranch(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}
