package hook

// worktree_base_branch_test.go — SPEC-WORKTREE-BASEREF-001 M2 (card t313).
//
// Consumer 1: the SessionStart origin/HEAD alignment step. Covers
// AC-WBR-003 (unset), -004 (match), -005 (mismatch), -006 (fail-open),
// -015 (unresolvable), and both halves of -016 (firing point).
//
// Every test drives the seams in worktree_base_branch.go — no test touches the
// developer's real repository, and no test spawns git.

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// worktreeBaseBranchFakes records seam invocations for one test.
type worktreeBaseBranchFakes struct {
	primary       bool     // what the primary-checkout discriminant reports
	configured    string   // what the alignment-entry (configured-value) seam returns
	originHead    string   // what the origin/HEAD read seam returns
	originHeadErr error    // ... or its error
	resolvable    bool     // what the shared resolvability predicate reports
	setHeadErr    error    // what the write seam returns

	entryCalls      int      // alignment-entry read seam invocation count (AC-WBR-016)
	originHeadCalls int      // origin/HEAD read seam invocation count
	setHeadCalls    int      // write seam invocation count
	setHeadArgs     []string // values passed to the write seam
	resolvableArgs  []string // values passed to the resolvability predicate
	stderr          bytes.Buffer
}

// swapWorktreeBaseBranchSeams installs fakes for every seam and restores the
// real implementations on cleanup. Modelled on the seam-swap idiom already used
// by internal/cli/session_worktree.go.
func swapWorktreeBaseBranchSeams(t *testing.T, f *worktreeBaseBranchFakes) {
	t.Helper()
	origPrimary := worktreeBaseBranchInPrimaryCheckout
	origRead := worktreeBaseBranchReadConfig
	origHead := worktreeBaseBranchReadOriginHead
	origResolvable := WorktreeBaseBranchResolvable
	origSet := worktreeBaseBranchSetHead
	origErr := worktreeBaseBranchStderr

	worktreeBaseBranchInPrimaryCheckout = func() bool { return f.primary }
	worktreeBaseBranchReadConfig = func(string) string {
		f.entryCalls++
		return f.configured
	}
	worktreeBaseBranchReadOriginHead = func() (string, error) {
		f.originHeadCalls++
		return f.originHead, f.originHeadErr
	}
	WorktreeBaseBranchResolvable = func(branch string) bool {
		f.resolvableArgs = append(f.resolvableArgs, branch)
		return f.resolvable
	}
	worktreeBaseBranchSetHead = func(branch string) error {
		f.setHeadCalls++
		f.setHeadArgs = append(f.setHeadArgs, branch)
		return f.setHeadErr
	}
	worktreeBaseBranchStderr = &f.stderr

	t.Cleanup(func() {
		worktreeBaseBranchInPrimaryCheckout = origPrimary
		worktreeBaseBranchReadConfig = origRead
		worktreeBaseBranchReadOriginHead = origHead
		WorktreeBaseBranchResolvable = origResolvable
		worktreeBaseBranchSetHead = origSet
		worktreeBaseBranchStderr = origErr
	})
}

