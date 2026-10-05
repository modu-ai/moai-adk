//go:build integration
// +build integration

// Package harness_integration — SPEC-FACTORY-RUN-RETIRE-001 (card t1107).
//
// IT-08 exercises the owner-liveness predicate and the reconciler on the one
// path the three-OS `test-integration` job runs, so darwin, linux, and windows
// each execute the REAL platform probe rather than a stub.
//
// The identities below are constructed so the verdict is deterministic on
// every platform without spawning or racing anything:
//
//   - live — this process's own pid with the fingerprint the probe reports;
//   - dead — this process's own pid with a DIFFERENT fingerprint, which is the
//     pid-reuse shape: the pid is live, the process is not the recorded one;
//   - indeterminate — no identity from either source.
package harness_integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryRetireSandbox returns a project root OUTSIDE this repository's
// worktree set with HOME / MOAI_HOME / MOAI_CLAUDE_BIN scrubbed to it
// (REQ-012). CanonicalProjectRoot converges a linked worktree onto the primary
// checkout, so "use a temp dir" alone would mutate real factory state.
func factoryRetireSandbox(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
	t.Setenv("MOAI_CLAUDE_BIN", filepath.Join(home, "bin", "claude"))
	root := filepath.Join(t.TempDir(), "sandbox-project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatalf("create sandbox project root: %v", err)
	}
	return root
}

func TestFactoryRunRetire(t *testing.T) {
	liveFingerprint := homestate.CurrentProcessFingerprint()
	if liveFingerprint == "" {
		t.Fatalf("this host cannot report its own process-start fingerprint; the liveness predicate is unexercisable here")
	}
	const deadFingerprint = "not-the-fingerprint-this-pid-was-started-with"
	pid := os.Getpid()

	// The predicate itself, through the real platform probe.
	t.Run("predicate", func(t *testing.T) {
		if got := homestate.DefaultOwnerClassifier(pid, liveFingerprint); got != homestate.OwnerLive {
			t.Fatalf("matching fingerprint = %q, want live", got)
		}
		if got := homestate.DefaultOwnerClassifier(pid, deadFingerprint); got != homestate.OwnerDead {
			t.Fatalf("pid-reuse shape = %q, want dead (a predicate consulting only the pid fails here)", got)
		}
		if got := homestate.DefaultOwnerClassifier(0, ""); got != homestate.OwnerIndeterminate {
			t.Fatalf("absent identity = %q, want indeterminate", got)
		}
	})

	// The reconciler: only the dead owner's run leaves 'active'.
	t.Run("reconciler", func(t *testing.T) {
		db := openSandboxFactory(t)
		record(t, db, "run-live", pid, liveFingerprint)
		record(t, db, "run-dead", pid, deadFingerprint)
		record(t, db, "run-unknown", 0, "")

		rec, err := db.ReconcileActiveRuns(context.Background(), homestate.ReconcileOptions{})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(rec.Retired) != 1 || rec.Retired[0].RunID != "run-dead" {
			t.Fatalf("retired = %+v, want exactly run-dead", rec.Retired)
		}
		if got := status(t, db, "run-dead"); got != "retired" {
			t.Fatalf("run-dead status = %q, want \"retired\"", got)
		}
		for _, id := range []string{"run-live", "run-unknown"} {
			if got := status(t, db, id); got != "active" {
				t.Fatalf("%s status = %q, want \"active\" — a live or unprobeable owner is never retired", id, got)
			}
		}
		// The row survives retirement and the transition is observable.
		var events int
		if err := db.DB.QueryRow(`SELECT count(*) FROM events WHERE run_id='run-dead' AND kind='run.retired'`).Scan(&events); err != nil {
			t.Fatalf("count run.retired events: %v", err)
		}
		if events != 1 {
			t.Fatalf("run.retired event count = %d, want 1", events)
		}
	})

	// Retirement is gated on a POSITIVE dead: a classification the code does
	// not enumerate declines by default rather than falling through to retire.
	t.Run("unenumerated classification declines", func(t *testing.T) {
		db := openSandboxFactory(t)
		record(t, db, "run-4th", pid, liveFingerprint)
		fourth := homestate.OwnerClassification("unknown-host")
		rec, err := db.ReconcileActiveRuns(context.Background(), homestate.ReconcileOptions{
			Classify: func(int, string) homestate.OwnerClassification { return fourth },
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(rec.Retired) != 0 {
			t.Fatalf("retired %+v on an unenumerated classification", rec.Retired)
		}
		if got := status(t, db, "run-4th"); got != "active" {
			t.Fatalf("status = %q, want \"active\"", got)
		}
	})
}

func openSandboxFactory(t *testing.T) *homestate.FactoryDB {
	t.Helper()
	db, err := homestate.OpenFactory(factoryRetireSandbox(t))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func record(t *testing.T, db *homestate.FactoryDB, runID string, pid int, processStart string) {
	t.Helper()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: "claude", ManifestJSON: "{}",
		LeadPID: pid, LeadProcessStart: processStart,
	}); err != nil {
		t.Fatalf("record run %s: %v", runID, err)
	}
}

func status(t *testing.T, db *homestate.FactoryDB, runID string) string {
	t.Helper()
	var s string
	if err := db.DB.QueryRow(`SELECT status FROM runs WHERE run_id=?`, runID).Scan(&s); err != nil {
		t.Fatalf("read status %s: %v", runID, err)
	}
	return s
}
