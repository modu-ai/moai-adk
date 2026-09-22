package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	wtcmd "github.com/modu-ai/moai-adk/internal/cli/worktree"
)

func TestWorktreeNew_MissingNameFangErrorNamesExpectedArgument(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"worktree", "new"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	err := runFang(context.Background(), rootCmd)
	if err == nil {
		t.Fatal("missing worktree name accepted")
	}
	if !strings.Contains(err.Error(), "expected argument <name>") {
		t.Fatalf("returned error = %q, want expected argument <name>", err)
	}
	if !strings.Contains(strings.ToLower(stderr.String()), "expected argument <name>") {
		t.Fatalf("stderr = %q, want expected argument <name>", stderr.String())
	}
}

func TestWorktreeNew_WiresTheSharedMaterializer(t *testing.T) {
	root := t.TempDir()
	var addCalls int
	var gotDest, gotBranch string
	swapSessionWorktreeSeams(t, swSeams{
		add: func(dest, branch, _ string) (string, error) {
			addCalls++
			gotDest, gotBranch = dest, branch
			return dest, nil
		},
		commonDir: func() (string, error) { return filepath.Join(root, ".git"), nil },
		configSet: func(string, string, string) error { return nil },
	})

	if wtcmd.WorktreeCreator == nil {
		t.Fatal("worktree creator adapter is not wired")
	}
	gotPath, err := wtcmd.WorktreeCreator("probe-card", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("create through adapter: %v", err)
	}
	wantPath := filepath.Join(root, ".claude", "worktrees", "probe-card")
	if gotPath != wantPath || gotDest != wantPath || gotBranch != "probe-card" {
		t.Fatalf("path=%q dest=%q branch=%q, want path/dest=%q branch=probe-card", gotPath, gotDest, gotBranch, wantPath)
	}
	if addCalls != 1 {
		t.Fatalf("git worktree add seam calls = %d, want exactly 1", addCalls)
	}
}

func TestWorktreeNew_RefusesPlainDirectoryBeforeGit(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, ".claude", "worktrees", "plain-dir")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	var addCalls int
	swapSessionWorktreeSeams(t, swSeams{
		add: func(_, _, _ string) (string, error) {
			addCalls++
			return "", nil
		},
		commonDir: func() (string, error) { return filepath.Join(root, ".git"), nil },
		configSet: func(string, string, string) error { return nil },
	})

	_, err := wtcmd.WorktreeCreator("plain-dir", &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("plain-directory collision error = %v", err)
	}
	if addCalls != 0 {
		t.Fatalf("git worktree add seam calls = %d, want 0", addCalls)
	}
	if info, statErr := os.Stat(dest); statErr != nil || !info.IsDir() {
		t.Fatalf("plain directory was modified or removed: info=%v err=%v", info, statErr)
	}
}
