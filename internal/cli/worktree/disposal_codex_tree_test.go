package worktree

// disposal_codex_tree_test.go — the disposal paths that can reach an L1 tree
// `moai codex -w` created, measured against Claude-created L1 trees in the
// same four states (unmerged commit, uncommitted change, live lock, clean and
// merged). Every cell runs against a REAL repository under t.TempDir(); the
// only seams touched are the ones that would otherwise read the developer's
// own checkout (the lock listing and the protected-root lookup) and the
// launch-ledger prune.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

// disposalState names one of the four tree states the cells compare.
type disposalState string

const (
	stateUnmerged disposalState = "unmerged"
	stateDirty    disposalState = "dirty"
	stateLocked   disposalState = "locked"
	stateClean    disposalState = "clean"
)

var disposalStates = []disposalState{stateUnmerged, stateDirty, stateLocked, stateClean}

// disposalTree is one fixture tree: who created it and which state it holds.
type disposalTree struct {
	creator string // "codex" or "claude"
	state   disposalState
	path    string
	branch  string
}

// disposalFixture is one repository holding all eight trees.
type disposalFixture struct {
	repo  string
	base  string
	trees []disposalTree
}

func disposalGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %s %v: %v\n%s", dir, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// lockReasonFor renders the lock each creator writes: the Codex launcher's
// own reason, and the reason Claude Code writes at EnterWorktree.
func lockReasonFor(creator, name string, pid int) string {
	if creator == "codex" {
		return session.CodexAnchorLockReason(name, pid, "")
	}
	return "claude session " + name + " (pid " + strconv.Itoa(pid) + ")"
}

