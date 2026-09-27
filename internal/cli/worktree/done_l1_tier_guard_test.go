package worktree

// L1 tier-guard reproduction tests (SPEC-WORKTREE-DONE-TIER-001, card t1073).
//
// Each test pins a loss shape from spec.md §D and asserts the POST-guard
// behavior, so every refusal cell fails RED against the unguarded tree.
// All git semantics run against REAL repositories created under t.TempDir()
// — never the project's own .claude/worktrees/ (plan §G). The guard's
// target-derived repo-root resolver is exercised through its DEFAULT path
// here: no seam is overridden anywhere in this file, including AC-010
// (CWD-independence mutation cell, plan-audit D6).

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// tierFixture is the sandbox for one real-git tier-guard scenario:
// base is the t.TempDir() root, repo the git repository under it.
// Keeping the repo in a subdirectory leaves room for sibling trees (the L2
// shape in AC-006 and the second linked worktree in AC-010) that sit OUTSIDE
// the repo and therefore outside any .claude/worktrees/ prefix.
type tierFixture struct {
	base string
	repo string
}

// tierRunGit runs a git subcommand in dir, failing the test on error.
func tierRunGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// newTierRepo builds a real git repository with one seed commit.
func newTierRepo(t *testing.T) tierFixture {
	t.Helper()
	base := t.TempDir()
	repo := filepath.Join(base, "main")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	tierRunGit(t, repo, "init", "-q")
	tierRunGit(t, repo, "config", "user.email", "tier-guard-test@example.com")
	tierRunGit(t, repo, "config", "user.name", "Tier Guard Test")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	tierRunGit(t, repo, "add", ".")
	tierRunGit(t, repo, "commit", "-q", "-m", "seed")
	return tierFixture{base: base, repo: repo}
}

// addTierWorktree creates a linked worktree on a new branch at HEAD.
func addTierWorktree(t *testing.T, f tierFixture, path, branch string) {
	t.Helper()
	tierRunGit(t, f.repo, "worktree", "add", "-b", branch, "--", path)
}

// withTierTestEnv wires the real WorktreeProvider for repo and stubs the
// launch-ledger seam so the removal path never touches the developer's
// ~/.moai profile state. The tier-guard resolver seam is deliberately NOT
// stubbed: every cell exercises the DEFAULT target-derived mechanism.
func withTierTestEnv(t *testing.T, repo string) {
	t.Helper()
	origProvider := WorktreeProvider
	WorktreeProvider = git.NewWorktreeManager(repo)
	t.Cleanup(func() { WorktreeProvider = origProvider })

	origPrune := pruneLaunchLedgerFn
	pruneLaunchLedgerFn = func() ([]string, error) { return nil, nil }
	t.Cleanup(func() { pruneLaunchLedgerFn = origPrune })
}

// executeDoneForTierGuard runs the done command with captured output
// buffers and returns the command error (nil == exit 0).
func executeDoneForTierGuard(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newDoneCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// realPath resolves the macOS /var/folders -> /private/var symlink so path
// substrings asserted against refusal messages match what git (and the
// guard's canonicalization) reports.
func realPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", path, err)
	}
	return resolved
}

// assertL1Refusal asserts the common post-guard contract: non-nil error
// carrying the tier sentinel, and the L1 tree still on disk.
func assertL1Refusal(t *testing.T, err error, l1 string) {
	t.Helper()
	if err == nil {
		t.Fatal("done must refuse an L1 session worktree (nil error = removal proceeded)")
	}
	if !strings.Contains(err.Error(), "L1_SESSION_WORKTREE") {
		t.Errorf("refusal must carry the L1_SESSION_WORKTREE sentinel, got: %v", err)
	}
	if _, statErr := os.Stat(l1); statErr != nil {
		t.Errorf("L1 tree must survive the refusal, got: %v", statErr)
	}
}

// assertL1Gone asserts the RED-side observation: the tree was removed.
func assertL1Gone(t *testing.T, l1 string) {
	t.Helper()
	if _, statErr := os.Stat(l1); !os.IsNotExist(statErr) {
		t.Errorf("expected %s to be removed, stat error: %v", l1, statErr)
	}
}

// TestDoneL1TierGuard_RefusesCleanL1 pins loss shape A1 (AC-001): a cooled,
// clean L1 tree under .claude/worktrees/ must be REFUSED by done, not
// silently removed. Today (RED) the tree is removed with exit 0.
func TestDoneL1TierGuard_RefusesCleanL1(t *testing.T) {
	f := newTierRepo(t)
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-a1")
	addTierWorktree(t, f, l1, "feature/SPEC-TIER-A1")
	withTierTestEnv(t, f.repo)

	err := executeDoneForTierGuard(t, "feature/SPEC-TIER-A1")

	assertL1Refusal(t, err, l1)
}

