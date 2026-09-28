package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveCodexWorktreeDirDualRoot(t *testing.T) {
	root := t.TempDir()
	modern := filepath.Join(root, ".moai", "worktrees", "card")
	legacy := filepath.Join(root, ".claude", "worktrees", "card")
	if err := os.MkdirAll(modern, 0o755); err != nil {
		t.Fatal(err)
	}
	if got, err := resolveCodexWorktreeDir(root, "card"); err != nil || got != modern {
		t.Fatalf("new root: got %q, %v; want %q", got, err, modern)
	}
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveCodexWorktreeDir(root, "card"); err == nil || !strings.Contains(err.Error(), "use an absolute path") {
		t.Fatalf("duplicate name must be ambiguous, got %v", err)
	}
	if err := os.RemoveAll(modern); err != nil {
		t.Fatal(err)
	}
	if got, err := resolveCodexWorktreeDir(root, "card"); err != nil || got != legacy {
		t.Fatalf("legacy fallback: got %q, %v; want %q", got, err, legacy)
	}
}

func TestResolveCodexWorktreeDirRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"..", "../outside", "child/nested"} {
		if _, err := resolveCodexWorktreeDir(root, name); err == nil {
			t.Errorf("accepted relative worktree name %q", name)
		}
	}
}
