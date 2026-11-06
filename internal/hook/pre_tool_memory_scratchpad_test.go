// Path-boundary tests for card t2102 / GitHub issue #1951: the PreToolUse
// traversal guard must ALLOW the two locations Claude Code itself directs the
// model to write — the project's auto-memory directory and the per-project
// session scratchpad root — while every unrelated outside path stays denied.
//
// Each test runs in a fully isolated environment (t.Setenv for HOME,
// USERPROFILE, CLAUDE_CONFIG_DIR, TMPDIR), so no assertion depends on the
// developer machine's real Claude home or temp layout.
//
// dashEncode: the per-project directory name Claude Code derives from the
// absolute project path. This suite reuses the hook package's own projectSlug
// (maps '/', '\\', '.', ':' to '-'), which reproduces the observed real-world
// form `-Users-goos-MoAI-moai-adk-go` (leading dash from the absolute path's
// leading separator). The issue's inline formula
// strings.ReplaceAll(strings.TrimPrefix(abs, "/"), "/", "-") agrees with it on
// paths without '.' or ':'; projectSlug additionally maps dots (e.g.
// /Users/x/.claude → -Users-x--claude), matching empirically observed
// ~/.claude/projects directory names (see projectSlug's own doc comment).
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// boundaryWriteDecision runs a single Write through the handler's path
// boundary and returns the emitted decision and reason. It follows the
// established constructor pattern of pre_tool_test.go
// (TestPreToolHandler_AllowedExternalPaths): direct preToolHandler struct
// construction with a minimal, allowlist-free SecurityPolicy.
func boundaryWriteDecision(t *testing.T, projectDir, filePath string) (string, string) {
	t.Helper()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := &SecurityPolicy{} // minimal: no allowlist, no deny/ask patterns

	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{
		"file_path": filePath,
		"content":   "t2102 boundary test payload",
	})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-t2102",
		CWD:           projectDir,
		HookEventName: "PreToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(toolInput),
	}

	got, handleErr := handler.Handle(context.Background(), input)
	if handleErr != nil {
		t.Fatalf("unexpected handler error: %v", handleErr)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("got nil output or nil HookSpecificOutput")
	}
	return got.HookSpecificOutput.PermissionDecision, got.HookSpecificOutput.PermissionDecisionReason
}

// wantBoundaryAllow asserts the handler ALLOWED the write.
func wantBoundaryAllow(t *testing.T, name, filePath, decision, reason string) {
	t.Helper()
	if decision != DecisionAllow {
		t.Errorf("%s: decision = %q, want %q (file=%q, reason=%q)",
			name, decision, DecisionAllow, filePath, reason)
	}
}

// wantBoundaryTraversalDeny asserts the handler DENIED the write with the
// path-traversal reason (issue #1951's observed failure surface).
func wantBoundaryTraversalDeny(t *testing.T, name, filePath, decision, reason string) {
	t.Helper()
	if decision != DecisionDeny {
		t.Errorf("%s: decision = %q, want %q (file=%q)", name, decision, DecisionDeny, filePath)
		return
	}
	if !strings.Contains(reason, "Path traversal detected") {
		t.Errorf("%s: reason = %q, want it to contain %q (file=%q)",
			name, reason, "Path traversal detected", filePath)
	}
}

// isolatedBoundaryEnv points HOME/USERPROFILE at tmpHome (default claude home
// tmpHome/.claude), relocates the Claude home to claudeConfigDir, and scopes
// TMPDIR to tmpRoot. It returns tmpHome, claudeConfigDir.
func isolatedBoundaryEnv(t *testing.T, tmpHome, tmpRoot, claudeConfigDir string) {
	t.Helper()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome) // os.UserHomeDir reads USERPROFILE on Windows
	t.Setenv("CLAUDE_CONFIG_DIR", claudeConfigDir)
	t.Setenv("TMPDIR", tmpRoot)
}

