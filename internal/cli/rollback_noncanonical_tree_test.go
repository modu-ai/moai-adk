package cli

// rollback_noncanonical_tree_test.go — t1013.
//
// t961 landed the package-layer refusal and recorded one Gap: it claimed, from
// source reading alone, that `migrate home-state rollback` is now refused when
// issued from a linked worktree, because the cobra RunE hands rollbackHomeState
// the raw cwd instead of a canonicalized root. It could not observe that claim,
// for a stated reason — rollback reaches verifiedBackup BEFORE it reaches the
// refusal, so a probe without a genuinely verified backup fails for the wrong
// reason and observes nothing.
//
// This file removes that reason. It builds a backup that verifiedBackup
// actually accepts, asserts that acceptance as a premise, and only then fires
// the rollback. The premise is what turns a RED into an observation: without
// it, a refusal could be verifiedBackup's and would be read as the guard's.

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t1013Fixture builds a real repository plus a real linked worktree, seeds the
// primary's source queue, and isolates MOAI_HOME — the exact configuration of a
// caller who believes a worktree plus a private home is fully isolated.
func t1013Fixture(t *testing.T) (primary, worktree string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	base := t.TempDir()
	primary = filepath.Join(base, "primary")
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

	worktree = filepath.Join(base, "linked")
	runGit(primary, "worktree", "add", "-q", "-b", "t1013-linked", worktree)

	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "isolated-home"))

	source := filepath.Join(primary, ".moai", "state", "todo", "backlog.db")
	if err := os.MkdirAll(filepath.Dir(source), 0o700); err != nil {
		t.Fatal(err)
	}
	execSQLiteForT1013(t, source, `CREATE TABLE items(id TEXT PRIMARY KEY, text TEXT NOT NULL); INSERT INTO items VALUES('t1','one'),('t2','two')`)
	return primary, worktree
}

func execSQLiteForT1013(t *testing.T, path, stmt string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(stmt)
	if closeErr := db.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
}

// t1013ArmedRollback produces the state a rollback actually needs and returns
// the target path plus its pre-rollback bytes:
//
//	(1) a verified backup, made by an apply from the canonical root, and
//	(2) a target that DIVERGES from that backup.
//
// (2) is load-bearing and easy to omit. rollbackHomeState returns nil early
// when the target already equals the backup, so a fixture that skips the
// divergence never reaches the refusal at all — and that nil reads exactly like
// "not refused".
func t1013ArmedRollback(t *testing.T, primary string) (target string, before []byte) {
	t.Helper()
	if _, err := runHomeState(t, primary, true); err != nil {
		t.Fatalf("apply from the canonical root must succeed to produce a backup: %v", err)
	}
	target, err := homestate.BacklogDBPath(primary)
	if err != nil {
		t.Fatal(err)
	}
	execSQLiteForT1013(t, target, `INSERT INTO items VALUES('t3','three')`)
	before, err = os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	return target, before
}