// newDisposalFixture builds the repository and its eight L1 trees under
// <repo>/.claude/worktrees/, then wires the worktree CLI to it.
func newDisposalFixture(t *testing.T) *disposalFixture {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	disposalGit(t, repo, "init", "-q")
	disposalGit(t, repo, "config", "user.email", "disposal-test@example.com")
	disposalGit(t, repo, "config", "user.name", "Disposal Test")
	disposalGit(t, repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	disposalGit(t, repo, "add", "README.md")
	disposalGit(t, repo, "commit", "-q", "-m", "seed")
	f := &disposalFixture{repo: repo, base: disposalGit(t, repo, "branch", "--show-current")}

	for _, creator := range []string{"codex", "claude"} {
		for _, state := range disposalStates {
			name := creator + "-" + string(state)
			tree := disposalTree{
				creator: creator, state: state,
				path:   filepath.Join(repo, ".claude", "worktrees", name),
				branch: "WT-" + name,
			}
			disposalGit(t, repo, "worktree", "add", "-q", "-b", tree.branch, tree.path, f.base)
			switch state {
			case stateUnmerged:
				if err := os.WriteFile(filepath.Join(tree.path, "work.txt"), []byte("committed work\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				disposalGit(t, tree.path, "add", "work.txt")
				disposalGit(t, tree.path, "commit", "-q", "-m", "unintegrated work")
			case stateDirty:
				if err := os.WriteFile(filepath.Join(tree.path, "draft.txt"), []byte("uncommitted\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			case stateLocked:
				disposalGit(t, repo, "worktree", "lock", "--reason", lockReasonFor(creator, name, os.Getpid()), tree.path)
			}
			f.trees = append(f.trees, tree)
		}
	}

	// The lock listing and the protected-root lookup would otherwise read the
	// checkout the test binary runs in.
	origCmd := gitWorktreeCmd
	gitWorktreeCmd = func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "worktree" {
			args = append([]string{"-C", repo}, args...)
		}
		return origCmd(args...)
	}
	origRoot := gitRepoRootFunc
	gitRepoRootFunc = func() (string, error) { return repo, nil }
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	withTierTestEnv(t, repo)
	t.Cleanup(func() {
		gitWorktreeCmd = origCmd
		gitRepoRootFunc = origRoot
		// Unlock so t.TempDir() cleanup is not refused by nothing — git keeps
		// no state outside the temp dir, but a clean exit is cheaper to read.
		for _, tr := range f.trees {
			_ = exec.Command("git", "-C", repo, "worktree", "unlock", tr.path).Run()
		}
	})
	return f
}

// outcome is what one disposal path did to one tree: whether the tree is
// still on disk, and the reason class the path reported for it.
type outcome struct {
	kept   bool
	reason string
}

func treeExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// assertParity checks the headline property: for every state, the Codex tree
// and the Claude tree came out the same.
func assertParity(t *testing.T, got map[string]outcome) {
	t.Helper()
	for _, state := range disposalStates {
		codex, claude := got["codex-"+string(state)], got["claude-"+string(state)]
		if codex != claude {
			t.Errorf("state %s: codex tree %+v differs from claude tree %+v", state, codex, claude)
		}
	}
}

// TestWorktreeDisposalRefusesUnintegratedCodexTree is AC-DHR-008's
// worktree-package half: `done` keeps its L1 refusal with or without --force,
// `clean --stale --yes` keeps the three trees that hold something and removes
// only the clean merged one, and `remove` refuses a lock-anchored tree naming
// the lock and its holder — each with the Codex tree matching the Claude one.
func TestWorktreeDisposalRefusesUnintegratedCodexTree(t *testing.T) {
	t.Run("done", func(t *testing.T) {
		f := newDisposalFixture(t)
		got := map[string]outcome{}
		for _, tr := range f.trees {
			err := executeDoneForTierGuard(t, tr.branch)
			if err == nil || !strings.Contains(err.Error(), "L1_SESSION_WORKTREE") {
				t.Errorf("%s: done must refuse with L1_SESSION_WORKTREE, got %v", tr.path, err)
			}
			got[tr.creator+"-"+string(tr.state)] = outcome{kept: treeExists(tr.path), reason: "L1_SESSION_WORKTREE"}
			if !treeExists(tr.path) {
				t.Errorf("%s: tree removed by done", tr.path)
			}
		}
		assertParity(t, got)
	})

	t.Run("done_force", func(t *testing.T) {
		f := newDisposalFixture(t)
		got := map[string]outcome{}
		for _, tr := range f.trees {
			err := executeDoneForTierGuard(t, "--force", tr.branch)
			if err == nil || !strings.Contains(err.Error(), "L1_SESSION_WORKTREE") {
				t.Errorf("%s: done --force must still refuse with L1_SESSION_WORKTREE, got %v", tr.path, err)
			}
			got[tr.creator+"-"+string(tr.state)] = outcome{kept: treeExists(tr.path), reason: "L1_SESSION_WORKTREE"}
			if !treeExists(tr.path) {
				t.Errorf("%s: tree removed by done --force", tr.path)
			}
		}
		assertParity(t, got)
	})

	t.Run("clean", func(t *testing.T) {
		f := newDisposalFixture(t)
		out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true", "base": f.base})
		if err != nil {
			t.Fatalf("clean --stale --yes: %v\n%s", err, out)
		}
		wantReason := map[disposalState]string{
			stateUnmerged: "branch has commits not in " + f.base,
			stateDirty:    "uncommitted or untracked changes",
			stateLocked:   "(source: lock)",
		}
		got := map[string]outcome{}
		for _, tr := range f.trees {
			o := outcome{kept: treeExists(tr.path)}
			for _, line := range strings.Split(out, "\n") {
				if strings.Contains(line, "Keeping "+tr.path+" ") {
					o.reason = line[strings.Index(line, "]: ")+3:]
				}
			}
			if tr.state == stateClean {
				if o.kept {
					t.Errorf("%s: clean merged tree was kept (%q)", tr.path, o.reason)
				}
			} else {
				if !o.kept {
					t.Errorf("%s: tree holding %s was removed", tr.path, tr.state)
				}
				if !strings.Contains(o.reason, wantReason[tr.state]) {
					t.Errorf("%s: keep reason %q lacks %q", tr.path, o.reason, wantReason[tr.state])
				}
			}
			// Reason texts embed pids and the tree name; compare their class.
			o.reason = string(tr.state)
			if !o.kept && tr.state == stateClean {
				o.reason = "removed"
			}
			got[tr.creator+"-"+string(tr.state)] = o
		}
		assertParity(t, got)
	})

	t.Run("remove", func(t *testing.T) {
		f := newDisposalFixture(t)
		got := map[string]outcome{}
		for _, tr := range f.trees {
			combined, err := runRemoveCmd(t, tr.path)
			o := outcome{kept: treeExists(tr.path)}
			switch tr.state {
			case stateLocked:
				if err == nil {
					t.Fatalf("%s: remove must refuse a lock-anchored tree", tr.path)
				}
				msg := err.Error()
				if !strings.Contains(msg, "ANCHORED_SESSIONS_PRESENT") || !strings.Contains(msg, "source: lock") {
					t.Errorf("%s: refusal must name the anchor source (lock), got: %v", tr.path, err)
				}
				if !strings.Contains(msg, "pid "+strconv.Itoa(os.Getpid())) {
					t.Errorf("%s: refusal must name the holder pid %d, got: %v", tr.path, os.Getpid(), err)
				}
				o.reason = "anchored-by-lock"
			case stateDirty:
				if err == nil {
					t.Errorf("%s: remove must refuse a tree with uncommitted changes (output %q)", tr.path, combined)
				}
				o.reason = "dirty"
			default:
				// Explicit removal keeps its integration-state semantics: a
				// clean tree goes, committed-but-unmerged or not.
				if err != nil {
					t.Errorf("%s: remove of a clean unanchored tree failed: %v", tr.path, err)
				}
				o.reason = "removed"
			}
			if (tr.state == stateLocked || tr.state == stateDirty) && !o.kept {
				t.Errorf("%s: refused tree is gone", tr.path)
			}
			got[tr.creator+"-"+string(tr.state)] = o
		}
		assertParity(t, got)
	})
}
