package worktree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// yaml_quote_test.go exercises SPEC-CLIFIX-HYGIENE-001 AC-HYG-001-008:
//   - commented-out team_mode / tmux_preferred lines MUST NOT activate CG mode
//     or flip the launcher-selection parse (the pre-fix strings.Contains
//     line-matcher matched comments);
//   - a worktree path containing a space MUST survive the tmux initial command
//     (the pre-fix unquoted `cd <path>` broke on spaces).

// writeConfig writes the given YAML body to <dir>/.moai/config/sections/<name>.
func writeConfig(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, ".moai", "config", "sections", name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// TestYAMLCommentCGTeamMode reproduces the comment-matching defect in
// GetActiveMode: a llm.yaml where team_mode exists ONLY as a comment must
// yield "cc" (no CG activation). The pre-fix strings.Contains(trimmed, "team_mode:")
// matches the comment line and returns "cg".
func TestYAMLCommentCGTeamMode(t *testing.T) {
	dir := t.TempDir()
	// team_mode appears ONLY in a comment — the real field is absent/empty.
	writeConfig(t, dir, "llm.yaml", "llm:\n    mode: \"\"\n    # team_mode: cg\n")
	got, err := GetActiveMode(dir)
	if err != nil {
		t.Fatalf("GetActiveMode: %v", err)
	}
	if got != "cc" {
		t.Errorf("GetActiveMode with comment-only team_mode = %q, want %q (comment must not activate CG)", got, "cc")
	}
}

// TestTmuxPreferredParse reproduces the comment-matching defect in
// isTmuxPreferred: a workflow.yaml where tmux_preferred exists ONLY as a
// comment must yield false. The pre-fix strings.Contains(trimmed, "tmux_preferred:")
// matches the comment and returns true.
func TestTmuxPreferredParse(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "workflow.yaml", "workflow:\n    worktree:\n        auto_create: false\n        # tmux_preferred: true\n")
	orig := gitRepoRootFunc
	gitRepoRootFunc = func() (string, error) { return dir, nil }
	t.Cleanup(func() { gitRepoRootFunc = orig })
	if got := isTmuxPreferred(); got {
		t.Errorf("isTmuxPreferred with comment-only tmux_preferred = true, want false (comment must not flip launcher parse)")
	}
}

// TestTmuxPreferredParseRealTrue verifies the real (non-commented) value still
// parses correctly after the yaml.v3 migration — a regression guard so the
// comment fix does not swallow genuine values.
func TestTmuxPreferredParseRealTrue(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "workflow.yaml", "workflow:\n    worktree:\n        tmux_preferred: true\n")
	orig := gitRepoRootFunc
	gitRepoRootFunc = func() (string, error) { return dir, nil }
	t.Cleanup(func() { gitRepoRootFunc = orig })
	if got := isTmuxPreferred(); !got {
		t.Errorf("isTmuxPreferred with real tmux_preferred: true = false, want true")
	}
}

// TestTmuxPathQuote reproduces the unquoted-path defect: a worktree path
// containing a space must appear quoted in the tmux initial command so the
// shell receives the path intact. The pre-fix `cd %s` produced `cd my project`.
func TestTmuxPathQuote(t *testing.T) {
	cfg := &TmuxSessionConfig{
		ProjectName:  "proj",
		SpecID:       "SPEC-X-001",
		WorktreePath: "/tmp/my project",
		ActiveMode:   "cc",
	}
	cmd := buildTmuxInitialCommand(cfg)
	// The cd segment must quote the path (double-quoted) so the space survives.
	if !strings.Contains(cmd, `cd "/tmp/my project"`) {
		t.Errorf("tmux initial command does not quote the spaced path: %q (want cd \"/tmp/my project\")", cmd)
	}
}