// TestDoneL1TierGuard_RefusesDirtyL1ByTier pins loss shape A2 (AC-002): a
// dirty L1 tree must be refused BY THE TIER GUARD, not merely by git's own
// dirty check — the refusal message must carry the tier sentinel.
func TestDoneL1TierGuard_RefusesDirtyL1ByTier(t *testing.T) {
	f := newTierRepo(t)
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-a2")
	addTierWorktree(t, f, l1, "feature/SPEC-TIER-A2")
	if err := os.WriteFile(filepath.Join(l1, "README.md"), []byte("uncommitted\n"), 0o644); err != nil {
		t.Fatalf("dirty the tree: %v", err)
	}
	withTierTestEnv(t, f.repo)

	err := executeDoneForTierGuard(t, "feature/SPEC-TIER-A2")

	assertL1Refusal(t, err, l1)
}

// TestDoneL1TierGuard_ForceDoesNotBypassL1 pins loss shape A3 (AC-003): the
// SPEC's point — --force must NOT destroy a dirty L1 tree. Today (RED)
// --force removes the tree and the uncommitted work with it.
func TestDoneL1TierGuard_ForceDoesNotBypassL1(t *testing.T) {
	f := newTierRepo(t)
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-a3")
	addTierWorktree(t, f, l1, "feature/SPEC-TIER-A3")
	dirty := filepath.Join(l1, "README.md")
	if err := os.WriteFile(dirty, []byte("precious uncommitted work\n"), 0o644); err != nil {
		t.Fatalf("dirty the tree: %v", err)
	}
	withTierTestEnv(t, f.repo)

	err := executeDoneForTierGuard(t, "--force", "feature/SPEC-TIER-A3")

	assertL1Refusal(t, err, l1)
	content, readErr := os.ReadFile(dirty)
	if readErr != nil {
		t.Fatalf("uncommitted work must survive the refusal, got: %v", readErr)
	}
	if !strings.Contains(string(content), "precious") {
		t.Errorf("uncommitted content must be intact, got: %q", content)
	}
}

// TestDoneL1TierGuard_RefusesIgnoredOnlyDirtyL1 pins loss shape C (AC-004):
// an L1 tree whose only dirtiness is an IGNORED file is removed today
// WITHOUT --force (git's own refusal never sees ignored files) — the one
// unconditional-loss shape. The guard must refuse and keep the file.
func TestDoneL1TierGuard_RefusesIgnoredOnlyDirtyL1(t *testing.T) {
	f := newTierRepo(t)
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-c")
	addTierWorktree(t, f, l1, "feature/SPEC-TIER-C")
	exclude := filepath.Join(f.repo, ".git", "info", "exclude")
	fh, err := os.OpenFile(exclude, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open exclude: %v", err)
	}
	if _, err := fh.WriteString("ignored.env\n"); err != nil {
		t.Fatalf("append exclude: %v", err)
	}
	_ = fh.Close()
	ignored := filepath.Join(l1, "ignored.env")
	if err := os.WriteFile(ignored, []byte("local secret\n"), 0o644); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}
	withTierTestEnv(t, f.repo)

	err = executeDoneForTierGuard(t, "feature/SPEC-TIER-C")

	assertL1Refusal(t, err, l1)
	if _, statErr := os.Stat(ignored); statErr != nil {
		t.Errorf("ignored file must survive the refusal, got: %v", statErr)
	}
}

// TestDoneL1TierGuard_DeleteBranchRefusedPreRemoval pins loss shape A4
// (AC-005): done --delete-branch on an L1 tree must refuse BEFORE removal —
// today (RED) the tree is removed while the unmerged branch survives,
// a half-completed disposal.
func TestDoneL1TierGuard_DeleteBranchRefusedPreRemoval(t *testing.T) {
	f := newTierRepo(t)
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-a4")
	branch := "feature/SPEC-TIER-A4"
	addTierWorktree(t, f, l1, branch)
	// Commit inside the L1 tree so the branch is genuinely unmerged
	// (a branch still at HEAD would pass git's safe -d).
	tierRunGit(t, l1, "commit", "-q", "--allow-empty", "-m", "unmerged card work")
	withTierTestEnv(t, f.repo)

	err := executeDoneForTierGuard(t, "--delete-branch", branch)

	assertL1Refusal(t, err, l1)
	if out := tierRunGit(t, f.repo, "branch", "--list", branch); strings.TrimSpace(out) == "" {
		t.Errorf("branch %s must survive the pre-removal refusal", branch)
	}
}