// TestT1013_PremiseVerifiedBackupAcceptsTheWorktreeCaller is the premise this
// whole file rests on, and the one t961 could not establish. If verifiedBackup
// rejected the worktree caller, every refusal below would be ITS refusal, and
// reading it as the non-canonical-tree guard's would be a false observation.
func TestT1013_PremiseVerifiedBackupAcceptsTheWorktreeCaller(t *testing.T) {
	primary, worktree := t1013Fixture(t)
	t1013ArmedRollback(t, primary)

	// The fixture must actually discriminate, or nothing below means anything.
	resolvedPrimary, err := filepath.EvalSymlinks(primary)
	if err != nil {
		t.Fatal(err)
	}
	resolvedWorktree, err := filepath.EvalSymlinks(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if canonical := homestate.CanonicalProjectRoot(worktree); canonical != resolvedPrimary {
		t.Fatalf("fixture premise: canonical root %q should be the primary %q", canonical, resolvedPrimary)
	}
	if resolvedPrimary == resolvedWorktree {
		t.Fatalf("fixture premise: the two trees must be different paths (%q)", resolvedPrimary)
	}

	id := t1013LatestBackupID(t, primary)
	if _, _, err := verifiedBackup(worktree, id); err != nil {
		t.Fatalf("premise failed: verifiedBackup must ACCEPT the worktree caller, otherwise a rollback refusal cannot be attributed to the tree guard: %v", err)
	}
	// Symmetric statement of the same premise from the canonical side, so a
	// future change that breaks only one of the two is visible.
	if _, _, err := verifiedBackup(primary, id); err != nil {
		t.Fatalf("premise failed: verifiedBackup must accept the canonical caller: %v", err)
	}
}

// TestT1013_PremiseTargetDivergesFromBackup pins the second premise: the early
// no-op return is NOT the branch under test.
func TestT1013_PremiseTargetDivergesFromBackup(t *testing.T) {
	primary, _ := t1013Fixture(t)
	target, _ := t1013ArmedRollback(t, primary)

	id := t1013LatestBackupID(t, primary)
	backup, _, err := verifiedBackup(primary, id)
	if err != nil {
		t.Fatal(err)
	}
	targetCensus, err := sqliteCensus(target)
	if err != nil {
		t.Fatal(err)
	}
	backupCensus, err := sqliteCensus(backup)
	if err != nil {
		t.Fatal(err)
	}
	if targetCensus == backupCensus {
		t.Fatal("fixture premise: target must diverge from the backup, or rollback returns nil before the guard is reached")
	}
}

// TestT1013_RollbackFromWorktreeIsRefusedBeforeQuarantine is the observation
// t961 deferred. The refusal must be the tree guard's — named by its own words —
// and it must land before anything is quarantined or restored.
func TestT1013_RollbackFromWorktreeIsRefusedBeforeQuarantine(t *testing.T) {
	primary, worktree := t1013Fixture(t)
	target, before := t1013ArmedRollback(t, primary)
	id := t1013LatestBackupID(t, primary)

	err := rollbackHomeState(context.Background(), worktree, id)
	if err == nil {
		t.Fatal("a rollback issued from a linked worktree must be refused")
	}

	// Attribute the failure verbatim. A refusal for ANY other reason — a
	// missing, mismatched or unreadable backup above all — is not an
	// observation of this guard, so it fails here rather than being counted.
	msg := err.Error()
	// Logged, not merely matched: the verbatim reason is the evidence, and a
	// later reader must be able to see WHICH refusal this was without rerunning
	// with a debugger attached.
	t.Logf("verbatim refusal: %s", msg)
	resolvedPrimary, evalErr := filepath.EvalSymlinks(primary)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	resolvedWorktree, evalErr := filepath.EvalSymlinks(worktree)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	for _, want := range []string{"refused", resolvedWorktree, resolvedPrimary, "MOAI_HOME"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal must be the non-canonical-tree guard's and must name %q; got %q", want, msg)
		}
	}
	if strings.Contains(msg, "backup") {
		t.Errorf("wrong-reason RED: this failed in backup verification, not in the tree guard: %q", msg)
	}

	// Nothing may have moved. The guard sits ahead of both the quarantine
	// rename and the restore copy, so both the target's bytes and the absence
	// of scratch siblings are part of the claim.
	after, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("refused rollback must not change the target")
	}
	for _, pattern := range []string{target + ".rollback-*", target + ".restore", target + ".quarantine-*"} {
		if hits, globErr := filepath.Glob(pattern); globErr == nil && len(hits) != 0 {
			t.Errorf("refused rollback left %v behind; the refusal must precede quarantine and restore", hits)
		}
	}
}

// TestT1013_RollbackFromCanonicalRootStillProceeds is the positive control. A
// guard that refused both callers would satisfy the test above while being a
// blunt ban, and the instrument would be indistinguishable from one that always
// errors. The canonical caller is the one the refusal message tells the
// operator to become, so it must get through and restore parity.
func TestT1013_RollbackFromCanonicalRootStillProceeds(t *testing.T) {
	primary, _ := t1013Fixture(t)
	target, before := t1013ArmedRollback(t, primary)
	id := t1013LatestBackupID(t, primary)

	if err := rollbackHomeState(context.Background(), primary, id); err != nil {
		t.Fatalf("a rollback issued from the canonical root must NOT be refused: %v", err)
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) {
		t.Fatal("positive control did no work: the diverged target should have been restored")
	}
	backup, manifest, err := verifiedBackup(primary, id)
	if err != nil {
		t.Fatal(err)
	}
	final, err := sqliteCensus(target)
	if err != nil {
		t.Fatal(err)
	}
	backupCensus, err := sqliteCensus(backup)
	if err != nil {
		t.Fatal(err)
	}
	if final != backupCensus || final.Digest != manifestDigest(manifest) {
		t.Fatalf("restored target must match the backup: target=%+v backup=%+v", final, backupCensus)
	}
}

func t1013LatestBackupID(t *testing.T, primary string) string {
	t.Helper()
	base := filepath.Join(os.Getenv("MOAI_HOME"), "backups", homestate.ProjectKey(primary))
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("fixture expects exactly one backup; got %d", len(entries))
	}
	return entries[0].Name()
}
