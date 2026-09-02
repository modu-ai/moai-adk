package cli

// Tests for the launcher's existing-branch worktree creation path (card t295):
// flag parsing, no-op behavior without the flag, seam-level wiring, and a
// real-git integration test of the materialization + registry registration.

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// errBranchMaterializeFailed is the sentinel the materialize seam returns to
// prove the launcher propagates creation failures.
var errBranchMaterializeFailed = errors.New("materialize failed (sentinel)")

func TestSplitWorktreeBranchFlag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		args     []string
		wantHas  bool
		want     string
		wantErr  bool
		wantRest []string
	}{
		{name: "absent", args: []string{"-w", "feat"}, wantHas: false, wantRest: []string{"-w", "feat"}},
		{name: "two-token", args: []string{"-w", "feat", "--branch", "develop"}, wantHas: true, want: "develop", wantRest: []string{"-w", "feat"}},
		{name: "joined", args: []string{"--branch=develop", "-w", "feat"}, wantHas: true, want: "develop", wantRest: []string{"-w", "feat"}},
		{name: "after-dashdash-untouched", args: []string{"-w", "feat", "--", "--branch", "develop"}, wantHas: false, wantRest: []string{"-w", "feat", "--", "--branch", "develop"}},
		{name: "missing-value", args: []string{"-w", "feat", "--branch"}, wantHas: true, wantErr: true},
		{name: "joined-empty-value", args: []string{"--branch=", "-w", "feat"}, wantHas: true, wantErr: true},
		{name: "option-shaped-value", args: []string{"--branch", "-x", "-w", "feat"}, wantHas: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rest, branch, has, err := splitWorktreeBranchFlag(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if has != tt.wantHas {
				t.Fatalf("has = %v, want %v", has, tt.wantHas)
			}
			if branch != tt.want {
				t.Fatalf("branch = %q, want %q", branch, tt.want)
			}
			if strings.Join(rest, "\x00") != strings.Join(tt.wantRest, "\x00") {
				t.Fatalf("rest = %v, want %v", rest, tt.wantRest)
			}
		})
	}
}

