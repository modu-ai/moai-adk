package hook

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WorktreeEntry records a currently active isolated worktree.
type WorktreeEntry struct {
	Path      string    `json:"path"`
	Branch    string    `json:"branch"`
	AgentName string    `json:"agent_name"`
	CreatedAt time.Time `json:"created_at"`
}

// @MX:WARN: [AUTO] Package-level mutex guards shared worktree state file; all reads and writes must acquire this lock
// @MX:REASON: global state mutation — concurrent hook handlers (WorktreeCreate/WorktreeRemove) share this lock; missing lock acquisition causes data races on the JSON state file
var worktreeMu sync.Mutex

// registerWorktree appends a new WorktreeEntry to the persistent state file.
// If the state directory or file does not exist it is created.
// Non-blocking on error: failures are logged as warnings.
func registerWorktree(projectDir, path, branch, agentName string) {
	worktreeMu.Lock()
	defer worktreeMu.Unlock()

	stateFile := worktreeStateFile(projectDir)
	entries := loadWorktreeEntries(stateFile)

	entries = append(entries, WorktreeEntry{
		Path:      path,
		Branch:    branch,
		AgentName: agentName,
		CreatedAt: time.Now(),
	})

	saveWorktreeEntries(stateFile, entries)
}

// unregisterWorktree removes the entry matching the given path from the
// persistent state file.  Non-blocking on error: failures are logged.
func unregisterWorktree(projectDir, path string) {
	worktreeMu.Lock()
	defer worktreeMu.Unlock()

	stateFile := worktreeStateFile(projectDir)
	entries := loadWorktreeEntries(stateFile)

	filtered := make([]WorktreeEntry, 0, len(entries))
	for _, e := range entries {
		if e.Path != path {
			filtered = append(filtered, e)
		}
	}

	saveWorktreeEntries(stateFile, filtered)
}

// worktreeStateFile returns the canonical path to the worktrees state file.
func worktreeStateFile(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state", "worktrees.json")
}

// loadWorktreeEntries reads entries from the state file.
// Returns an empty slice on any error (missing file, parse error, etc.).
func loadWorktreeEntries(path string) []WorktreeEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		// File not existing is normal; other errors are logged.
		if !os.IsNotExist(err) {
			slog.Warn("worktree registry: failed to read state file",
				"path", path,
				"error", err,
			)
		}
		return []WorktreeEntry{}
	}

	var entries []WorktreeEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		slog.Warn("worktree registry: failed to parse state file",
			"path", path,
			"error", err,
		)
		return []WorktreeEntry{}
	}
	return entries
}

// saveWorktreeEntries writes entries to the state file.
// Creates intermediate directories as needed.
// Logs a warning on error rather than propagating.
func saveWorktreeEntries(path string, entries []WorktreeEntry) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		slog.Warn("worktree registry: failed to create state directory",
			"path", filepath.Dir(path),
			"error", err,
		)
		return
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		slog.Warn("worktree registry: failed to marshal entries", "error", err)
		return
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		slog.Warn("worktree registry: failed to write state file",
			"path", path,
			"error", err,
		)
	}
}
