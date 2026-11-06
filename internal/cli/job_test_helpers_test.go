package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// jobTestRoot is a MoAI project root for the background-job tests: a git
// repository with a .moai directory, the shape a mutating call's project_root
// must have (SPEC-MCP-NEUTRALITY-001 REQ-MN-004/005). It lived beside the codex
// job tests until the general codex execution surface was removed (card t1909);
// the GLM job tests still need it.
func jobTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// thisRepoRoot returns the toplevel of the repository containing the test's
// cwd — a registered worktree of the default (real) serving repository, so raw
// handler calls can pass the project_root gate.
func thisRepoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	top, err := codexAuditGit(context.Background(), cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatalf("test cwd is not inside a git worktree: %v", err)
	}
	return top
}