func TestResolveWorktreeExistingBranch_NoFlagIsNoop(t *testing.T) {
	t.Parallel()
	original := launcherWorktreeMaterialize
	launcherWorktreeMaterialize = func(string, string, string, io.Writer) error {
		t.Error("materialize must not run without --branch")
		return nil
	}
	t.Cleanup(func() { launcherWorktreeMaterialize = original })

	args := []string{"-w", "feat", "--permission-mode", "auto"}
	got, err := resolveWorktreeExistingBranch(args, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(got, "\x00") != strings.Join(args, "\x00") {
		t.Fatalf("args mutated without --branch: %v", got)
	}
}

// branchMaterializeRecorder captures seam calls for wiring assertions.
type branchMaterializeRecorder struct {
	calls   []string // "projectRoot|name|branch"
	failErr error
}

func (r *branchMaterializeRecorder) materialize(projectRoot, name, branch string, warn io.Writer) error {
	r.calls = append(r.calls, projectRoot+"|"+name+"|"+branch)
	return r.failErr
}

func TestResolveWorktreeExistingBranch_WiresAndStrips(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	originalFind := findProjectRootFn
	originalMat := launcherWorktreeMaterialize
	rec := &branchMaterializeRecorder{}
	findProjectRootFn = func() (string, error) { return dir, nil }
	launcherWorktreeMaterialize = rec.materialize
	t.Cleanup(func() {
		findProjectRootFn = originalFind
		launcherWorktreeMaterialize = originalMat
	})

	got, err := resolveWorktreeExistingBranch(
		[]string{"-w", "develop", "--branch", "develop", "-c"}, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.calls) != 1 {
		t.Fatalf("materialize calls = %d, want 1", len(rec.calls))
	}
	want := dir + "|develop|develop"
	if rec.calls[0] != want {
		t.Fatalf("materialize args = %q, want %q", rec.calls[0], want)
	}
	if strings.Join(got, "\x00") != strings.Join([]string{"-w", "develop", "-c"}, "\x00") {
		t.Fatalf("--branch tokens not stripped: %v", got)
	}
}

func TestResolveWorktreeExistingBranch_RejectsBadUsage(t *testing.T) {
	t.Parallel()
	originalFind := findProjectRootFn
	originalMat := launcherWorktreeMaterialize
	rec := &branchMaterializeRecorder{}
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	launcherWorktreeMaterialize = rec.materialize
	t.Cleanup(func() {
		findProjectRootFn = originalFind
		launcherWorktreeMaterialize = originalMat
	})

	tests := []struct {
		name string
		args []string
	}{
		{name: "no -w value", args: []string{"--branch", "develop"}},
		{name: "bare -w", args: []string{"-w", "--branch", "develop"}},
		{name: "absolute -w", args: []string{"-w", "/tmp/somewhere", "--branch", "develop"}},
		{name: "separator in name", args: []string{"-w", "a/b", "--branch", "develop"}},
		{name: "dot name", args: []string{"-w", "..", "--branch", "develop"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveWorktreeExistingBranch(tt.args, os.Stderr)
			if err == nil {
				t.Fatalf("expected error for %v", tt.args)
			}
			if len(rec.calls) != 0 {
				t.Fatalf("materialize must not run on invalid usage, got %v", rec.calls)
			}
		})
	}
}

func TestResolveWorktreeExistingBranch_MaterializeErrorPropagates(t *testing.T) {
	t.Parallel()
	originalFind := findProjectRootFn
	originalMat := launcherWorktreeMaterialize
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	launcherWorktreeMaterialize = func(string, string, string, io.Writer) error {
		return errBranchMaterializeFailed
	}
	t.Cleanup(func() {
		findProjectRootFn = originalFind
		launcherWorktreeMaterialize = originalMat
	})

	_, err := resolveWorktreeExistingBranch([]string{"-w", "feat", "--branch", "develop"}, os.Stderr)
	if err != errBranchMaterializeFailed {
		t.Fatalf("err = %v, want %v", err, errBranchMaterializeFailed)
	}
}

// TestLauncherWorktreeMaterializeReal_Integration runs the real materializer
// against a throwaway repository and pins the contract: the tree exists with
// the existing branch checked out, the registry carries the entry, a missing
// branch fails without creating anything, and a second run re-enters
// idempotently.
func TestLauncherWorktreeMaterializeReal_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("real-git integration test")
	}
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve symlinks: %v", err)
	}
	runTestGit(t, repo, "init", "-b", "main")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test User")
	if werr := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# Test\n"), 0o644); werr != nil {
		t.Fatal(werr)
	}
	runTestGit(t, repo, "add", "README.md")
	runTestGit(t, repo, "commit", "-m", "init")
	runTestGit(t, repo, "branch", "develop")

	var notice strings.Builder
	if merr := launcherWorktreeMaterializeReal(repo, "develop", "develop", &notice); merr != nil {
		t.Fatalf("materialize: %v", merr)
	}

	treePath := filepath.Join(repo, ".claude", "worktrees", "develop")
	if _, statErr := os.Stat(treePath); statErr != nil {
		t.Fatalf("worktree not created at %s: %v", treePath, statErr)
	}
	out, cmdErr := exec.Command("git", "-C", treePath, "symbolic-ref", "--short", "HEAD").Output()
	if cmdErr != nil {
		t.Fatalf("read tree branch: %v", cmdErr)
	}
	if got := strings.TrimSpace(string(out)); got != "develop" {
		t.Fatalf("tree branch = %q, want develop", got)
	}

	// Registry: the entry must be in the same state file the hook maintains.
	data, readErr := os.ReadFile(filepath.Join(repo, ".moai", "state", "worktrees.json"))
	if readErr != nil {
		t.Fatalf("read registry: %v", readErr)
	}
	var entries []hook.WorktreeEntry
	if jsonErr := json.Unmarshal(data, &entries); jsonErr != nil {
		t.Fatalf("parse registry: %v", jsonErr)
	}
	if len(entries) != 1 || entries[0].Path != treePath || entries[0].Branch != "develop" {
		t.Fatalf("registry entries = %+v, want one entry for %s on develop", entries, treePath)
	}

	// Missing branch: hard error, nothing created.
	if merr := launcherWorktreeMaterializeReal(repo, "typo-tree", "nonexistent-branch", &notice); merr == nil {
		t.Fatal("expected error for nonexistent branch")
	}
	if _, statErr := os.Stat(filepath.Join(repo, ".claude", "worktrees", "typo-tree")); statErr == nil {
		t.Fatal("typo-tree must not be created for a missing branch")
	}

	// Idempotent re-entry: existing tree is reported and reused.
	notice.Reset()
	if merr := launcherWorktreeMaterializeReal(repo, "develop", "develop", &notice); merr != nil {
		t.Fatalf("re-entry: %v", merr)
	}
	if !strings.Contains(notice.String(), "already exists") {
		t.Fatalf("re-entry notice = %q, want an already-exists line", notice.String())
	}
}
