package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
)

// SessionEndHandler handles session-end events
type SessionEndHandler struct {
	projectRoot string
}

// NewSessionEndHandler creates a new session-end handler
func NewSessionEndHandler() *SessionEndHandler {
	return &SessionEndHandler{
		projectRoot: findProjectRoot(),
	}
}

// Handle executes the session-end hook
func (h *SessionEndHandler) Handle(input *protocol.HookInput) (*protocol.HookResponse, error) {
	var output []string

	output = append(output, "Session Ended")

	// Get current branch
	branch := h.getBranch()
	output = append(output, fmt.Sprintf("   Branch: %s", branch))

	// Get uncommitted changes count
	changes := h.getUncommittedChangesCount()
	output = append(output, fmt.Sprintf("   Uncommitted changes: %d", changes))

	// Add reminder if there are uncommitted changes
	if changes > 0 {
		output = append(output, fmt.Sprintf("   Reminder: You have %d uncommitted change(s). Consider committing before your next session.", changes))
	}

	return protocol.NewMessageResponse(strings.Join(output, "\n"), false), nil
}

// getBranch returns the current git branch name
func (h *SessionEndHandler) getBranch() string {
	// Check if git is initialized
	if _, err := os.Stat(filepath.Join(h.projectRoot, ".git")); os.IsNotExist(err) {
		return "Git not initialized"
	}

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = h.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "HEAD detached"
	}
	return branch
}

// getUncommittedChangesCount returns the number of uncommitted changes
func (h *SessionEndHandler) getUncommittedChangesCount() int {
	// Check if git is initialized
	if _, err := os.Stat(filepath.Join(h.projectRoot, ".git")); os.IsNotExist(err) {
		return 0
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = h.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	count := 0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
