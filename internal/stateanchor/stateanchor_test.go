package stateanchor

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initGitRepo prepares dir as a minimal git repository — enough for
// ResolveGitDirs to answer (rev-parse needs init, not a commit).
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "init", "--quiet")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", dir, err, out)
	}
}

// TestResolve_ProjectDirWins is chain step 1: the runtime-reported project
// root beats everything else, so a session that cd'd away still anchors to
// its project (the GH #1694 repair).
func TestResolve_ProjectDirWins(t *testing.T) {
	t.Parallel()

	got := Resolve(Session{
		ProjectDir:  "/proj",
		OriginalCwd: "/primary",
		CurrentDir:  "/somewhere/visited",
	})
	if got != "/proj" {
		t.Errorf("Resolve() = %q, want /proj (chain step 1)", got)
	}
}

// TestResolve_OriginalCwdSecond is chain step 2: a worktree session without
// project_dir anchors to the primary checkout via original_cwd.
func TestResolve_OriginalCwdSecond(t *testing.T) {
	t.Parallel()

	got := Resolve(Session{
		OriginalCwd: "/primary",
		CurrentDir:  "/primary/.claude/worktrees/card-x",
	})
	if got != "/primary" {
		t.Errorf("Resolve() = %q, want /primary (chain step 2)", got)
	}
}

// TestResolve_GitWalkUpFromSubdirectory is chain step 3: with no stdin
// fields, a session cd'd into a subdirectory of a repository anchors to the
// repository's primary checkout — the git common directory's parent, the
// same root from any subdirectory.
func TestResolve_GitWalkUpFromSubdirectory(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)
	sub := filepath.Join(root, "deep", "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got := Resolve(Session{CurrentDir: sub})
	// git --path-format=absolute reports the symlink-resolved spelling
	// (/var/folders → /private/var/folders on macOS), so the comparison
	// normalizes the fixture root the same way (see also queueRootInsideTemp
	// in internal/cli for the same dual-spelling handling).
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	if got != root {
		t.Errorf("Resolve() = %q, want %q (git common-dir parent)", got, root)
	}
}

// TestResolve_NonGitDirectoryIsEmpty pins the REQ-SA-003 edge: outside any
// git repository the anchor is "" — the caller skips the state write (no
// project, no state) instead of anchoring to the visited directory.
func TestResolve_NonGitDirectoryIsEmpty(t *testing.T) {
	t.Parallel()

	got := Resolve(Session{CurrentDir: t.TempDir()})
	if got != "" {
		t.Errorf("Resolve() = %q, want \"\" outside a git repository", got)
	}
}

// TestResolve_EmptySessionIsEmpty pins the hermeticity guard: a context with
// no directory fields resolves to "" — the process cwd is deliberately NOT
// consulted, so a payload-less render can never reach the operator's real
// checkout through the resolver (learned in M2: the walk-up from the test
// process cwd landed on the real repository's board).
func TestResolve_EmptySessionIsEmpty(t *testing.T) {
	t.Parallel()

	if got := Resolve(Session{}); got != "" {
		t.Errorf("Resolve(empty) = %q, want \"\" (no directory context, no anchor)", got)
	}
}

// TestResolve_RelativeProjectDirFallsThrough pins the absolute-path half of
// REQ-SAV-001 (SPEC-STATE-ANCHOR-VALIDATE-001): a relative project_dir would
// be interpreted against the statusline process's cwd — the wrong reference
// point — so it is rejected and the chain falls through. With no follow-up
// fields the fall-through lands on "" (the REQ-SA-003 skip path).
func TestResolve_RelativeProjectDirFallsThrough(t *testing.T) {
	t.Parallel()

	got := Resolve(Session{ProjectDir: "relative/dir"})
	if got != "" {
		t.Errorf("Resolve() = %q, want \"\" (relative candidate rejected, chain empty)", got)
	}
}