// TestDoneL1TierGuard_L2FlowUnchanged pins the B direction (AC-006): the
// deployed sync flow `done SPEC-{ID} --auto --delete-branch` against an L2
// tree OUTSIDE .claude/worktrees/ must keep removing the tree and the
// (merged) branch with exit 0 and no error. This cell is expected to pass
// BOTH before and after the guard — it is the no-mutation direction
// (memory lesson t1028: the two mutant directions are caught separately).
func TestDoneL1TierGuard_L2FlowUnchanged(t *testing.T) {
	f := newTierRepo(t)
	l2 := filepath.Join(f.base, "l2trees", "tier-b") // outside .claude/worktrees/
	addTierWorktree(t, f, l2, "feature/SPEC-TIER-B") // branch at HEAD == merged
	withTierTestEnv(t, f.repo)

	if err := executeDoneForTierGuard(t, "--auto", "--delete-branch", "SPEC-TIER-B"); err != nil {
		t.Fatalf("deployed L2 flow must keep succeeding, got: %v", err)
	}
	assertL1Gone(t, l2)
	if out := tierRunGit(t, f.repo, "branch", "--list", "feature/SPEC-TIER-B"); strings.TrimSpace(out) != "" {
		t.Errorf("merged branch must be deleted by --delete-branch, still listed: %s", out)
	}
}

// TestDoneL1TierGuard_RefusalMessageBothModes pins AC-007: the refusal
// surfaces with non-zero exit in BOTH modes and the message names the path,
// states L1 trees are session-scoped, and carries the two remedy forms
// (session-end keep/remove prompt; git worktree unlock + git worktree
// remove), in the style of lockGuidance.
func TestDoneL1TierGuard_RefusalMessageBothModes(t *testing.T) {
	for _, mode := range []struct {
		name string
		args []string
	}{
		{name: "interactive", args: nil},
		{name: "auto", args: []string{"--auto"}},
	} {
		t.Run(mode.name, func(t *testing.T) {
			f := newTierRepo(t)
			l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-msg")
			addTierWorktree(t, f, l1, "feature/SPEC-TIER-MSG")
			withTierTestEnv(t, f.repo)

			args := append(append([]string{}, mode.args...), "feature/SPEC-TIER-MSG")
			err := executeDoneForTierGuard(t, args...)

			if err == nil {
				t.Fatal("both modes must exit non-zero on an L1 target")
			}
			msg := err.Error()
			for _, want := range []string{
				realPath(t, l1), // the canonicalized target path
				"L1_SESSION_WORKTREE",
				"session worktree",
				"session-scoped",
				"session-end",
				"git worktree unlock " + realPath(t, l1),
				"git worktree remove " + realPath(t, l1),
			} {
				if !strings.Contains(msg, want) {
					t.Errorf("refusal message must contain %q, got: %s", want, msg)
				}
			}
		})
	}
}

// TestDoneL1TierGuard_PredicateDirections pins the two mutant directions at
// the predicate level (plan M2e, memory lesson t1028): the predicate FIRES
// on an L1-shaped path (positive control) and does NOT fire on an L2-shaped
// path outside .claude/worktrees/ (no-mutation control). Real git, DEFAULT
// resolver — no seam override.
func TestDoneL1TierGuard_PredicateDirections(t *testing.T) {
	f := newTierRepo(t)
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-pred")
	addTierWorktree(t, f, l1, "feature/SPEC-TIER-PRED")
	l2 := filepath.Join(f.base, "l2trees", "tier-pred-l2")
	addTierWorktree(t, f, l2, "feature/SPEC-TIER-PRED-L2")

	if !isL1WorktreePath(l1) {
		t.Error("predicate must fire on an L1 path under <mainRoot>/.claude/worktrees/ (positive control)")
	}
	if isL1WorktreePath(l2) {
		t.Error("predicate must not fire on an L2 path outside .claude/worktrees/ (no-mutation control)")
	}
}

// TestDoneL1TierGuard_CWDIndependent pins AC-010 (plan-audit D1/D6): with
// the PROCESS CWD inside a SECOND linked worktree, done must STILL refuse
// the L1 target. The resolver is the DEFAULT target-derived mechanism — no
// seam override (plan.md M2f mandate). A CWD-anchored prefix would compute
// <secondWorktree>/.claude/worktrees/ and fail open exactly here while
// passing AC-001.
func TestDoneL1TierGuard_CWDIndependent(t *testing.T) {
	f := newTierRepo(t)
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "tier-cwd")
	addTierWorktree(t, f, l1, "feature/SPEC-TIER-CWD")
	second := filepath.Join(f.base, "second")
	addTierWorktree(t, f, second, "feature/SPEC-TIER-SECOND")
	withTierTestEnv(t, f.repo)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(second); err != nil {
		t.Fatalf("chdir into the second linked worktree: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	err = executeDoneForTierGuard(t, "feature/SPEC-TIER-CWD")

	assertL1Refusal(t, err, l1)
}
