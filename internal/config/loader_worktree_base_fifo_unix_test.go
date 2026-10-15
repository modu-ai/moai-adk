//go:build unix

package config

// loader_worktree_base_fifo_unix_test.go — card t1729 audit round-5 P2:
// LoadWorktreeBaseBranch runs on the SessionStart path (the role-rules
// refresh guidance resolves the project's base branch), so a FIFO at
// git-strategy.yaml in a POSIX-linked worktree must not block the session
// start — the reader refuses non-regular files, same guard shape as the
// hook package's rule reads (role_rules_read_unix.go).

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestLoadWorktreeBaseBranchSkipsFIFO(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := syscall.Mkfifo(filepath.Join(cfgDir, "git-strategy.yaml"), 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	done := make(chan string, 1)
	go func() { done <- LoadWorktreeBaseBranch(root) }()
	select {
	case got := <-done:
		if got != "" {
			t.Fatalf("a FIFO git-strategy.yaml must be skipped without a value, got %q", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("LoadWorktreeBaseBranch blocked on a FIFO git-strategy.yaml: the read must refuse non-regular files")
	}
}