// stderrLines returns the non-empty lines the alignment step emitted.
func (f *worktreeBaseBranchFakes) stderrLines() []string {
	trimmed := strings.TrimSpace(f.stderr.String())
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// --- AC-WBR-003: unset / empty setting is byte-identical to today ---

func TestWorktreeBaseBranchAlignmentUnsetIsSilentNoOp(t *testing.T) {
	f := &worktreeBaseBranchFakes{primary: true, configured: "", originHead: "main", resolvable: true}
	swapWorktreeBaseBranchSeams(t, f)

	runWorktreeBaseAlignment(t.TempDir())

	if f.setHeadCalls != 0 {
		t.Errorf("empty setting: write seam invoked %d times, want 0", f.setHeadCalls)
	}
	if f.originHeadCalls != 0 {
		t.Errorf("empty setting: origin/HEAD read invoked %d times, want 0 "+
			"(REQ-WBR-005 forbids a git-metadata read on the empty path)", f.originHeadCalls)
	}
	if got := f.stderrLines(); len(got) != 0 {
		t.Errorf("empty setting: stderr = %q, want empty", got)
	}
}

// --- AC-WBR-004: matching setting produces no write and no output ---

func TestWorktreeBaseBranchAlignmentMatchIsSilentNoOp(t *testing.T) {
	f := &worktreeBaseBranchFakes{primary: true, configured: "develop", originHead: "develop", resolvable: true}
	swapWorktreeBaseBranchSeams(t, f)

	runWorktreeBaseAlignment(t.TempDir())

	if f.setHeadCalls != 0 {
		t.Errorf("match: write seam invoked %d times, want 0", f.setHeadCalls)
	}
	if got := f.stderrLines(); len(got) != 0 {
		t.Errorf("match: stderr = %q, want empty", got)
	}
}

// --- AC-WBR-005: mismatch produces the write and EXACTLY one notice line ---

func TestWorktreeBaseBranchAlignmentMismatchWritesAndAnnounces(t *testing.T) {
	f := &worktreeBaseBranchFakes{primary: true, configured: "develop", originHead: "main", resolvable: true}
	swapWorktreeBaseBranchSeams(t, f)

	runWorktreeBaseAlignment(t.TempDir())

	if f.setHeadCalls != 1 {
		t.Fatalf("mismatch: write seam invoked %d times, want exactly 1", f.setHeadCalls)
	}
	if f.setHeadArgs[0] != "develop" {
		t.Errorf("mismatch: write seam got %q, want \"develop\"", f.setHeadArgs[0])
	}
	lines := f.stderrLines()
	if len(lines) != 1 {
		t.Fatalf("mismatch: stderr has %d lines, want exactly 1 — two lines FAIL as surely as zero: %q", len(lines), lines)
	}
	for _, want := range []string{"main", "develop", "worktree_base_branch"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("notice line %q does not name %q (previous branch / new branch / the setting are all required)", lines[0], want)
		}
	}
}

// --- AC-WBR-015: unresolvable value writes nothing ---

func TestWorktreeBaseBranchAlignmentUnresolvableWritesNothing(t *testing.T) {
	f := &worktreeBaseBranchFakes{primary: true, configured: "no-such-branch", originHead: "main", resolvable: false}
	swapWorktreeBaseBranchSeams(t, f)

	runWorktreeBaseAlignment(t.TempDir())

	if f.setHeadCalls != 0 {
		t.Errorf("unresolvable: write seam invoked %d times, want 0 — refs/remotes/origin/HEAD must never name a ref that does not exist", f.setHeadCalls)
	}
	if len(f.resolvableArgs) != 1 || f.resolvableArgs[0] != "no-such-branch" {
		t.Errorf("unresolvable: resolvability predicate called with %v, want exactly one call with \"no-such-branch\"", f.resolvableArgs)
	}
	lines := f.stderrLines()
	if len(lines) != 1 {
		t.Fatalf("unresolvable: stderr has %d lines, want exactly 1: %q", len(lines), lines)
	}
	if !strings.Contains(lines[0], "no-such-branch") {
		t.Errorf("diagnostic line %q does not name the offending value", lines[0])
	}
}

// TestWorktreeBaseBranchAlignmentUnresolvableHandleStillAllows is the Handle-level
// half of AC-WBR-015: session start proceeds normally.
func TestWorktreeBaseBranchAlignmentUnresolvableHandleStillAllows(t *testing.T) {
	f := &worktreeBaseBranchFakes{primary: true, configured: "no-such-branch", originHead: "main", resolvable: false}
	swapWorktreeBaseBranchSeams(t, f)

	out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
		SessionID: "s1", ProjectDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Handle must be non-blocking on an unresolvable value, got: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}
	if f.setHeadCalls != 0 {
		t.Errorf("write seam invoked %d times via Handle, want 0", f.setHeadCalls)
	}
}

// --- AC-WBR-006: fail-open ---

func TestWorktreeBaseBranchAlignmentFailOpenOnGitError(t *testing.T) {
	f := &worktreeBaseBranchFakes{
		primary: true, configured: "develop", originHead: "main",
		resolvable: true, setHeadErr: exec.ErrNotFound,
	}
	swapWorktreeBaseBranchSeams(t, f)

	out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
		SessionID: "s1", ProjectDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Handle must return a nil error when the git seam fails, got: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}
}

