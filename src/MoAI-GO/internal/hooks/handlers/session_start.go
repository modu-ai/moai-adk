package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
)

// SessionStartHandler handles session-start events
type SessionStartHandler struct {
	projectRoot string
}

// NewSessionStartHandler creates a new session-start handler
func NewSessionStartHandler() *SessionStartHandler {
	return &SessionStartHandler{
		projectRoot: findProjectRoot(),
	}
}

// Handle executes the session-start hook
func (h *SessionStartHandler) Handle(input *protocol.HookInput) (*protocol.HookResponse, error) {
	var output []string

	output = append(output, "🚀 MoAI-ADK Session Started")

	// Get version info
	version := h.getVersion()
	output = append(output, fmt.Sprintf("   📦 Version: %s", version))

	// Get git info
	gitInfo := h.getGitInfo()
	output = append(output,
		fmt.Sprintf("   🌿 Branch: %s", gitInfo.branch),
		fmt.Sprintf("   🔨 Last Commit: %s", gitInfo.lastCommit),
		fmt.Sprintf("   🔄 Changes: %d", gitInfo.changes),
	)

	// Get language info
	langInfo := h.getLanguageInfo()
	output = append(output, fmt.Sprintf("   🌐 Language: %s (%s)", langInfo.name, langInfo.code))

	return protocol.NewMessageResponse(strings.Join(output, "\n"), false), nil
}

// getVersion gets the MoAI-ADK version
func (h *SessionStartHandler) getVersion() string {
	// Try to get version from moai-adk command
	cmd := exec.Command("moai", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse version from output
	version := strings.TrimSpace(string(output))
	if strings.HasPrefix(version, "MoAI version") {
		parts := strings.Fields(version)
		if len(parts) >= 3 {
			return parts[2]
		}
	}

	return version
}

// gitInfo contains git repository information
type gitInfo struct {
	branch     string
	lastCommit string
	changes    int
}

// getGitInfo gets git repository information
func (h *SessionStartHandler) getGitInfo() gitInfo {
	info := gitInfo{
		branch:     "Git not initialized",
		lastCommit: "No commits yet",
		changes:    0,
	}

	// Check if git is initialized
	if _, err := os.Stat(filepath.Join(h.projectRoot, ".git")); os.IsNotExist(err) {
		return info
	}

	// Get branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = h.projectRoot
	if output, err := cmd.Output(); err == nil {
		info.branch = strings.TrimSpace(string(output))
		if info.branch == "" {
			info.branch = "HEAD detached"
		}
	}

	// Get last commit
	cmd = exec.Command("git", "log", "--pretty=format:%h %s", "-1")
	cmd.Dir = h.projectRoot
	if output, err := cmd.Output(); err == nil {
		info.lastCommit = strings.TrimSpace(string(output))
	}

	// Get changes count
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = h.projectRoot
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		changes := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				changes++
			}
		}
		info.changes = changes
	}

	return info
}

// languageInfo contains programming language information
type languageInfo struct {
	code string
	name string
}

// getLanguageInfo detects the project's primary programming language
func (h *SessionStartHandler) getLanguageInfo() languageInfo {
	// Check for common project files
	projectFiles := map[string]languageInfo{
		"pyproject.toml":   {code: "py", name: "Python"},
		"setup.py":         {code: "py", name: "Python"},
		"package.json":     {code: "js", name: "JavaScript/TypeScript"},
		"go.mod":           {code: "go", name: "Go"},
		"Cargo.toml":       {code: "rs", name: "Rust"},
		"pom.xml":          {code: "java", name: "Java"},
		"build.gradle":     {code: "java", name: "Java"},
		"Gemfile":          {code: "rb", name: "Ruby"},
		"composer.json":    {code: "php", name: "PHP"},
		"requirements.txt": {code: "py", name: "Python"},
		"tsconfig.json":    {code: "ts", name: "TypeScript"},
	}

	for file, info := range projectFiles {
		if _, err := os.Stat(filepath.Join(h.projectRoot, file)); err == nil {
			return info
		}
	}

	// Default to unknown
	return languageInfo{code: "unknown", name: "Unknown"}
}
