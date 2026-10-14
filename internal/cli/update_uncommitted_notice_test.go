package cli

// t1729 — tests for the end-of-update uncommitted-template notice.
//
// The notice fires when the working tree carries uncommitted changes to the
// template files the sync just analyzed: worktrees branched from the
// committed base cannot see them until the operator commits (then syncs),
// which is exactly the stale-deployed-rule shape the SessionStart role-rule
// injection logs in factory sessions.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTrackedFile turns dir into a git repository with one tracked, committed
// file at rel (content v1). Returns the file's absolute path.
func initTrackedFile(t *testing.T, dir, rel, content string) string {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for seed file: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	run("add", ".")
	run("-c", "user.email=fixture@example.com", "-c", "user.name=fixture", "commit", "-q", "-m", "seed")
	return path
}

func TestUncommittedTemplateNotice(t *testing.T) {
	notice := func(t *testing.T, root string, paths ...string) string {
		t.Helper()
		return uncommittedTemplateNotice(root, paths)
	}

	t.Run("no_change_set_no_notice", func(t *testing.T) {
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if got := notice(t, dir); got != "" {
			t.Errorf("empty change set must produce no notice, got: %s", got)
		}
	})

	t.Run("clean_tree_no_notice", func(t *testing.T) {
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md"); got != "" {
			t.Errorf("committed (clean) tree must produce no notice, got: %s", got)
		}
	})

	t.Run("uncommitted_template_change_fires", func(t *testing.T) {
		dir := t.TempDir()
		path := initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		// The update's uncommitted write: the working tree diverges from HEAD.
		if err := os.WriteFile(path, []byte("v2 refreshed by the update\n"), 0o644); err != nil {
			t.Fatalf("rewrite: %v", err)
		}
		got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md")
		if got == "" {
			t.Fatal("uncommitted template change must fire the notice")
		}
		if !strings.Contains(got, "moai worktree sync") {
			t.Errorf("notice does not name the worktree refresh verb: %s", got)
		}
		if !strings.Contains(got, "1 template file") {
			t.Errorf("notice does not carry the changed-file count: %s", got)
		}
	})

	t.Run("staged_only_change_still_fires", func(t *testing.T) {
		dir := t.TempDir()
		path := initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		if err := os.WriteFile(path, []byte("v2\n"), 0o644); err != nil {
			t.Fatalf("rewrite: %v", err)
		}
		cmd := exec.Command("git", "-C", dir, "add", ".")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git add: %v: %s", err, out)
		}
		if got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md"); got == "" {
			t.Fatal("a staged-but-uncommitted change is still uncommitted — the notice must fire")
		}
	})

	t.Run("unrelated_change_no_notice", func(t *testing.T) {
		dir := t.TempDir()
		initTrackedFile(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md", "v1\n")
		other := filepath.Join(dir, "user-notes.md")
		if err := os.WriteFile(other, []byte("the operator's own edit\n"), 0o644); err != nil {
			t.Fatalf("write unrelated: %v", err)
		}
		if got := notice(t, dir, ".claude/rules/moai/workflow/factory-dispatch.md"); got != "" {
			t.Errorf("a user's unrelated edit must not trigger the notice, got: %s", got)
		}
	})

	t.Run("not_a_repo_fails_open", func(t *testing.T) {
		dir := t.TempDir()
		if got := notice(t, dir, "any/path.md"); got != "" {
			t.Errorf("non-repository root must fail open (no notice), got: %s", got)
		}
	})
}