func TestWorktreeBaseBranchAlignmentFailOpenOnOriginHeadError(t *testing.T) {
	f := &worktreeBaseBranchFakes{
		primary: true, configured: "develop",
		originHeadErr: exec.ErrNotFound, resolvable: true,
	}
	swapWorktreeBaseBranchSeams(t, f)

	runWorktreeBaseAlignment(t.TempDir())

	if f.setHeadCalls != 0 {
		t.Errorf("origin/HEAD read failure: write seam invoked %d times, want 0", f.setHeadCalls)
	}
	if got := f.stderrLines(); len(got) != 0 {
		t.Errorf("origin/HEAD read failure must be silent, got stderr %q", got)
	}
}

// --- AC-WBR-016 half 1: fires exactly once from the primary checkout ---
//
// The seam counted here is the ALIGNMENT-ENTRY seam — the configured-value read
// (acceptance.md AC-WBR-016 preamble; plan.md §A D3.2). It is NOT the
// origin/HEAD read and NOT the write seam, so the count is 1 for EVERY
// configured value, empty included.

func TestWorktreeBaseBranchFiresExactlyOnceFromPrimaryCheckout(t *testing.T) {
	for _, tc := range []struct {
		name       string
		configured string
	}{
		{"set value", "develop"},
		{"empty value", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &worktreeBaseBranchFakes{primary: true, configured: tc.configured, originHead: "main", resolvable: true}
			swapWorktreeBaseBranchSeams(t, f)

			out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
				SessionID: "s1", ProjectDir: t.TempDir(),
			})
			if err != nil || out == nil {
				t.Fatalf("Handle failed: out=%v err=%v", out, err)
			}
			if f.entryCalls != 1 {
				t.Errorf("alignment-entry read seam invoked %d times, want exactly 1 "+
					"(0 = the errgroup task was never registered; 2+ = registered twice or invoked outside the group)", f.entryCalls)
			}
		})
	}
}

// --- AC-WBR-016 half 2: never fires from a linked worktree ---

func TestWorktreeBaseBranchNotFiredFromLinkedWorktree(t *testing.T) {
	// The exact state AC-WBR-005 requires a write for: a resolvable value that
	// differs from what origin/HEAD names.
	f := &worktreeBaseBranchFakes{primary: false, configured: "develop", originHead: "main", resolvable: true}
	swapWorktreeBaseBranchSeams(t, f)

	out, err := NewSessionStartHandler(nil).Handle(context.Background(), &HookInput{
		SessionID: "s1", ProjectDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Handle must return nil error from a linked worktree, got: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}
	if f.entryCalls != 0 {
		t.Errorf("alignment-entry read seam invoked %d times from a linked worktree, want 0 "+
			"(the primary-checkout gate precedes the configured-value read)", f.entryCalls)
	}
	if f.setHeadCalls != 0 {
		t.Errorf("write seam invoked %d times from a linked worktree, want 0", f.setHeadCalls)
	}
	if got := f.stderrLines(); len(got) != 0 {
		t.Errorf("linked worktree must emit nothing, got stderr %q", got)
	}
}

// --- The shared resolvability predicate (REQ-WBR-009) ---

// TestWorktreeBaseBranchResolvableTreatsOnlyRcZeroAsResolvable pins the
// predicate against a real temporary repository. plan.md §B G7: a missing ref
// exits 128, NOT 1 — an rc == 1 test would misclassify every missing ref as an
// execution error.
func TestWorktreeBaseBranchResolvableTreatsOnlyRcZeroAsResolvable(t *testing.T) {
	repo := newTempRepoWithRemoteRef(t, "develop")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if !worktreeBaseBranchResolvableReal("develop") {
		t.Error("an existing remote-tracking branch must be resolvable")
	}
	if worktreeBaseBranchResolvableReal("no-such-branch") {
		t.Error("a missing remote-tracking branch must be unresolvable (git exits 128, not 1)")
	}
}

// newTempRepoWithRemoteRef builds a throwaway repository carrying
// refs/remotes/origin/<branch>, so the predicate can be exercised without
// touching the developer's repository or the network.
func newTempRepoWithRemoteRef(t *testing.T, branch string) string {
	t.Helper()
	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	if err := os.WriteFile(filepath.Join(repo, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "f.txt")
	run("commit", "-q", "-m", "init")
	// Fabricate the remote-tracking ref directly — no network, no remote.
	run("update-ref", "refs/remotes/origin/"+branch, "HEAD")
	return repo
}
