package homestate

// noncanonical_tree_guard_test.go — t961.
//
// The package-layer counterpart of the CLI gate landed by t952. That gate lives
// in internal/cli and therefore protects exactly one caller: homeStateRunner.
// The hazard it names, however, belongs to this package — a mutation addressed
// by a linked worktree is resolved back to the primary checkout by
// CanonicalProjectRoot, while MOAI_HOME moves only the TARGET, so a caller who
// isolates its home and works from a worktree still mutates the primary
// checkout's live state and nothing in the result contradicts that belief.
//
// The fixture is the same discriminant t952 used: a real repository plus a real
// linked worktree, so the caller's tree and the canonical root are necessarily
// different paths. No migration is ever executed here — the refusal is asserted
// before any mutation, and the marker file is the mutation being watched.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// twoTreeFixture builds a git repository plus a linked worktree and returns
// (primary, worktree). It skips rather than fails when git is unavailable so
// the suite stays runnable on a bare machine.
func twoTreeFixture(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	if err := os.MkdirAll(primary, 0o700); err != nil {
		t.Fatal(err)
	}
	runGit := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t.t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t.t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	runGit(primary, "init", "-q")
	runGit(primary, "commit", "-q", "--allow-empty", "-m", "base")

	worktree := filepath.Join(base, "linked")
	runGit(primary, "worktree", "add", "-q", "-b", "linked-branch", worktree)
	return primary, worktree
}

// isolateHome points MOAI_HOME at a private directory, which is what a caller
// who believes a worktree plus an isolated home is fully isolated would do.
func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "isolated-home"))
}

// TestT961_FixtureSplitsCallerTreeFromCanonicalRoot is the premise assertion.
// Every other test in this file is meaningless if the fixture does not actually
// place the caller's tree and the canonical root at different paths, so this
// states it first and fails loudly when the fixture stops discriminating.
func TestT961_FixtureSplitsCallerTreeFromCanonicalRoot(t *testing.T) {
	primary, worktree := twoTreeFixture(t)
	isolateHome(t)

	canonical := CanonicalProjectRoot(worktree)
	resolvedPrimary, err := filepath.EvalSymlinks(primary)
	if err != nil {
		t.Fatal(err)
	}
	resolvedWorktree, err := filepath.EvalSymlinks(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if canonical != resolvedPrimary {
		t.Fatalf("fixture premise: canonical root %q should be the primary %q", canonical, resolvedPrimary)
	}
	if canonical == resolvedWorktree {
		t.Fatalf("fixture premise: canonical root must differ from the caller's worktree %q", resolvedWorktree)
	}

	// And both trees address the SAME marker file — that shared address is the
	// asymmetry itself, not an accident of the fixture.
	fromWorktree, err := MigrationBarrierPath(worktree)
	if err != nil {
		t.Fatal(err)
	}
	fromPrimary, err := MigrationBarrierPath(primary)
	if err != nil {
		t.Fatal(err)
	}
	if fromWorktree != fromPrimary {
		t.Fatalf("fixture premise: both trees must address one marker; worktree=%q primary=%q", fromWorktree, fromPrimary)
	}
}

// TestT961_InstallMigrationMarkerLockedRefusesNonCanonicalCaller is the
// acceptance criterion for the package-layer gate. The marker is this package's
// own statement that a mutation is about to happen, so refusing there is what
// protects a caller that never passes through the CLI runner.
func TestT961_InstallMigrationMarkerLockedRefusesNonCanonicalCaller(t *testing.T) {
	primary, worktree := twoTreeFixture(t)
	isolateHome(t)

	release, err := InstallMigrationMarkerLocked(worktree, "t961-install")
	if err == nil {
		_ = release(true)
		t.Fatal("installing a migration marker from a linked worktree must be refused")
	}
	assertRefusalNamesBothTrees(t, err.Error(), primary, worktree)
	assertNoMarkerWritten(t, primary)
}

// TestT961_AcquireMigrationAdmissionRefusesNonCanonicalCaller covers the second
// entry point of the same family. Guarding only one of the two would leave the
// other as an unguarded spelling of the identical mutation.
func TestT961_AcquireMigrationAdmissionRefusesNonCanonicalCaller(t *testing.T) {
	primary, worktree := twoTreeFixture(t)
	isolateHome(t)

	release, err := AcquireMigrationAdmission(worktree, "t961-admission")
	if err == nil {
		_ = release(true)
		t.Fatal("acquiring migration admission from a linked worktree must be refused")
	}
	assertRefusalNamesBothTrees(t, err.Error(), primary, worktree)
	assertNoMarkerWritten(t, primary)
}

// TestT961_CanonicalCallerStillInstallsMarker is the counter-case that keeps the
// gate from being a blunt ban: the operator who runs from the canonical root is
// exactly the caller the refusal message tells people to become, so that call
// must still succeed and still write the marker. It doubles as the positive
// control — a silent no-op instrument would fail here, not pass.
func TestT961_CanonicalCallerStillInstallsMarker(t *testing.T) {
	primary, _ := twoTreeFixture(t)
	isolateHome(t)

	release, err := InstallMigrationMarkerLocked(primary, "t961-canonical")
	if err != nil {
		t.Fatalf("a caller at the canonical root must still install the marker: %v", err)
	}
	path, err := MigrationBarrierPath(primary)
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("marker must exist after a permitted install: %v", statErr)
	}
	if err := release(true); err != nil {
		t.Fatalf("release: %v", err)
	}
}

// TestT961_GateIsVacuousForCallersPassingTheCanonicalRoot pins the boundary that
// makes this a preemptive gate rather than a behaviour change: the CLI apply
// path resolves the root itself before it reaches this package, so the gate can
// never fire for it. If a later edit makes the CLI pass its own tree instead,
// this test is where that shows up.
func TestT961_GateIsVacuousForCallersPassingTheCanonicalRoot(t *testing.T) {
	_, worktree := twoTreeFixture(t)
	isolateHome(t)

	if err := RefuseMutationFromNonCanonicalTree(CanonicalProjectRoot(worktree)); err != nil {
		t.Fatalf("a caller that already resolved the canonical root must pass: %v", err)
	}
}

// TestT961_GateIsInertOutsideARepository keeps ordinary temp-directory callers —
// every other test in this package — out of the gate's way.
func TestT961_GateIsInertOutsideARepository(t *testing.T) {
	isolateHome(t)
	if err := RefuseMutationFromNonCanonicalTree(t.TempDir()); err != nil {
		t.Fatalf("a non-repository root must pass: %v", err)
	}
}

func assertRefusalNamesBothTrees(t *testing.T, msg, primary, worktree string) {
	t.Helper()
	resolvedPrimary, err := filepath.EvalSymlinks(primary)
	if err != nil {
		t.Fatal(err)
	}
	resolvedWorktree, err := filepath.EvalSymlinks(worktree)
	if err != nil {
		t.Fatal(err)
	}
	// The operator has to be able to act on the message, so it names the tree
	// they are in, the tree the mutation would reach, and why their isolation
	// did not hold.
	for _, want := range []string{resolvedWorktree, resolvedPrimary, "MOAI_HOME"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal must name %q; got %q", want, msg)
		}
	}
}

func assertNoMarkerWritten(t *testing.T, canonicalRoot string) {
	t.Helper()
	path, err := MigrationBarrierPath(canonicalRoot)
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("a refused call must write no marker at %q (stat err=%v)", path, statErr)
	}
}