// TestPreToolHandler_AllowsClaudeMemoryDirWrite covers the relocated-home case:
// with CLAUDE_CONFIG_DIR set (MoAI launcher profiles relocate the Claude home),
// a Write into <claude-config-dir>/projects/<slug>/memory/ must be allowed
// (issue #1951 case A).
func TestPreToolHandler_AllowsClaudeMemoryDirWrite(t *testing.T) {
	tmpHome := t.TempDir()
	tmpRoot := t.TempDir()
	projectDir := filepath.Join(tmpRoot, "proj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	claudeConfigDir := filepath.Join(tmpHome, "claude-config")
	isolatedBoundaryEnv(t, tmpHome, tmpRoot, claudeConfigDir)

	memoryFile := filepath.Join(claudeConfigDir, "projects", projectSlug(projectDir), "memory", "MEMORY.md")
	decision, reason := boundaryWriteDecision(t, projectDir, memoryFile)
	wantBoundaryAllow(t, "memory dir under CLAUDE_CONFIG_DIR", memoryFile, decision, reason)
}

// TestPreToolHandler_AllowsDefaultHomeMemoryDirWrite covers the default-home
// case: without CLAUDE_CONFIG_DIR, the memory dir lives under
// ~/.claude/projects/<slug>/memory/.
func TestPreToolHandler_AllowsDefaultHomeMemoryDirWrite(t *testing.T) {
	tmpHome := t.TempDir()
	tmpRoot := t.TempDir()
	projectDir := filepath.Join(tmpRoot, "proj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	isolatedBoundaryEnv(t, tmpHome, tmpRoot, "") // empty value = unset

	memoryFile := filepath.Join(tmpHome, ".claude", "projects", projectSlug(projectDir), "memory", "MEMORY.md")
	decision, reason := boundaryWriteDecision(t, projectDir, memoryFile)
	wantBoundaryAllow(t, "memory dir under ~/.claude", memoryFile, decision, reason)
}

// TestPreToolHandler_AllowsSessionScratchpadWrite covers issue #1951 case B:
// a Write under <tmp>/claude-<uid>/<slug>/…/<session>/scratchpad/ must be
// allowed. The WHOLE per-project scratchpad subtree is allowed because the
// hook has no way to know the current session id — covering every session's
// scratchpad is the only form the allowlist can take.
func TestPreToolHandler_AllowsSessionScratchpadWrite(t *testing.T) {
	tmpHome := t.TempDir()

	tests := []struct {
		name string
		mk   func(t *testing.T) (tmpRoot, projectDir, file string)
	}{
		{
			name: "scratchpad session dir exists",
			mk: func(t *testing.T) (string, string, string) {
				tmpRoot := t.TempDir()
				projectDir := filepath.Join(tmpRoot, "proj")
				if err := os.MkdirAll(projectDir, 0o755); err != nil {
					t.Fatalf("create project dir: %v", err)
				}
				scratch := filepath.Join(tmpRoot, fmt.Sprintf("claude-%d", os.Getuid()),
					projectSlug(projectDir), "test-session", "scratchpad")
				if err := os.MkdirAll(scratch, 0o755); err != nil {
					t.Fatalf("create scratchpad dir: %v", err)
				}
				return tmpRoot, projectDir, filepath.Join(scratch, "note.txt")
			},
		},
		{
			name: "fresh session scratchpad not yet created",
			mk: func(t *testing.T) (string, string, string) {
				// A new session's first write may race ahead of any directory
				// creation: the target's leaf dirs do not exist yet. Purely
				// lexical Rel against the computed base still decides this.
				tmpRoot := t.TempDir()
				projectDir := filepath.Join(tmpRoot, "proj")
				if err := os.MkdirAll(projectDir, 0o755); err != nil {
					t.Fatalf("create project dir: %v", err)
				}
				file := filepath.Join(tmpRoot, fmt.Sprintf("claude-%d", os.Getuid()),
					projectSlug(projectDir), "fresh-session", "scratchpad", "note.txt")
				return tmpRoot, projectDir, file
			},
		},
		{
			name: "symlinked TMPDIR with fresh session",
			mk: func(t *testing.T) (string, string, string) {
				// macOS resolves /tmp -> /private/tmp. The target-side walk
				// resolves through the symlinked ancestor while a lexical
				// allowlist entry would not, so the entry needs the same
				// existing-parent resolution the target gets.
				tmpRoot := t.TempDir()
				real := filepath.Join(tmpRoot, "real-tmp")
				if err := os.MkdirAll(real, 0o755); err != nil {
					t.Fatalf("create real tmp dir: %v", err)
				}
				link := filepath.Join(tmpRoot, "link-tmp")
				if err := os.Symlink(real, link); err != nil {
					t.Fatalf("create tmp symlink: %v", err)
				}
				projectDir := filepath.Join(real, "proj")
				if err := os.MkdirAll(projectDir, 0o755); err != nil {
					t.Fatalf("create project dir: %v", err)
				}
				file := filepath.Join(link, fmt.Sprintf("claude-%d", os.Getuid()),
					projectSlug(projectDir), "fresh-session", "scratchpad", "note.txt")
				return link, projectDir, file
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpRoot, projectDir, file := tt.mk(t)
			isolatedBoundaryEnv(t, tmpHome, tmpRoot, "")
			decision, reason := boundaryWriteDecision(t, projectDir, file)
			wantBoundaryAllow(t, tt.name, file, decision, reason)
		})
	}
}

// TestPreToolHandler_StillDeniesUnrelatedExternalPaths is the negative
// control for issue #1951: the two computed allow entries must not open the
// whole home or tmp tree. Unrelated outside paths — including a sibling
// project's memory dir and a non-memory path inside THIS project's Claude
// projects subtree — stay denied with the path-traversal reason.
func TestPreToolHandler_StillDeniesUnrelatedExternalPaths(t *testing.T) {
	tmpHome := t.TempDir()
	tmpRoot := t.TempDir()
	projectDir := filepath.Join(tmpRoot, "proj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}
	otherProject := filepath.Join(tmpRoot, "other-project")

	claudeConfigDir := filepath.Join(tmpHome, "claude-config")
	isolatedBoundaryEnv(t, tmpHome, tmpRoot, claudeConfigDir)

	tests := []struct {
		name string
		file string
	}{
		{
			name: "unrelated tmp path",
			file: filepath.Join(tmpRoot, "unrelated", "secret.txt"),
		},
		{
			name: "another project's memory dir",
			file: filepath.Join(claudeConfigDir, "projects", projectSlug(otherProject), "memory", "MEMORY.md"),
		},
		{
			name: "non-memory path inside this project's claude subtree",
			file: filepath.Join(claudeConfigDir, "projects", projectSlug(projectDir), "not-memory.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, reason := boundaryWriteDecision(t, projectDir, tt.file)
			wantBoundaryTraversalDeny(t, tt.name, tt.file, decision, reason)
		})
	}
}