// TestResolve_StaleProjectDirFallsThrough pins the existence half of
// REQ-SAV-001: an absolute path that no longer exists must not anchor —
// anchor-derived MkdirAll writes would resurrect it as a fresh directory
// tree (SPEC §1.1, the t510 F1 revival mechanism).
func TestResolve_StaleProjectDirFallsThrough(t *testing.T) {
	t.Parallel()

	stale := filepath.Join(t.TempDir(), "gone")
	if err := os.Mkdir(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(stale); err != nil {
		t.Fatal(err)
	}

	got := Resolve(Session{ProjectDir: stale})
	if got != "" {
		t.Errorf("Resolve() = %q, want \"\" (stale candidate rejected, chain empty)", got)
	}
}

// TestResolve_ProjectDirFileCandidateRejected pins the mode half of the
// predicate: an absolute path that EXISTS but is a FILE is rejected —
// os.Stat succeeding is not enough, the candidate must name a directory
// (REQ-SAV-001). This is the admission-condition test for the
// Stat-success-regardless-of-mode mutant.
func TestResolve_ProjectDirFileCandidateRejected(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "plainfile")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := Resolve(Session{ProjectDir: file})
	if got != "" {
		t.Errorf("Resolve() = %q, want \"\" (file candidate rejected, chain empty)", got)
	}
}

// TestResolve_InvalidProjectDirFallsThroughToOriginalCwd pins REQ-SAV-002's
// fall-through against the hard-"" mutant: an invalid chain step 1 must give
// step 2 its chance — a valid original_cwd rescues the session, and the
// exact fixture value is returned (equality, not merely non-empty, so
// fall-through is distinguishable from a hard-"" early return only when both
// would differ; here a hard-"" returns "" while fall-through returns the
// fixture path).
func TestResolve_InvalidProjectDirFallsThroughToOriginalCwd(t *testing.T) {
	t.Parallel()

	primary := t.TempDir()

	got := Resolve(Session{
		ProjectDir:  "/nonexistent/anchor-candidate",
		OriginalCwd: primary,
	})
	if got != primary {
		t.Errorf("Resolve() = %q, want %q (invalid step 1 falls through to valid step 2)", got, primary)
	}
}

// TestResolve_OriginalCwdInvalidFallsThrough pins chain step 2 under the
// same predicate (REQ-SAV-001/002): an original_cwd that is not an absolute
// existing directory is rejected and the chain continues — with nothing
// behind it, "" (REQ-SA-003).
func TestResolve_OriginalCwdInvalidFallsThrough(t *testing.T) {
	t.Parallel()

	got := Resolve(Session{OriginalCwd: "relative/cwd"})
	if got != "" {
		t.Errorf("Resolve() = %q, want \"\" (relative original_cwd rejected, chain empty)", got)
	}
}

// TestResolve_OriginalCwdValidReturned pins the step-2 happy path: a valid
// absolute existing directory as original_cwd is returned AS-IS — no
// EvalSymlinks rewriting of the value (REQ-SAV-006).
func TestResolve_OriginalCwdValidReturned(t *testing.T) {
	t.Parallel()

	primary := t.TempDir()

	got := Resolve(Session{OriginalCwd: primary})
	if got != primary {
		t.Errorf("Resolve() = %q, want %q (valid step 2 returned as-is)", got, primary)
	}
}

// TestFromDirectory_WorktreeResolvesPrimaryCheckout pins the one-root
// property the seam exists for: from a linked worktree the git common
// directory still points at the primary checkout, so the anchor is the same
// as from the primary itself.
func TestFromDirectory_WorktreeResolvesPrimaryCheckout(t *testing.T) {
	primary := t.TempDir()
	initGitRepo(t, primary)
	// A commit is required before a worktree can be added.
	cmd := exec.Command("git", "-C", primary, "commit", "--allow-empty", "--message", "init",
		"--quiet", "--no-verify", "--no-gpg-sign")
	cmd.Env = append(cmd.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	wt := filepath.Join(t.TempDir(), "wt")
	cmd = exec.Command("git", "-C", primary, "worktree", "add", "--quiet", wt, "-b", "wt-anchor-test")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}

	fromPrimary := FromDirectory(primary)
	fromWorktree := FromDirectory(wt)
	if fromPrimary == "" || fromWorktree == "" {
		t.Fatalf("resolution failed: primary=%q worktree=%q", fromPrimary, fromWorktree)
	}
	if fromPrimary != fromWorktree {
		t.Errorf("anchor differs by checkout: primary=%q worktree=%q — want one root", fromPrimary, fromWorktree)
	}
}

// TestFromDirectory_EmptyIsEmpty pins the guard: an empty directory resolves
// to "" without spawning git.
func TestFromDirectory_EmptyIsEmpty(t *testing.T) {
	t.Parallel()

	if got := FromDirectory(""); got != "" {
		t.Errorf("FromDirectory(\"\") = %q, want \"\"", got)
	}
}
