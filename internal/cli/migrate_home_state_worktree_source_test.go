package cli

// migrate_home_state_worktree_source_test.go — t952.
//
// The defect this file pins: `MOAI_HOME` isolates the migration's TARGET and
// nothing else. The SOURCE is assembled from CanonicalProjectRoot, which
// deliberately resolves a linked worktree back to the primary checkout so that
// one repository keeps one queue. A caller who isolates the home and runs from
// a worktree therefore believes both ends are isolated while apply would move
// the PRIMARY checkout's live state.
//
// The fixture below is the discriminant the repair is judged on: it builds a
// real repository plus a real linked worktree, so the caller's own tree and the
// canonical root are necessarily different paths. Nothing here runs a live
// apply — the refusal is asserted before any mutation.

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"

	_ "modernc.org/sqlite"
)

// worktreeSourceFixture builds a git repository with one seeded source db and a
// linked worktree, returning (primary, worktree). It skips rather than fails
// when git is unavailable, so the suite stays runnable on a bare machine.
func worktreeSourceFixture(t *testing.T) (string, string) {
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

	// Seed the legacy source only in the PRIMARY tree — the worktree has none.
	// That asymmetry is what makes a wrong resolution visible.
	source := filepath.Join(primary, ".moai", "state", "todo", "backlog.db")
	if err := os.MkdirAll(filepath.Dir(source), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", source)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE items(id TEXT PRIMARY KEY, text TEXT NOT NULL); INSERT INTO items VALUES('p1','primary-only')`)
	if closeErr := db.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}

	worktree := filepath.Join(base, "linked")
	runGit(primary, "worktree", "add", "-q", "-b", "linked-branch", worktree)
	return primary, worktree
}

// TestT952_WorktreeCallerAndCanonicalRootDiverge is the premise assertion. Every
// other test here is meaningless if the fixture does not actually place the
// caller's tree and the canonical root at different paths, so this states it
// first and fails loudly when the fixture stops discriminating.
func TestT952_WorktreeCallerAndCanonicalRootDiverge(t *testing.T) {
	primary, worktree := worktreeSourceFixture(t)

	canonical := homestate.CanonicalProjectRoot(worktree)
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
	// And the source really is primary-only, so a wrong resolution is a real move.
	if _, err := os.Stat(filepath.Join(resolvedWorktree, ".moai", "state", "todo", "backlog.db")); !os.IsNotExist(err) {
		t.Fatalf("fixture premise: the worktree must carry no source of its own (stat err=%v)", err)
	}
}

// TestT952_ApplyRefusesWhenCallerTreeIsNotCanonicalRoot is the repair's
// acceptance criterion. An apply issued from a linked worktree operates on the
// primary's live state no matter what MOAI_HOME says; it must refuse before any
// mutation and name both paths, so the operator sees why their isolation did
// not hold.
func TestT952_ApplyRefusesWhenCallerTreeIsNotCanonicalRoot(t *testing.T) {
	primary, worktree := worktreeSourceFixture(t)
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "isolated-home"))

	var out bytes.Buffer
	r := homeStateRunner{projectRoot: worktree, apply: true, verifiedLive: true, stdout: &out}
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("apply from a linked worktree must be refused")
	}
	msg := err.Error()
	for _, want := range []string{"worktree", "canonical"} {
		if !strings.Contains(strings.ToLower(msg), want) {
			t.Errorf("refusal must name %q; got %q", want, msg)
		}
	}

	// Refused BEFORE mutation: the primary's source is still there, untouched.
	if _, statErr := os.Stat(filepath.Join(primary, ".moai", "state", "todo", "backlog.db")); statErr != nil {
		t.Fatalf("primary source must survive a refused apply: %v", statErr)
	}
}

// TestT952_DryRunFromWorktreeStillReports is the counter-case that keeps the
// repair from being a blunt ban. Dry-run is how an operator DISCOVERS this
// situation, so it must keep working from a worktree and keep naming the
// canonical root it resolved.
func TestT952_DryRunFromWorktreeStillReports(t *testing.T) {
	primary, worktree := worktreeSourceFixture(t)
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "isolated-home"))

	var out bytes.Buffer
	r := homeStateRunner{projectRoot: worktree, apply: false, stdout: &out}
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("dry-run from a worktree must still report: %v", err)
	}
	resolvedPrimary, err := filepath.EvalSymlinks(primary)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), resolvedPrimary) {
		t.Errorf("dry-run must name the canonical root it resolved; output=%q", out.String())
	}
}
